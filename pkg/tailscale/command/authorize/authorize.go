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

package authorize

import (
	"context"
	"os"
	"strings"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud"
)

// AuthorizeDevice retrieves a device by name / hostname and then uses the
// device ID to authorize the device.
func AuthorizeDevice(ctx context.Context, c Config) error {
	if strings.HasPrefix(c.APIConfig.APIKey, "file:") {
		filename := strings.TrimPrefix(c.APIConfig.APIKey, "file:")
		cli.Printf(ctx, "Reading Tailscale API key from: %s\n", filename)
		apiKeyBytes, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		c.APIConfig.APIKey = string(apiKeyBytes)
	}

	hostname := c.Hostname
	var err error
	if hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			return err
		}
	}
	cli.Printf(ctx, "Using hostname: %s\n", hostname)

	device, err := GetDeviceByHostname(ctx, c.APIConfig, hostname)
	if err != nil {
		return err
	}

	if device.Authorized {
		cli.Printf(ctx, "Device %s is already authorized\n", device.ID)
		return nil
	}

	cli.Printf(ctx, "Authorizing device ID: %s...\n", device.ID)
	adr := cloud.AuthorizeDeviceRequest{ID: device.ID, Authorized: true}
	_, err = cloud.AuthorizeDevice(ctx, c.APIConfig, adr)
	if err != nil {
		return err
	}

	cli.Printf(ctx, "Authorized device ID: %s\n", device.ID)
	return nil
}
