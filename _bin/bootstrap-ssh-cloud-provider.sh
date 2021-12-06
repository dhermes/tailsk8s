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
#  ./bootstrap-ssh-cloud-provider.sh DESIRED_HOSTNAME EXTRA_AUTHORIZED_KEYS_FILENAME
# Does minimal configuration on a **fresh** Ubuntu 20.04 VM in any cloud
# provider by changing the hostname and adding authorized public SSH key(s).
# (The assumption is that the cloud provider has already set up SSH on the
# machine and the extra authorized keys were placed via `scp`.)

set -e -x

## Validate and read inputs

if [ "${#}" -ne 2 ]
then
  echo "Usage: ./bootstrap-ssh-cloud-provider.sh DESIRED_HOSTNAME EXTRA_AUTHORIZED_KEYS_FILENAME" >&2
  exit 1
fi
DESIRED_HOSTNAME="${1}"
EXTRA_AUTHORIZED_KEYS_FILENAME="${2}"

## Ensure hostname matches desired Tailscale machine name

sudo hostnamectl set-hostname "${DESIRED_HOSTNAME}"

## Ensure extra `authorized_keys` file exists

if [ ! -f "${EXTRA_AUTHORIZED_KEYS_FILENAME}" ]
then
    echo "No file located at ${EXTRA_AUTHORIZED_KEYS_FILENAME}" >&2
    exit 1
fi

## Add Extra Authorized Key(s)

touch "${HOME}/.ssh/authorized_keys"
chmod 644 "${HOME}/.ssh/authorized_keys"

cat "${EXTRA_AUTHORIZED_KEYS_FILENAME}" >> "${HOME}/.ssh/authorized_keys"
rm --force "${EXTRA_AUTHORIZED_KEYS_FILENAME}"
