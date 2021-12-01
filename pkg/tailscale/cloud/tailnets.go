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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

const (
	debugCurlGetDevices = `Calling "get devices in Tailnet" API route:
> curl \
>   --user "...redacted API Key...:" \
>   %s
`
)

// GetDevices lists the devices for a Tailnet.
func GetDevices(ctx context.Context, c Config, _ Empty) (*GetDevicesResponse, error) {
	url := fmt.Sprintf(
		"%s/api/v2/tailnet/%s/devices",
		c.Addr,
		url.PathEscape(c.Tailnet),
	)
	cli.DebugPrintf(ctx, debugCurlGetDevices, url)
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get devices (status %d, body %q)", resp.StatusCode, body)
	}

	var gdr GetDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&gdr)
	if err != nil {
		return nil, err
	}

	return &gdr, nil
}
