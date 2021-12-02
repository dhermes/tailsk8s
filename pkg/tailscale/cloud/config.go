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

import (
	"context"
	"net/http"

	tailscalecli "github.com/dhermes/tailsk8s/pkg/tailscale/cli"
)

// Config provides helpers that are required to interact with the Tailscale
// Cloud API.
type Config struct {
	Addr    string
	Tailnet string
	APIKey  string
}

// NewConfig returns a new `Config` with all relevant defaults provided and
// options for overriding.
func NewConfig(opts ...Option) (Config, error) {
	c := Config{Addr: "https://api.tailscale.com"}
	for _, opt := range opts {
		err := opt(&c)
		if err != nil {
			return Config{}, err
		}
	}
	return c, nil
}

// HTTPClient returns an HTTP client associated with this config.
//
// NOTE: For now this is just a stub wrapper around `http.DefaultClient` but
//       it's provided here to make the code easier to test at a later date.
func (c Config) HTTPClient() *http.Client {
	return http.DefaultClient
}

// Resolve sets defaults based on default conventions or based on the local
// environment.
// - `Addr` defaults to `https://api.tailscale.com`
// - If `APIKey` is prefixed with `file:`, the file will be read from the
//   filesystem
// - If `Tailnet is unset, the local `tailscaled` API will be used to query
//   for the magic DNS name.
func (c *Config) Resolve(ctx context.Context) error {
	apiKey, err := tailscalecli.ReadAPIKey(ctx, c.APIKey)
	if err != nil {
		return err
	}
	tailnet, err := tailscalecli.DefaultTailnet(ctx, c.Tailnet)
	if err != nil {
		return err
	}

	c.Addr = stringDefault(c.Addr, "https://api.tailscale.com")
	c.APIKey = apiKey
	c.Tailnet = tailnet
	return nil
}

func stringDefault(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return s1
}
