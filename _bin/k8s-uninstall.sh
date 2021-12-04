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
#  ./k8s-uninstall.sh
# Uninstall all Kubernetes specific dependencies on a machine and remove all
# related configuration.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 0 ]
then
  echo "Usage: ./k8s-uninstall.sh" >&2
  exit 1
fi

## Remove Kubernetes / CRI / CNI binaries

sudo rm --force \
  /usr/local/bin/crictl \
  /usr/local/bin/kubeadm \
  /usr/local/bin/kubectl \
  /usr/local/bin/kubelet
sudo rm --force --recursive /opt/cni/bin

## Remove Docker Daemon Configuration (`systemd`` as cgroup driver)

sudo rm --force --recursive /etc/docker/daemon.json
sudo systemctl enable docker
sudo systemctl daemon-reload
sudo systemctl restart docker

## Remove sysctl / lsmod modifications

sudo rm --force /etc/modules-load.d/k8s.conf
sudo rm --force /etc/sysctl.d/k8s.conf
sudo sysctl --system

## Remove the `kubelet` unit from `systemd`

sudo systemctl disable --now kubelet
sudo rm --force /etc/systemd/system/kubelet.service
sudo rm --force --recursive /etc/systemd/system/kubelet.service.d
sudo systemctl reset-failed
sudo systemctl daemon-reload
