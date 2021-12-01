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

// Package advertise uses local and cloud Tailscale APIs to advertise routes to the Tailnet.
//
// It uses the local API to advertise a route to peers and make sure peer
// routes are accepted locally. Then it uses the Tailscale Cloud API to accept
// the newly advertised route.
//
// This is provided in a way to optimize the testable surface area (even for
// untested parts of the code) without having any usage of `os.Exit()`.
package advertise
