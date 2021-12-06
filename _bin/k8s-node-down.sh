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
#  ./k8s-node-down.sh
# Removes a node from a Kubernetes cluster. See
# https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#tear-down

set -e -x

## Validate and read inputs

if [ "${#}" -ne 0 ]
then
  echo "Usage: ./k8s-node-down.sh" >&2
  exit 1
fi

## Computed Variables

CURRENT_HOSTNAME="$(hostname)"
OWNER_GROUP="$(id --user):$(id --group)"
K8S_BOOTSTRAP_DIR="/var/data/tailsk8s-bootstrap"
TAILSCALE_API_KEY_FILENAME="${K8S_BOOTSTRAP_DIR}/tailscale-api-key.txt"
ADVERTISE_SUBNET_NORMALIZED="$(kubectl get node "${CURRENT_HOSTNAME}" --output go-template='{{ index .metadata.labels "tailsk8s.io/advertise-subnet" }}')"
ADVERTISE_SUBNET="${ADVERTISE_SUBNET_NORMALIZED/__/\/}"

## Drain Node

kubectl drain \
  "${CURRENT_HOSTNAME}" \
  --delete-emptydir-data \
  --force \
  --ignore-daemonsets

## Run `kubeadm reset`

sudo systemctl disable --now kubelet
sudo kubeadm reset --force

## Remove the Node from Kubernetes

kubectl delete node "${CURRENT_HOSTNAME}"

## Withdraw route managed by this node from Tailnet

sudo tailscale-withdraw \
  --debug \
  --api-key "file:${TAILSCALE_API_KEY_FILENAME}" \
  --cidr "${ADVERTISE_SUBNET}"

## Fully remove Kubernetes configuration directores

rm --force --recursive "${HOME}/.kube"
sudo rm --force --recursive /etc/cni/net.d/
sudo rm --force --recursive /etc/kubernetes/

## Clear Kubernetes bootstrap directory

sudo rm --force --recursive "${K8S_BOOTSTRAP_DIR}"
sudo mkdir --parents "${K8S_BOOTSTRAP_DIR}"
sudo chown "${OWNER_GROUP}" "${K8S_BOOTSTRAP_DIR}"
