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
#  ./k8s-join-refresh.sh
# Refresh the `kubeadm` join credentials on a machine that is already in
# the cluster.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 0 ]
then
  echo "Usage: ./k8s-join-refresh.sh" >&2
  exit 1
fi

## Computed Variables

K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap

## Generate and Store a New Join Token

rm --force "${K8S_BOOTSTRAP_DIR}/join-token.txt"
kubeadm token create --description 're-join after expiration' > "${K8S_BOOTSTRAP_DIR}/join-token.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/join-token.txt"

## Upload the CA Private Key(s) as an Encrypted Secret

kubeadm certs certificate-key > "${HOME}/new-certificate-key.txt"
sudo kubeadm init phase upload-certs \
  --certificate-key "$(cat "${HOME}/new-certificate-key.txt")" \
  --upload-certs \
  --skip-certificate-key-print

rm --force "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
mv "${HOME}/new-certificate-key.txt" "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
