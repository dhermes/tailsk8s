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
#  ./k8s-control-plane-join.sh NAMED_ARG1 NAMED_ARG2 ...
# Adds a **new** Kubernetes node to a control plane (the control plane must
# already exists, i.e. this isn't the first node).

set -e -x

echo "Not implemented" >&2
exit 1

####- List of known values that can be re-used as needed:
####- ${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt
####- ${K8S_BOOTSTRAP_DIR}/certificate-key.txt
####- ${K8S_BOOTSTRAP_DIR}/cluster-name.txt
####- ${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt
####- ${K8S_BOOTSTRAP_DIR}/join-token.txt
####- ${K8S_BOOTSTRAP_DIR}/kube-config.yaml
