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

// Package authorize uses the cloud API to authorize a new machine in a Tailnet.
//
// This package glues together functions in `pkg/tailscale/cloud` to authorize
// a new machine in a Tailnet without having to visit the web UI.
//
// This is provided in a way to optimize the testable surface area (even for
// untested parts of the code) without having any usage of `os.Exit()`.
package authorize
