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

package cloud

// Empty is a type with no fields, for API routes which have no inputs and / or
// have no outputs.
type Empty struct{}

// GetDevicesResponse is the response for the `GET /api/v2/tailnet/:t/devices`
// API route.
type GetDevicesResponse struct {
	Devices []Device `json:"devices"`
}

// AuthorizeDeviceRequest is the request for the `POST /api/v2/device/:d/authorized`
// API route.
type AuthorizeDeviceRequest struct {
	DeviceID   string `json:"-"`
	Authorized bool   `json:"authorized"`
}

// GetRoutesRequest is the request for the `GET /api/v2/device/:d/routes`
// API route.
type GetRoutesRequest struct {
	DeviceID string `json:"-"`
}

// RoutesResponse is the response for the `GET /api/v2/device/:d/routes`
// and `POST /api/v2/device/:d/routes` API routes.
type RoutesResponse struct {
	AdvertisedRoutes []string `json:"advertisedRoutes"`
	EnabledRoutes    []string `json:"enabledRoutes"`
}

// SetRoutesRequest is the request for the `POST /api/v2/device/:d/routes`
// API route.
type SetRoutesRequest struct {
	DeviceID string   `json:"-"`
	Routes   []string `json:"routes"`
}
