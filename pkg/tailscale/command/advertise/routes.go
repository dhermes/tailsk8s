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
	"context"
	"fmt"
	"time"

	"inet.af/netaddr"

	"github.com/dhermes/tailsk8s/pkg/cli"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud"
	"github.com/dhermes/tailsk8s/pkg/tailscale/cloud/remix"
)

// AcceptNewCIDR ensures that a newly advertised CIDR is an enabled subnet
// in the Tailscale cloud API.
func AcceptNewCIDR(ctx context.Context, c cloud.Config, cidr netaddr.IPPrefix, hostname string) error {
	// Retrieve the Tailscale node ID corresponding to the local hostname
	gdbhr := remix.GetDeviceByHostnameRequest{Hostname: hostname}
	device, err := remix.GetDeviceByHostname(ctx, c, gdbhr)
	if err != nil {
		return err
	}

	// TODO: Remove this sleep once a more sane way of handling the race
	//       condition below is implemented (see `!RoutesContain()` check).
	time.Sleep(5 * time.Second)

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

	// Ensure `cidr` is contained in `routes.AdvertisedRoutes`. If it **isn't**
	// it could be the fault of the caller (i.e. the CIDR was never advertised)
	// or it could be the result of a race condition (i.e. this request is made
	// **before** the newly advertised CIDR is acknowledged by the Tailscale
	// control plane).
	// TODO: This race condition has been encountered in the wild, implement
	//       retry / backoff / sleep.
	if !RoutesContain(rr.AdvertisedRoutes, cidr) {
		return fmt.Errorf("new route (%s) is not among list of advertised routes", cidr)
	}

	// If `cidr` is contained in `routes.EnabledRoutes`, there is nothing to do.
	if RoutesContain(rr.EnabledRoutes, cidr) {
		cli.Printf(ctx, "Device %s has already enabled route %s\n", device.ID, cidr)
		return nil
	}

	// ...otherwise, append it and call `SetRoutes()`
	routes := append(rr.EnabledRoutes, cidr.String())
	srr := cloud.SetRoutesRequest{DeviceID: device.ID, Routes: routes}
	cli.Printf(ctx, "Enabling route %s for device %s...\n", cidr, device.ID)
	_, err = cloud.SetRoutes(ctx, c, srr)
	if err != nil {
		return err
	}

	cli.Printf(ctx, "Enabled route %s for device %s\n", cidr, device.ID)
	return nil
}

// RoutesContain checks if a CIDR (`IPPrefix`) is contained in a slice of
// CIDR string. (This is a "fuzzy" check, it uses string value of the `IPPrefix`
// for comparison.) The type mismatch is due to the fact that the routes come
// from the Tailscale cloud API as a string slice.
func RoutesContain(routes []string, cidr netaddr.IPPrefix) bool {
	cidrString := cidr.String()
	for _, r := range routes {
		if r == cidrString {
			return true
		}
	}
	return false
}
