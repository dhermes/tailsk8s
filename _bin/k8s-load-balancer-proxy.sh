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
#  ./k8s-load-balancer-proxy.sh NAMED_ARG1 NAMED_ARG2 ...
# Starts or restarts HAProxy to act as a load balancer for the the Kubernetes
# API servers (running on control plane nodes). In a cloud environment a
# load balancer is a core primitive but in a bare metal cluster we approximate
# it by running HAProxy on a spare machine. This may be run repeatedly as
# control plane nodes join or leave the cluster.

set -e -x

## Enable non-local IPv4 bind for HAProxy

if sudo test -f /etc/sysctl.d/haproxy.conf; then
    echo "/etc/sysctl.d/haproxy.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/haproxy.conf
fi

cat <<EOF | sudo tee /etc/sysctl.d/haproxy.conf
net.ipv4.ip_nonlocal_bind = 1
EOF

sudo sysctl --system
