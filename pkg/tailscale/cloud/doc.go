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

// Package cloud provides helpers for using the Tailscale "cloud API".
//
// Here the "cloud API" refers to the API provided by the control plane. When
// using the Tailscale-provided control plane `controlplane.tailscale.com`,
// the cloud API is expected to be provided by `api.tailscale.com`.
//
// The cloud API is in contrast to the local API, which is provided by the
// `tailscaled` socket.
//
// See: https://github.com/tailscale/tailscale/blob/v1.18.1/api.md
package cloud
