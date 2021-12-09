#!/bin/bash
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

# Usage:
#  ./k8s-load-balancer-down.sh
# Disables and stops HAProxy load balancer and cleans up any specialized
# configs.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 0 ]
then
  echo "Usage: ./k8s-final-down.sh" >&2
  exit 1
fi

## Remove HAProxy `sysctl` modifications

sudo rm --force /etc/sysctl.d/haproxy.conf
sudo sysctl --system

## Restore Backed Up HAProxy Config

sudo rm --force /etc/haproxy/haproxy.cfg
sudo mv /etc/haproxy/haproxy.cfg.backup /etc/haproxy/haproxy.cfg

# Ensure HAProxy is disabled and stopped

sudo systemctl disable haproxy --now
sudo systemctl stop haproxy
