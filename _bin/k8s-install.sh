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
#  ./k8s-install.sh NAMED_ARG1 NAMED_ARG2 ...
# Install and configure all Kubernetes specific dependencies on a machine
# (bare metal or cloud VM). In particular:
# - Configure the Docker daemon to use `systemd` as `cgroup` driver
# - Allow `iptables` to see bridged traffic
# - Install `kubeadm`, `kubelet` and `kubectl` (v1.22.4)
# - Install core Kubernetes CNI (container networking) plugins (`bandwith`,
#   `bridge`, `dhcp`, etc.)
# - Install core CRI (container runtime) tools `crictl`
# - Install and enable `kubelet` as a `systemd` unit (with `kubeadm` support)
# - Install networking tools needed by Kubernetes (e.g. `conntrack`)
# - Disable swap for the hard drive
# - Prefetch all images needed by `kubeadm`

set -e -x

echo "Not implemented" >&2
exit 1
