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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

const (
	debugCurlGetRoutes = `Calling "get routes" cloud API route:
> curl \
>   --include \
>   --user "...redacted API Key...:" \
>   %s
`
	// debugCurlSetRoutes is a template to print (in debug mode) the
	// equivalent curl command to the outgoing request. The POST body is
	// not expected to be `shlex` quoted by the template user, but it should be.
	debugCurlSetRoutes = `Calling "set routes" cloud API route:
> curl \
>   --include \
>   --user "...redacted API Key...:" \
>   --data-binary '%s'
>   %s
`
)

// GetRoutes fetches subnet routes that are advertised and enabled for a device.
func GetRoutes(ctx context.Context, c Config, grr GetRoutesRequest) (*RoutesResponse, error) {
	url := fmt.Sprintf(
		"%s/api/v2/device/%s/routes",
		c.Addr,
		url.PathEscape(grr.DeviceID),
	)
	cli.DebugPrintf(ctx, debugCurlGetRoutes, url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.APIKey, "")
	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get routes (status %d, body %q)", resp.StatusCode, body)
	}

	var response RoutesResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SetRoutes sets the subnet routes that are enabled for a device.
func SetRoutes(ctx context.Context, c Config, srr SetRoutesRequest) (*RoutesResponse, error) {
	url := fmt.Sprintf(
		"%s/api/v2/device/%s/routes",
		c.Addr,
		url.PathEscape(srr.DeviceID),
	)
	asJSON, err := json.Marshal(srr)
	if err != nil {
		return nil, err
	}

	cli.DebugPrintf(ctx, debugCurlSetRoutes, string(asJSON), url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(asJSON))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.APIKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to set routes (status %d, body %q)", resp.StatusCode, body)
	}

	var response RoutesResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
