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
#  ./k8s-primary-init.sh NAMED_ARG1 NAMED_ARG2 ...
# Initializes a Kubernetes cluster on the "primary" control plane node. This
# node is special at cluster creation time because the cluster doesn't exist
# yet, but once the cluster exists, the "primary" node can be removed without
# issue (provided there are other control plane nodes).

set -e -x

echo "Not implemented" >&2
exit 1
