# Bootstrapping a Fresh Machine

We want to quickly make a bare metal machine or a VM from a cloud provider
look the same and be reachable from our primary jump host so we can control
it remotely.

## Bare Metal

If you have an external storage device, just copy
`_bin/bootstrap-ssh-bare-metal.sh` onto the machine. If not:

### Install All APT Packages Needed

```
sudo apt-get update
sudo apt-get install --yes \
  netcat \
  openssh-client \
  openssh-server \
  ubuntu-server
```

### Disable SSH Password Authentication and Restart SSH Server

```
sudo sed --in-place "s/.*PasswordAuthentication.*//g" /etc/ssh/sshd_config
cat <<EOF | sudo tee --append /etc/ssh/sshd_config
PasswordAuthentication no
EOF

sudo systemctl restart sshd.service
```

Restarting the SSH server should not be a problem; the assumption is that
during bootstrap you have physical access to the machine

### Receive Authorized Key(s) from a Peer on the Local Network

From the new machine, pick a `NETCAT_LISTEN_PORT` and listen for a raw TCP
request over the network

```
NETCAT_LISTEN_PORT=9107

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo "The list of all known IP addresses for this host is:"
hostname --all-ip-addresses

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo thanks | netcat -l "${NETCAT_LISTEN_PORT}" -b > "${HOME}/.extra_authorized_keys"
```

Again (as with `_bin/bootstrap-ssh-bare-metal.sh`), if you have an external
storage device, just copy the authorized SSH key(s) from that device.

While the `netcat` listener is active, send SSH public key(s) from your jump
host. For example:

```
SSH_PUBLIC_KEY_FILENAME=~/.ssh/id_ed25519.pub
LOCAL_IP=192.168.7.131
NETCAT_LISTEN_PORT=9107

cat "${SSH_PUBLIC_KEY_FILENAME}" | netcat -q 1 "${LOCAL_IP}" "${NETCAT_LISTEN_PORT}"
```

### Add Extra Authorized Key(s) to SSH.

In order to ensure the keys received over the local network are legitimate,
a prompt will be used first.

```
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
```

### Validate SSH Connection from the Jump Host

```
REMOTE_USERNAME=dhermes
LOCAL_IP=192.168.7.131

ssh "${REMOTE_USERNAME}@${LOCAL_IP}"
```

Once this connection has been confirmed, you can start to `scp` files over
to the new machine as needed and then use SSH sessions to run the scripts.
