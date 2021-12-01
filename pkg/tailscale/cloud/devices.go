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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

const (
	// debugCurlAuthorizeDevice is a template to print (in debug mode) the
	// equivalent curl command to the outgoing request. The POST body is
	// not expected to be `shlex` quoted by the template user, but it should be.
	debugCurlAuthorizeDevice = `Calling "authorize device" API route:
> curl \
>   --user "...redacted API Key...:" \
>   --data-binary '%s'
>   %s
`
)

// Device represents a device in a Tailnet; the set of fields here is not
// intended to be comprehensive, it is only intended to match the known
// uses cases. For example there is a `blocksIncomingConnections bool` field
// we don't have use for.
type Device struct {
	Addresses  []string `json:"addresses"`
	Authorized bool     `json:"authorized"`
	Hostname   string   `json:"hostname"`
	ID         string   `json:"id"`
	Name       string   `json:"name"`
}

// AuthorizeDevice marks a device as authorized.
//
// This is only needed in Tailnets where device authorization is required.
func AuthorizeDevice(ctx context.Context, c Config, adr AuthorizeDeviceRequest) (*Empty, error) {
	url := fmt.Sprintf(
		"%s/api/v2/device/%s/authorized",
		c.Addr,
		url.PathEscape(adr.ID),
	)
	asJSON, err := json.Marshal(adr)
	if err != nil {
		return nil, err
	}

	cli.DebugPrintf(ctx, debugCurlAuthorizeDevice, string(asJSON), url)
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to authorize (status %d, body %q)", resp.StatusCode, body)
	}
	return &Empty{}, nil
}
