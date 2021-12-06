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
#  ./k8s-control-plane-join.sh ADVERTISE_SUBNET
# Adds a **new** Kubernetes node to a control plane (the control plane must
# already exist, i.e. this isn't the first node).

set -e -x

## Validate and read inputs

if [ "${#}" -ne 1 ]
then
  echo "Usage: ./k8s-control-plane-join.sh ADVERTISE_SUBNET" >&2
  exit 1
fi
ADVERTISE_SUBNET="${1}"

## Computed Variables

CURRENT_HOSTNAME="$(hostname)"
HOST_IP="$(tailscale ip -4)"
K8S_BOOTSTRAP_DIR="/var/data/tailsk8s-bootstrap"
CA_CERT_HASH="sha256:$(cat "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt")"
CERTIFICATE_KEY="$(cat "${K8S_BOOTSTRAP_DIR}/certificate-key.txt")"
CONTROL_PLANE_LOAD_BALANCER="$(cat "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt")"
JOIN_TOKEN="$(cat "${K8S_BOOTSTRAP_DIR}/join-token.txt")"
TAILSCALE_API_KEY_FILENAME="${K8S_BOOTSTRAP_DIR}/tailscale-api-key.txt"
CONFIG_TEMPLATE_FILENAME="${K8S_BOOTSTRAP_DIR}/kubeadm-control-plane-join-config.yaml"

## Kubernetes Cluster Bootstrap

sudo rm --force --recursive /etc/kubernetes/

rm --force --recursive "${HOME}/.kube"
mkdir --parents "${HOME}/.kube"
cp "${K8S_BOOTSTRAP_DIR}/kube-config.yaml" "${HOME}/.kube/config"

## CNI via Kubenet (basic bridge mode)
#### See:
#### - https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/79a3f79b27bd28f82f071bb877a266c2e62ee506/docs/09-bootstrapping-kubernetes-workers.md#configure-cni-networking
#### - https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/#kubenet

sudo rm --force --recursive /etc/cni/net.d/
sudo mkdir --parents /etc/cni/net.d/

cat <<EOF | sudo tee /etc/cni/net.d/10-bridge.conf
{
    "cniVersion": "0.4.0",
    "name": "tailsk8s",
    "type": "bridge",
    "bridge": "cnio0",
    "isGateway": true,
    "ipMasq": true,
    "ipam": {
        "type": "host-local",
        "ranges": [
            [
                {
                    "subnet": "${ADVERTISE_SUBNET}"
                }
            ]
        ],
        "routes": [
            {
                "dst": "0.0.0.0/0"
            }
        ]
    }
}
EOF

cat <<EOF | sudo tee /etc/cni/net.d/99-loopback.conf
{
    "cniVersion": "0.4.0",
    "name": "lo",
    "type": "loopback"
}
EOF

## Advertise route managed by this node to Tailnet

sudo tailscale-advertise \
  --debug \
  --api-key "file:${TAILSCALE_API_KEY_FILENAME}" \
  --cidr "${ADVERTISE_SUBNET}"

## Configure `kubeadm`

echo "Populating \`kubeadm\` configuration via template:"
echo '================================================'
cat "${CONFIG_TEMPLATE_FILENAME}"

CA_CERT_HASH="${CA_CERT_HASH}" \
  CERTIFICATE_KEY="${CERTIFICATE_KEY}" \
  JOIN_TOKEN="${JOIN_TOKEN}" \
  CONTROL_PLANE_LOAD_BALANCER="${CONTROL_PLANE_LOAD_BALANCER}" \
  NODE_NAME="${CURRENT_HOSTNAME}" \
  HOST_IP="${HOST_IP}" \
  envsubst \
  < "${CONFIG_TEMPLATE_FILENAME}" \
  > "${HOME}/kubeadm-join-config.yaml"

## Run `kubeadm join`

sudo kubeadm join \
  --config "${HOME}/kubeadm-join-config.yaml"

## Label the newly added node with `tailsk8s`` label(s)

kubectl label node \
  "${CURRENT_HOSTNAME}" \
  "tailsk8s.io/advertise-subnet=${ADVERTISE_SUBNET/\//__}"

## Clean up temporary files

rm --force "${HOME}/kubeadm-join-config.yaml"
