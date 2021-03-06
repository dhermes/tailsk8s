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
#  ./new-machine.sh TAILSCALE_AUTHKEY_FILENAME
# Prepares a **fresh** Ubuntu 20.04 machine by installing required Debian
# packages, configuring `ufw` (Uncomplicated Firewall), enabling SSH (without
# password authentication) and joining a Tailscale Tailnet.

set -e -x

## Validate and read inputs

if [ "${#}" -ne 1 ]
then
  echo "Usage: ./new-machine.sh TAILSCALE_AUTHKEY_FILENAME" >&2
  exit 1
fi
TAILSCALE_AUTHKEY_FILENAME="${1}"

## Computed Variables

CURRENT_USER="$(whoami)"
CURRENT_HOSTNAME="$(hostname)"
OWNER_GROUP="$(id --user):$(id --group)"
K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap

## Ensure Tailscale Auth Key file exists

if [ ! -f "${TAILSCALE_AUTHKEY_FILENAME}" ]
then
  echo "No file located at ${TAILSCALE_AUTHKEY_FILENAME}" >&2
  exit 1
fi

## Ensure installation of **minimal** set of dependencies needed to add
## Tailscale and Docker custom APT repositories.

sudo apt-get update
sudo apt-get install --yes curl gnupg lsb-core

## Add Tailscale and Docker custom APT repositories.
#### - https://tailscale.com/download/linux/ubuntu-2004
#### - https://docs.docker.com/engine/install/ubuntu/
#### - https://docs.docker.com/engine/install/linux-postinstall/

curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.gpg \
  | sudo apt-key add -
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.list \
  | sudo tee /etc/apt/sources.list.d/tailscale.list
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

## Make sure the machine is up to date before installing new packages

sudo apt-get update
sudo apt-get --yes upgrade
sudo apt-get --yes dist-upgrade
sudo apt-get --yes autoremove

## Remove APT packages that may be outdated (really only relevant for machines
## that aren't a clean install)
#### - Older Docker installs

sudo apt-get remove --yes \
  containerd \
  docker \
  docker-engine \
  docker.io \
  runc

## Install all APT packages needed
#### - Networking tools used for Kubernetes (and debugging if needed);
####   `conntrack`, `ipset`, `socat`, `traceroute`
#### - Docker Engine client and server and containerd
#### - Tailscale
#### - Uncomplicated Firewall (ufw); this **should** be installed by default
#### - `envsubst` via `gettext-base` (will almost certainly be installed already)

sudo apt-get install --yes \
  conntrack \
  containerd.io \
  docker-ce \
  docker-ce-cli \
  ethtool \
  gettext-base \
  ipset \
  socat \
  tailscale \
  traceroute \
  ufw

## Ensure Timezone is UTC

echo 'Etc/UTC' | sudo tee /etc/timezone
sudo dpkg-reconfigure --frontend noninteractive tzdata

## Ensure SSH Password Authentication is disabled

if [ "$(grep '^PasswordAuthentication' /etc/ssh/sshd_config)" != "PasswordAuthentication no" ]
then
  echo "SSH Password Authentication is not disabled" >&2
  exit 1
fi

## Ensure the current user can use the Docker socket without `sudo`.
## This will (likely) not take effect until a new login shell.

sudo groupadd --force docker
sudo usermod --append --groups docker "${CURRENT_USER}"

## Prepare configuration bootstrap directory

sudo rm --force --recursive "${K8S_BOOTSTRAP_DIR}"
sudo mkdir --parents "${K8S_BOOTSTRAP_DIR}"
sudo chown "${OWNER_GROUP}" "${K8S_BOOTSTRAP_DIR}"

## Enable IP Forwarding for Tailscale
#### https://tailscale.com/kb/1104/enable-ip-forwarding/

if sudo test -f /etc/sysctl.d/tailscale.conf; then
  echo "/etc/sysctl.d/tailscale.conf exists, will be overwritten."
  sudo rm --force /etc/sysctl.d/tailscale.conf
fi

cat <<EOF | sudo tee /etc/sysctl.d/tailscale.conf
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
EOF

sudo sysctl --system

## Join Tailnet and remove the Tailscale Auth Key.
#### NOTE: For Tailnets where new devices must be manually authorized, the
####       `tailscale up` command will block until the current host is
####       authorized. The current host can be authorized in the web UI or the
####       `tailscale-authorize-linux-amd64-*` binary can be used to authorize
####       from the command line.

echo "Adding host ${CURRENT_HOSTNAME} to Tailnet..."
sudo tailscale up --authkey "file:${TAILSCALE_AUTHKEY_FILENAME}"
rm --force "${TAILSCALE_AUTHKEY_FILENAME}"

echo "IPv4 address in Tailnet: $(tailscale ip -4)"

## Set up `ufw` (Uncomplicated Firewall)
#### See: https://tailscale.com/kb/1077/secure-server-ubuntu-18-04/

sudo ufw allow in on tailscale0
sudo ufw allow 41641/udp
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw reload
sudo service ssh restart

sudo ufw status  # Sanity Check

echo "Close SSH session and re-connect over Tailscale"
echo "Current connection ::     SSH_CLIENT=${SSH_CLIENT}"
echo "Current connection :: SSH_CONNECTION=${SSH_CONNECTION}"
echo "(New) Tailscale IP :: $(tailscale ip -4)"
