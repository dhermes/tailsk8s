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
	"time"

	"inet.af/netaddr"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

// WithdrawAndDisable first uses the local `tailscaled` API to withdraw a
// CIDR from the Tailnet and then uses the cloud API to disable the withdrawn
// CIDR.
func WithdrawAndDisable(ctx context.Context, c Config) error {
	err := c.APIConfig.Resolve(ctx)
	if err != nil {
		return err
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

	// TODO: Remove this sleep once a more sane way of handling the race
	//       condition between the local change above and the remote change below.
	time.Sleep(5 * time.Second)

	return DisableWithdrawnCIDR(ctx, c.APIConfig, cidr, hostname)
}
