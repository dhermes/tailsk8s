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
	"os"
	"strings"

	"inet.af/netaddr"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

// WithdrawAndDisable first uses the local `tailscaled` API to withdraw a
// CIDR from the Tailnet and then uses the cloud API to disable the withdrawn
// CIDR.
func WithdrawAndDisable(ctx context.Context, c Config) error {
	if strings.HasPrefix(c.APIConfig.APIKey, "file:") {
		filename := strings.TrimPrefix(c.APIConfig.APIKey, "file:")
		cli.Printf(ctx, "Reading Tailscale API key from: %s\n", filename)
		apiKeyBytes, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		c.APIConfig.APIKey = string(apiKeyBytes)
	}

	cidr, err := netaddr.ParseIPPrefix(c.IPv4CIDR)
	if err != nil {
		return err
	}

	err = EditPrefsWithdrawCIDR(ctx, cidr)
	if err != nil {
		return err
	}

	// Use the local hostname to determine the Tailscale node ID
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	cli.Printf(ctx, "Using hostname: %s\n", hostname)

	return DisableWithdrawnCIDR(ctx, c.APIConfig, cidr, hostname)
}
