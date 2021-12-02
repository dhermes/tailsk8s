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

package cli

import (
	"context"
	"os"
	"strings"

	"github.com/dhermes/tailsk8s/pkg/cli"
)

// ReadAPIKey reads a Tailscale API key from file if the provided `apiKey`
// is prefixed with `file:`. The assumption is that this value has been passed
// via a CLI flag such as `--api-key`.
func ReadAPIKey(ctx context.Context, apiKey string) (string, error) {
	if !strings.HasPrefix(apiKey, "file:") {
		return apiKey, nil
	}

	filename := strings.TrimPrefix(apiKey, "file:")
	cli.Printf(ctx, "Reading Tailscale API key from: %s\n", filename)
	apiKeyBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(apiKeyBytes), nil
}
