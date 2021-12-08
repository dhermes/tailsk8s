# Bringing up a New Machine

This assumes the new machine has been bootstrapped so that the jump host
can SSH in. Use the `new-machine.sh` [script][4] to

```bash
SSH_TARGET=dhermes@192.168.7.131
TAILSCALE_AUTHKEY_FILENAME=./tailscale-one-off-key-KC

scp \
  _bin/new-machine.sh \
  "${TAILSCALE_AUTHKEY_FILENAME}" \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

Then on the new machine:

```bash
TAILSCALE_AUTHKEY_FILENAME=~/tailscale-one-off-key-KC

./new-machine.sh "${TAILSCALE_AUTHKEY_FILENAME}"
rm ./new-machine.sh
```

Below, let's dive into what `new-machine.sh` does.

## Add Tailscale and Docker Custom APT repositories.

See Tailscale [installation][1] and Docker [installation][2] and
[post-installation][3] documentation.

```bash
sudo apt-get update
sudo apt-get install --yes curl gnupg lsb-core

curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.gpg \
  | sudo apt-key add -
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.list \
  | sudo tee /etc/apt/sources.list.d/tailscale.list
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

## Install All Dependencies

```bash
sudo apt-get update
sudo apt-get --yes upgrade
sudo apt-get --yes dist-upgrade
sudo apt-get --yes autoremove

sudo apt-get remove --yes \
  containerd \
  docker \
  docker-engine \
  docker.io \
  runc

sudo apt-get install --yes \
  conntrack \
  containerd.io \
  docker-ce \
  docker-ce-cli \
  gettext-base \
  ipset \
  socat \
  tailscale \
  traceroute \
  ufw
```

## Ensure Timezone is UTC

```bash
echo 'Etc/UTC' | sudo tee /etc/timezone
sudo dpkg-reconfigure --frontend noninteractive tzdata
```

## Add Current User to `docker` Group

```bash
sudo groupadd --force docker
sudo usermod --append --groups docker "$(whoami)"
```

## Prepare Configuration Bootstrap Directory

```bash
K8S_BOOTSTRAP_DIR="/var/data/tailsk8s-bootstrap"

sudo rm --force --recursive "${K8S_BOOTSTRAP_DIR}"
sudo mkdir --parents "${K8S_BOOTSTRAP_DIR}"
sudo chown "$(id --user):$(id --group)" "${K8S_BOOTSTRAP_DIR}"
```

## Enable IP Forwarding for Tailscale

See [Enable IP forwarding on Linux][5] FAQ from Tailscale:

```bash
if sudo test -f /etc/sysctl.d/tailscale.conf; then
    echo "/etc/sysctl.d/tailscale.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/tailscale.conf
fi

cat <<EOF | sudo tee /etc/sysctl.d/tailscale.conf
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
EOF

sudo sysctl --system
```

## Join Tailnet

When `tailscale up` gets run on the new machine

```bash
sudo tailscale up --authkey "file:${TAILSCALE_AUTHKEY_FILENAME}"
rm --force "${TAILSCALE_AUTHKEY_FILENAME}"
```

if device authorization is enabled in the Tailnet, then the command will block
until the machine is authorized with:

```
To authorize your machine, visit (as admin):

        https://login.tailscale.com/admin/machines
```

On the jump host, the machine can be authorized using the Tailscale API key:

```bash
TAILSCALE_DEVICE_NAME=pedantic-yonath

./_bin/tailscale-authorize-linux-amd64-v1.20211203.1 \
  --debug \
  --hostname "${TAILSCALE_DEVICE_NAME}" \
  --api-key file:./k8s-bootstrap-shared/tailscale-api-key
```

## Set Up Uncomplicated Firewall (`ufw`)

See [Use UFW to lock down an Ubuntu server][6] FAQ from Tailscale:

```bash
sudo ufw allow in on tailscale0
sudo ufw allow 41641/udp
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw reload
sudo service ssh restart

sudo ufw status  # Sanity Check
```

[1]: https://tailscale.com/download/linux/ubuntu-2004
[2]: https://docs.docker.com/engine/install/ubuntu/
[3]: https://docs.docker.com/engine/install/linux-postinstall/
[4]: _bin/new-machine.sh
[5]: https://tailscale.com/kb/1104/enable-ip-forwarding/
[6]: https://tailscale.com/kb/1077/secure-server-ubuntu-18-04/
