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

package withdraw

import (
	"context"
	"encoding/json"

	"inet.af/netaddr"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/command/advertise"
)

// EditPrefsWithdrawCIDR updates existing Tailscale preferences to withdraw
// a routes advertised by the current Tailscale node.
//
// If the accept routes flag and the advertised CIDR are both present, this
// will make no changes.
func EditPrefsWithdrawCIDR(ctx context.Context, cidr netaddr.IPPrefix) error {
	cli.DebugPrintf(ctx, advertise.DebugCurlGetPrefs)
	before, err := tailscale.GetPrefs(ctx)
	if err != nil {
		return err
	}

	hasCIDR := advertise.IPPrefixesContain(before.AdvertiseRoutes, cidr)
	if !hasCIDR {
		cli.Println(ctx, "Route already withdrawn")
		return nil
	}

	patch := &ipn.MaskedPrefs{}
	patch.Prefs = *before.Clone()
	patch.Prefs.AdvertiseRoutes = ipPrefixesRemove(patch.Prefs.AdvertiseRoutes, cidr)
	patch.AdvertiseRoutesSet = true

	if cli.GetDebug(ctx) {
		asJSON, err := json.Marshal(patch)
		// Ignore error: failure to marshal in debug mode can't break the regular flow.
		if err != nil {
			asJSON = []byte("...")
		}
		cli.DebugPrintf(ctx, advertise.DebugCurlEditPrefs, string(asJSON))
	}

	after, err := tailscale.EditPrefs(ctx, patch)
	if err != nil {
		return err
	}

	advertise.DiffBeforeAfter(ctx, before, after)
	return nil
}

func ipPrefixesRemove(prefixes []netaddr.IPPrefix, cidr netaddr.IPPrefix) []netaddr.IPPrefix {
	keep := make([]netaddr.IPPrefix, 0, len(prefixes))
	cidrString := cidr.String()
	for _, prefix := range prefixes {
		if prefix.String() != cidrString {
			keep = append(keep, prefix)
		}
	}
	return keep
}
