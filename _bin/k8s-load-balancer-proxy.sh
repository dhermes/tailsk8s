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
#  ./k8s-load-balancer-proxy.sh CONTROL_PLANE [CONTROL_PLANE ...]
# Starts or restarts HAProxy to act as a load balancer for the the Kubernetes
# API servers (running on control plane nodes). In a cloud environment a
# load balancer is a core primitive but in a bare metal cluster we approximate
# it by running HAProxy on a "spare" machine. This may be run repeatedly as
# control plane nodes join or leave the cluster, forcing the list of hosts
# behind the load balancer to be updated.
#
# The `CONTROL_PLANE` inputs are expected to be of the form
# `{TAILSCALE_HOST} {TAILSCALE_IP}`, for example
# `pedantic-yonath 100.110.217.104`.

set -e -x

## Validate and read inputs

if [ "${#}" -eq 0 ]
then
  echo "Usage: ./k8s-load-balancer-proxy.sh CONTROL_PLANE [CONTROL_PLANE ...]" >&2
  exit 1
fi

## Computed Variables

HOST_IP="$(tailscale ip -4)"

## Enable non-local IPv4 bind for HAProxy

if sudo test -f /etc/sysctl.d/haproxy.conf; then
    echo "/etc/sysctl.d/haproxy.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/haproxy.conf
fi

cat <<EOF | sudo tee /etc/sysctl.d/haproxy.conf
net.ipv4.ip_nonlocal_bind = 1
EOF

sudo sysctl --system

## Backup HAProxy Config

sudo mv /etc/haproxy/haproxy.cfg /etc/haproxy/haproxy.cfg.backup

## Configure HAProxy

#### See: https://github.com/kubernetes/kubeadm/blob/e55c2a2b8e0b4e3079fd6a3586baf6472700428b/docs/ha-considerations.md#haproxy-configuration

cat <<EOF | sudo tee /etc/haproxy/haproxy.cfg
#---------------------------------------------------------------------
# Global settings
#---------------------------------------------------------------------
global
     log /dev/log local0
     log /dev/log local1 notice
     daemon
     user haproxy
     group haproxy

#---------------------------------------------------------------------
# common defaults that all the 'listen' and 'backend' sections will
# use if not designated in their block
#---------------------------------------------------------------------
defaults
     mode http
     log global
     option httplog
     option dontlognull
     option http-server-close
     option forwardfor except 127.0.0.0/8
     option redispatch
     retries 1
     timeout http-request    10s
     timeout queue           20s
     timeout connect         5s
     timeout client          20s
     timeout server          20s
     timeout http-keep-alive 10s
     timeout check           10s

#---------------------------------------------------------------------
# apiserver frontend which proxys to the control plane nodes
#---------------------------------------------------------------------
frontend apiserver
     bind ${HOST_IP}:6443
     mode tcp
     option tcplog
     default_backend apiserver

#---------------------------------------------------------------------
# round robin balancing for apiserver
#---------------------------------------------------------------------
backend apiserver
     option httpchk GET /healthz
     http-check expect status 200
     mode tcp
     option ssl-hello-chk
     balance roundrobin
EOF

for CONTROL_PLANE in "${@}"
do
  echo "     server ${CONTROL_PLANE}:6443 check fall 3 rise 2" | sudo tee --append /etc/haproxy/haproxy.cfg
done

# Ensure HAProxy is enabled (will run on reboot) and restart to reload new
# configuration

sudo systemctl enable haproxy --now
sudo systemctl restart haproxy
