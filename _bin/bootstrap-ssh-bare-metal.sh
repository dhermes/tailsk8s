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
#  ./bootstrap-ssh-bare-metal.sh NETCAT_LISTEN_PORT
# Does minimal configuration on a **fresh** Ubuntu 20.04 bare metal machine
# by starting an SSH server and using `netcat` to receive authorized public
# SSH key(s) over the local network.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 1 ]
then
  echo "Usage: ./bootstrap-ssh-bare-metal.sh NETCAT_LISTEN_PORT" >&2
  exit 1
fi
NETCAT_LISTEN_PORT="${1}"

## Install All APT Packages Needed
#### - `netcat` for using a raw TCP listener
#### - SSH client and server (should already be present from previous steps)

sudo apt-get update
sudo apt-get install --yes \
  netcat \
  openssh-client \
  openssh-server

## Disable SSH Password Authentication and Restart SSH Server

sudo sed --in-place "s/.*PasswordAuthentication.*//g" /etc/ssh/sshd_config
cat <<EOF | sudo tee --append /etc/ssh/sshd_config
PasswordAuthentication no
EOF

sudo systemctl restart sshd.service

## Receive Authorized Key(s) from a Peer on the Local Network

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo "The list of all known IP addresses for this host is:"
hostname --all-ip-addresses

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo thanks | netcat -l "${NETCAT_LISTEN_PORT}" -b > "${HOME}/.extra_authorized_keys"

## Add Extra Authorized Key(s) to SSH.
#### NOTE: In order to ensure the keys received over the local network are
####       legitimate, a prompt will be used first.

echo "Received extra \`authorized_keys\`:"
echo '================================='
cat "${HOME}/.extra_authorized_keys"
echo '================================='

read -r -p ">> Accept extra authorized keys (y/n)? " ACCEPT_PROMPT
if [ "${ACCEPT_PROMPT}" != "y" ] && [ "${ACCEPT_PROMPT}" != "Y" ]
then
  echo "Rejected extra authorized keys" >&2
  rm --force "${HOME}/.extra_authorized_keys"
  exit 1
fi

touch "${HOME}/.ssh/authorized_keys"
chmod 644 "${HOME}/.ssh/authorized_keys"
cat "${HOME}/.extra_authorized_keys" >> "${HOME}/.ssh/authorized_keys"
rm --force "${HOME}/.extra_authorized_keys"
