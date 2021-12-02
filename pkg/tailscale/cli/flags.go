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

package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"tailscale.com/client/tailscale"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

const (
	// debugCurlStatusWithoutPeers is a debug mode representation of the
	// "status without peers" curl command.
	debugCurlStatusWithoutPeers = `Calling ""status without peers" local API route:
> curl \
>   --include \
>   --unix-socket /var/run/tailscale/tailscaled.sock \
>   http://no-op-host.invalid/localapi/v0/status?peers=false
`
)

// ReadAPIKey reads a Tailscale API key from file if the provided `apiKey`
// is prefixed with `file:`. It is expected that this value has been passed
// via a CLI flag such as `--api-key`.
func ReadAPIKey(ctx context.Context, apiKey string) (string, error) {
	if !strings.HasPrefix(apiKey, "file:") {
		return apiKey, nil
	}

	filename := strings.TrimPrefix(apiKey, "file:")
	cli.Printf(ctx, "Reading Tailscale API key from: %s\n", filename)
	apiKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(apiKeyBytes), nil
}

// DefaultTailnet attempts to determine the locally active Tailnet
// via the local `tailscaled` API. If `tailnet` is already set, it will be
// used without checking the default. It is expected that this value has been
// passed via a CLI flag such as `--tailnet`.
func DefaultTailnet(ctx context.Context, tailnet string) (string, error) {
	if tailnet != "" {
		return tailnet, nil
	}

	cli.DebugPrintf(ctx, debugCurlStatusWithoutPeers)
	status, err := tailscale.StatusWithoutPeers(ctx)
	if err != nil {
		return "", nil
	}

	cli.Printf(ctx, "Inferring Tailnet from magic DNS suffix: %s\n", status.MagicDNSSuffix)
	return getTailnet(status.MagicDNSSuffix)
}

// getTailnet parses a magic DNS suffix to determine the Tailnet name. The
// assumption is that the magic DNS suffix is of the form
// `{TAILNET}.beta.tailscale.net`.
func getTailnet(magicDNSSuffix string) (string, error) {
	if !strings.HasSuffix(magicDNSSuffix, ".beta.tailscale.net") {
		return "", fmt.Errorf("bad %s", magicDNSSuffix)
	}

	return strings.TrimSuffix(magicDNSSuffix, ".beta.tailscale.net"), nil
}
