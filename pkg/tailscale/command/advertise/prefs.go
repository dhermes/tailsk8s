// Copyright 2021 Danny Hermes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package advertise

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"inet.af/netaddr"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/types/key"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

const (
	debugCurlGetPrefs = `Calling "get prefs" local API route:
> curl \
>   --include \
>   --unix-socket /var/run/tailscale/tailscaled.sock \
>   http://no-op-host.invalid/localapi/v0/prefs
`
	// debugCurlEditPrefs is a template to print (in debug mode) the
	// equivalent curl command to the outgoing request. The PATCH body is
	// not expected to be `shlex` quoted by the template user, but it should be.
	debugCurlEditPrefs = `Calling "edit prefs" local API route:
> curl \
>   --include \
>   --request PATCH \
>   --data-binary '%s' \
>   --unix-socket /var/run/tailscale/tailscaled.sock \
>   http://no-op-host.invalid/localapi/v0/prefs
`
)

// EditPrefsAdvertiseCIDR updates existing Tailscale preferences to
// - Accept routes advertised by other Tailscale nodes (equivalent to the
//   `--accept-routes` flag for `tailscale up`)
// - Advertise a new route to other Tailscale nodes in addition to any existing
//   routes (equivalent to the `--advertise-routes` flag for `tailscale up`,
//   but also ensures existing routes are not clobbered)
//
// If the accept routes flag and the advertised CIDR are both present, this
// will make no changes.
func EditPrefsAdvertiseCIDR(ctx context.Context, cidr netaddr.IPPrefix) error {
	cli.DebugPrintf(ctx, debugCurlGetPrefs)
	before, err := tailscale.GetPrefs(ctx)
	if err != nil {
		return err
	}

	hasCIDR := ipPrefixesContain(before.AdvertiseRoutes, cidr)
	if hasCIDR && before.RouteAll {
		cli.Println(ctx, "Routes already accepted and advertised")
		return nil
	}

	patch := &ipn.MaskedPrefs{}
	patch.Prefs = *before.Clone()
	if !before.RouteAll {
		patch.Prefs.RouteAll = true
		patch.RouteAllSet = true
	}
	if !hasCIDR {
		patch.Prefs.AdvertiseRoutes = append(patch.Prefs.AdvertiseRoutes, cidr)
		patch.AdvertiseRoutesSet = true
	}

	if cli.GetDebug(ctx) {
		asJSON, err := json.Marshal(patch)
		// Ignore error: failure to marshal in debug mode can't break the regular flow.
		if err != nil {
			asJSON = []byte("...")
		}
		cli.DebugPrintf(ctx, debugCurlEditPrefs, string(asJSON))
	}

	after, err := tailscale.EditPrefs(ctx, patch)
	if err != nil {
		return err
	}

	diffBeforeAfter(ctx, before, after)
	return nil
}

func ipPrefixesContain(prefixes []netaddr.IPPrefix, cidr netaddr.IPPrefix) bool {
	for _, prefix := range prefixes {
		if prefix.String() == cidr.String() {
			return true
		}
	}
	return false
}

// diffBeforeAfter writes `before` and `after` to JSON and execs out to
// `diff` to compare them. If **any** of these steps errors, it will just exit
// since this is meant to **aid** understanding during debug mode (vs. to
// provide actual functionality).
func diffBeforeAfter(ctx context.Context, before, after *ipn.Prefs) {
	if !cli.GetDebug(ctx) {
		return
	}

	// Clean sensitive fields before writing
	before.Persist.LegacyFrontendPrivateMachineKey = key.MachinePrivate{}
	before.Persist.PrivateNodeKey = key.NodePrivate{}
	before.Persist.OldPrivateNodeKey = key.NodePrivate{}
	after.Persist.LegacyFrontendPrivateMachineKey = key.MachinePrivate{}
	after.Persist.PrivateNodeKey = key.NodePrivate{}
	after.Persist.OldPrivateNodeKey = key.NodePrivate{}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return
	}

	// Write `before` to a temp file
	filenameBefore := filepath.Join(dir, "before.json")
	asJSON, err := json.MarshalIndent(before, "", "    ")
	if err != nil {
		return
	}
	err = os.WriteFile(filenameBefore, asJSON, 0644)
	if err != nil {
		return
	}

	// Write `after` to a temp file
	filenameAfter := filepath.Join(dir, "after.json")
	asJSON, err = json.MarshalIndent(after, "", "    ")
	if err != nil {
		return
	}
	err = os.WriteFile(filenameAfter, asJSON, 0644)
	if err != nil {
		return
	}

	// Exec out to `diff` and capture STDOUT
	cmd := exec.CommandContext(
		ctx, "diff", "--report-identical-files", "--unified",
		filenameBefore, filenameAfter,
	)
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		// If `diff` exits with a status code of 1 or 2, that's OK, it just
		// means it found some differences.
		_, ok := err.(*exec.ExitError)
		if !ok {
			return
		}
	}

	cli.DebugPrintf(ctx, b.String())
}
