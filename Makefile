# Copyright 2021 Danny Hermes
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: help
help:
	@echo 'Makefile for the `tailsk8s` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make tailscale-advertise-linux-amd64    Build static `tailscale-advertise` binary for linux/amd64'
	@echo '   make tailscale-authorize-linux-amd64    Build static `tailscale-authorize` binary for linux/amd64'
	@echo '   make tailscale-withdraw-linux-amd64     Build static `tailscale-withdraw` binary for linux/amd64'
	@echo '   make release                            Build all static binaries'
	@echo ''

################################################################################
# Meta-variables
################################################################################
VERSION ?= $(shell git log -1 --pretty=%H 2> /dev/null)
UPX_BIN := $(shell command -v upx 2> /dev/null)

# NOTE: Targets to build Go binaries are marked `.PHONY` even though they
#       produce real files. We do this intentionally to defer to Go's build
#       caching and related tooling rather than relying on `make` for this.
#
#       For more on strategies to keep binaries small, see:
#       https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/

.PHONY: tailscale-advertise-linux-amd64
tailscale-advertise-linux-amd64: _require-upx _require-version
	rm -f "./_bin/tailscale-advertise-linux-amd64-"*
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -installsuffix static -o "./_bin/tailscale-advertise-linux-amd64-$(VERSION)" ./cmd/tailscale-advertise/
	upx -q -9 "./_bin/tailscale-advertise-linux-amd64-$(VERSION)"

.PHONY: tailscale-authorize-linux-amd64
tailscale-authorize-linux-amd64: _require-upx _require-version
	rm -f "./_bin/tailscale-authorize-linux-amd64-"*
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -installsuffix static -o "./_bin/tailscale-authorize-linux-amd64-$(VERSION)" ./cmd/tailscale-authorize/
	upx -q -9 "./_bin/tailscale-authorize-linux-amd64-$(VERSION)"

.PHONY: tailscale-withdraw-linux-amd64
tailscale-withdraw-linux-amd64: _require-upx _require-version
	rm -f "./_bin/tailscale-withdraw-linux-amd64-"*
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -installsuffix static -o "./_bin/tailscale-withdraw-linux-amd64-$(VERSION)" ./cmd/tailscale-withdraw/
	upx -q -9 "./_bin/tailscale-withdraw-linux-amd64-$(VERSION)"

.PHONY: release
release: tailscale-advertise-linux-amd64 tailscale-authorize-linux-amd64 tailscale-withdraw-linux-amd64

################################################################################
# Doctor Commands (these do not show up in `make help`)
################################################################################

.PHONY: _require-upx
_require-upx:
ifndef UPX_BIN
	$(error 'upx is not installed, it can be installed via "apt-get install upx", "apk add upx" or "brew install upx".')
endif

.PHONY: _require-version
_require-version:
ifeq ($(VERSION),)
	$(error 'VERSION variable is not set.')
endif
ifndef VERSION
	$(error 'VERSION variable is not set.')
endif
