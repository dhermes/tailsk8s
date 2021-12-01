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

package remix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud"
)

// GetDeviceByHostname looks up a device either directly by the literal
// hostname or by the machine name in the Tailnet. Concretely, this lists all
// devices in the Tailnet, then matches directly against `hostname` **OR**
// matches that `name` is equal to `{hostname}.{ac.Tailnet}`.
func GetDeviceByHostname(ctx context.Context, c cloud.Config, req GetDeviceByHostnameRequest) (*cloud.Device, error) {
	devices, err := cloud.GetDevices(ctx, c, cloud.Empty{})
	if err != nil {
		return nil, err
	}

	deviceName := fmt.Sprintf("%s.%s", req.Hostname, c.Tailnet)
	matches := []cloud.Device{}
	for _, device := range devices.Devices {
		if device.Hostname == req.Hostname || device.Name == deviceName {
			matches = append(matches, device)
		}
	}

	if len(matches) != 1 {
		return nil, fmt.Errorf("could not find unique device matching hostname %q (%d matches)", req.Hostname, len(matches))
	}

	device := matches[0]
	if cli.GetDebug(ctx) {
		// Ignore error: failure to marshal in debug mode can't break the regular flow.
		asJSON, _ := json.MarshalIndent(device, "> ", "    ")
		cli.DebugPrintln(ctx, "Matched device:")
		cli.DebugPrintln(ctx, "> "+string(asJSON))
	}
	return &device, nil
}
