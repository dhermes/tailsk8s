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
#  ./k8s-install.sh
# Install and configure all Kubernetes specific dependencies on a machine
# (bare metal or cloud VM).

set -e -x

## Validate and read inputs

if [ "${#}" -ne 0 ]
then
  echo "Usage: ./k8s-install.sh" >&2
  exit 1
fi

## Input Variables

ARCH="amd64"
CNI_VERSION="v0.8.2"
CRICTL_VERSION="v1.22.0"
K8S_VERSION="v1.22.4"
K8S_RELEASE_VERSION="v0.4.0"  # NOTE: This is the version of the `kubernetes/release` project
K8S_DOWNLOAD_DIR=/usr/local/bin

## Use `systemd` as cgroup driver for Docker

#### H/T: https://kubernetes.io/docs/setup/production-environment/container-runtimes/
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
sudo systemctl enable docker
sudo systemctl daemon-reload
sudo systemctl restart docker

## Let `iptables`` see bridged traffic

if sudo test -f /etc/modules-load.d/k8s.conf; then
    echo "/etc/modules-load.d/k8s.conf exists, will be overwritten."
    sudo rm --force /etc/modules-load.d/k8s.conf
fi
if sudo test -f /etc/sysctl.d/k8s.conf; then
    echo "/etc/sysctl.d/k8s.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/k8s.conf
fi

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

sudo sysctl --system

## Install kubeadm, kubelet and kubectl

cd "${K8S_DOWNLOAD_DIR}"
sudo curl --location --remote-name-all \
  "https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/{kubeadm,kubelet,kubectl}"
sudo chmod +x {kubeadm,kubelet,kubectl}

## Install standard CNI plugins

sudo rm --force --recursive /opt/cni/bin
sudo mkdir --parents /opt/cni/bin
curl --location \
  "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" \
  | sudo tar --directory /opt/cni/bin --extract --gzip

## Install `crictl`

sudo mkdir --parents "${K8S_DOWNLOAD_DIR}"
curl --location \
  "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" \
  | sudo tar --directory "${K8S_DOWNLOAD_DIR}" --extract --gzip

## Configure systemd to run `kubelet` unit and support `kubeadm`

curl --silent --show-error --location \
  "https://raw.githubusercontent.com/kubernetes/release/${K8S_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" \
  | sed "s:/usr/bin:${K8S_DOWNLOAD_DIR}:g" \
  | sudo tee /etc/systemd/system/kubelet.service
sudo mkdir --parents /etc/systemd/system/kubelet.service.d
curl --silent --show-error --location \
  "https://raw.githubusercontent.com/kubernetes/release/${K8S_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" \
  | sed "s:/usr/bin:${K8S_DOWNLOAD_DIR}:g" \
  | sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

## Disable Swap

sudo swapoff --all
free --human  ## Sanity Check
sudo sed -i '/ swap / s/^/#/' /etc/fstab

## Pre-fetch all images used by `kubeadm`

kubeadm config images pull
