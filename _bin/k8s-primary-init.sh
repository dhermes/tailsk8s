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
#  ./k8s-primary-init.sh CLUSTER_NAME POD_SUBNET SERVICE_SUBNET ADVERTISE_SUBNET CONTROL_PLANE_LOAD_BALANCER
# Initializes a Kubernetes cluster on the "primary" control plane node. This
# node is special at cluster creation time because the cluster doesn't exist
# yet, but once the cluster exists, the "primary" node can be removed without
# issue (provided there are other control plane nodes).
#
# To distribute bootstrap configuration to other nodes, a dedicated directory
# (/var/data/tailsk8s-bootstrap/) will be essentially as a key-value store to
# stash configuration files that can be securely copied (via `scp`) back to the
# host that is bringing up the Kubernetes cluster.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 5 ]
then
  echo "Usage: ./k8s-primary-init.sh CLUSTER_NAME POD_SUBNET SERVICE_SUBNET ADVERTISE_SUBNET CONTROL_PLANE_LOAD_BALANCER" >&2
  exit 1
fi
CLUSTER_NAME="${1}"
POD_SUBNET="${2}"
SERVICE_SUBNET="${3}"
ADVERTISE_SUBNET="${4}"
CONTROL_PLANE_LOAD_BALANCER="${5}"

## Computed Variables

CURRENT_HOSTNAME="$(hostname)"
OWNER_GROUP="$(id --user):$(id --group)"
HOST_IP="$(tailscale ip -4)"
K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap
TAILSCALE_API_KEY_FILENAME="${K8S_BOOTSTRAP_DIR}/tailscale-api-key"
CONFIG_TEMPLATE_FILENAME="${K8S_BOOTSTRAP_DIR}/kubeadm-init-config.yaml"

## Ensure `kubeadm-init-config.yaml` template file exists

if [ ! -f "${CONFIG_TEMPLATE_FILENAME}" ]
then
  echo "No file located at ${CONFIG_TEMPLATE_FILENAME}" >&2
  exit 1
fi

## Ensure Tailscale API Key file exists

if [ ! -f "${TAILSCALE_API_KEY_FILENAME}" ]
then
  echo "No file located at ${TAILSCALE_API_KEY_FILENAME}" >&2
  exit 1
fi

## Kubernetes Cluster Bootstrap (Before)

rm --force "${K8S_BOOTSTRAP_DIR}/join-token.txt"
kubeadm token generate > "${K8S_BOOTSTRAP_DIR}/join-token.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/join-token.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
kubeadm certs certificate-key > "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"
echo "${CONTROL_PLANE_LOAD_BALANCER}" > "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"
chmod 444 "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"

## Computed Variables Part Two (After Writing to ${K8S_BOOTSTRAP_DIR})

JOIN_TOKEN="$(cat "${K8S_BOOTSTRAP_DIR}/join-token.txt")"
CERTIFICATE_KEY="$(cat "${K8S_BOOTSTRAP_DIR}/certificate-key.txt")"

## CNI via `kubenet` (basic bridge mode)
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

## Advertise routes managed by this node to Tailnet

sudo tailscale-advertise \
  --debug \
  --api-key "file:${TAILSCALE_API_KEY_FILENAME}" \
  --cidr "${ADVERTISE_SUBNET}"

## Configure `kubeadm`

echo "Populating \`kubeadm\` configuration via template:"
echo '================================================'
cat "${CONFIG_TEMPLATE_FILENAME}"

CERTIFICATE_KEY="${CERTIFICATE_KEY}" \
  JOIN_TOKEN="${JOIN_TOKEN}" \
  CLUSTER_NAME="${CLUSTER_NAME}" \
  POD_SUBNET="${POD_SUBNET}" \
  SERVICE_SUBNET="${SERVICE_SUBNET}" \
  CONTROL_PLANE_LOAD_BALANCER="${CONTROL_PLANE_LOAD_BALANCER}" \
  NODE_NAME="${CURRENT_HOSTNAME}" \
  HOST_IP="${HOST_IP}" \
  envsubst \
  < "${CONFIG_TEMPLATE_FILENAME}" \
  > "${HOME}/kubeadm-init-config.yaml"

## Run `kubeadm init`

sudo rm --force --recursive /etc/kubernetes/
sudo rm --force --recursive /var/lib/etcd/
sudo rm --force --recursive /var/lib/kubelet/
sudo kubeadm init \
  --config "${HOME}/kubeadm-init-config.yaml" \
  --upload-certs \
  --skip-token-print \
  --skip-certificate-key-print
rm --force "${HOME}/kubeadm-init-config.yaml"

## Set Up Kubernetes Config

rm --force --recursive "${HOME}/.kube"
mkdir --parents "${HOME}/.kube"

sudo cp /etc/kubernetes/admin.conf "${HOME}/.kube/config"
sudo chown "${OWNER_GROUP}" "${HOME}/.kube/config"

## Label the newly added node with `tailsk8s` label(s)

kubectl label node \
  "${CURRENT_HOSTNAME}" \
  "tailsk8s.io/advertise-subnet=${ADVERTISE_SUBNET/\//__}"

## Kubernetes Cluster Bootstrap (After)
#### See: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-join/#token-based-discovery-with-ca-pinning

rm --force "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt \
  | openssl rsa -pubin -outform der 2>/dev/null \
  | openssl dgst -sha256 -hex \
  | sed 's/^.* //' \
  > "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"
chmod 444 "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
cp "${HOME}/.kube/config" "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
chmod 444 "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
