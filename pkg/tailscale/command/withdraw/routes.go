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
	"fmt"

	"inet.af/netaddr"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud/remix"
	"github.com/dhermes/tailsk8s/pkg/tailscale/command/advertise"
)

// DisableWithdrawnCIDR ensures that a recently withdrawn CIDR is removed from
// the set of enabled routes in the Tailscale cloud API.
func DisableWithdrawnCIDR(ctx context.Context, c cloud.Config, cidr netaddr.IPPrefix, hostname string) error {
	// Retrieve the Tailscale node ID corresponding to the local hostname
	gdbhr := remix.GetDeviceByHostnameRequest{Hostname: hostname}
	device, err := remix.GetDeviceByHostname(ctx, c, gdbhr)
	if err != nil {
		return err
	}

	// Find all the **current** routes for the device.
	grr := cloud.GetRoutesRequest{DeviceID: device.ID}
	rr, err := cloud.GetRoutes(ctx, c, grr)
	if err != nil {
		return err
	}
	if len(rr.AdvertisedRoutes) > 0 {
		cli.Printf(ctx, "Advertised routes for device %s:\n", device.ID)
		for _, ar := range rr.AdvertisedRoutes {
			cli.Printf(ctx, "- %s\n", ar)
		}
	}
	if len(rr.EnabledRoutes) > 0 {
		cli.Printf(ctx, "Enabled routes for device %s:\n", device.ID)
		for _, ar := range rr.EnabledRoutes {
			cli.Printf(ctx, "- %s\n", ar)
		}
	}

	// Ensure `cidr` is **not** contained in `routes.AdvertisedRoutes`. If it **is**
	// it could be the fault of the caller (i.e. the CIDR was never withdrawn)
	// or it could be the result of a race condition (i.e. this request is made
	// **before** the newly withdrawn CIDR is acknowledged by the Tailscale
	// control plane).
	// TODO: This race condition has been encountered in the wild, implement
	//       retry / backoff / sleep.
	if advertise.RoutesContain(rr.AdvertisedRoutes, cidr) {
		return fmt.Errorf("withdrawn route (%s) is still among list of advertised routes", cidr)
	}

	// If `cidr` is not contained in `routes.EnabledRoutes`, there is nothing to do.
	if !advertise.RoutesContain(rr.EnabledRoutes, cidr) {
		cli.Printf(ctx, "Device %s has already disabled route %s\n", device.ID, cidr)
		return nil
	}

	// ...otherwise, remove it and call `SetRoutes()`
	routes := routesRemove(rr.EnabledRoutes, cidr)
	srr := cloud.SetRoutesRequest{DeviceID: device.ID, Routes: routes}
	cli.Printf(ctx, "Disabling route %s for device %s...\n", cidr, device.ID)
	_, err = cloud.SetRoutes(ctx, c, srr)
	if err != nil {
		return err
	}

	cli.Printf(ctx, "Disabled route %s for device %s\n", cidr, device.ID)
	return nil
}

func routesRemove(routes []string, cidr netaddr.IPPrefix) []string {
	keep := make([]string, 0, len(routes))
	cidrString := cidr.String()
	for _, r := range routes {
		if r != cidrString {
			keep = append(keep, r)
		}
	}
	return keep
}
