# New Machine: Allow SSH in from Jump Host

Our goal is to administer all machines from the jump host, but a brand new
(bare metal) machine won't be accessible to the jump host.

The assumption here is that we have physical access to the new bare metal
machines and they are on the same local network as the jump host. We just want
to do the **bare minimum** to get SSH access from the jump host and nothing
more.

We'll use `netcat` to send the `.extra_authorized_keys` file over the network
(be sure to validate after receiving). If you have an external storage device,
this is not necessary, just copy the `bootstrap-ssh-bare-metal.sh` [script][1]
and an `.extra_authorized_keys` file onto the machine and a modified version
of the script that doesn't need to get extra authorized keys.

## Installation

Install minimal set of packages needed for SSH and our `netcat` listener:

```bash
sudo apt-get update
sudo apt-get install --yes \
  netcat \
  openssh-client \
  openssh-server
```

## Securely Configure SSH

Disable password authentication by removing all SSH config lines containing
`PasswordAuthentication` and add a `PasswordAuthentication no` line:

```bash
sudo sed --in-place "s/.*PasswordAuthentication.*//g" /etc/ssh/sshd_config
cat <<EOF | sudo tee --append /etc/ssh/sshd_config
PasswordAuthentication no
EOF

sudo systemctl restart sshd.service
```

## Receive `.extra_authorized_keys` from a Peer on the Local Network

### Listen

First list all addresses for the new machine so that the jump host knows
where to send the `.extra_authorized_keys`. Then start a `netcat` listener
to receive the authorized keys:

```bash
NETCAT_LISTEN_PORT=9107

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo "The list of all known IP addresses for this host is:"
hostname --all-ip-addresses

echo "Please send authorized keys to raw TCP listener on port ${NETCAT_LISTEN_PORT}"
echo thanks | netcat -l "${NETCAT_LISTEN_PORT}" -b > "${HOME}/.extra_authorized_keys"
```

### Send

Then from the jump host, send the extra authorized keys. For example:

```bash
EXTRA_AUTHORIZED_KEYS_FILENAME=.extra_authorized_keys
LOCAL_IP=192.168.7.131
NETCAT_LISTEN_PORT=9107

cat "${EXTRA_AUTHORIZED_KEYS_FILENAME}" | netcat -q 1 "${LOCAL_IP}" "${NETCAT_LISTEN_PORT}"
```

### Add Extra Authorized Key(s) to SSH Configuration

Back on the new machine, validate the received keys and then add them to
`~/.ssh/authorized_keys`:

```bash
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

```bash
SSH_TARGET=dhermes@192.168.7.131

ssh "${SSH_TARGET}"
```

Once this connection has been confirmed, you can start to `scp` files over
to the new machine as needed and then use SSH sessions to run the scripts.

## Extra

For machines in my cluster, I have intentionally installed [Ubuntu Desktop][3]
so I can use them as a Desktop when it's convenient. However, I primarily
want them as servers:

```bash
sudo apt-get install --yes ubuntu-server
```

Additionally, I want my servers to stay running in clamshell mode i.e. when
the laptop lid is closed (I am aware this is problematic for laptop cooling).
To do this, I [configure][4] `/etc/systemd/logind.conf` to ignore lid
switch when plugged into power:

```bash
dhermes@pedantic-yonath:~$ cat /etc/systemd/logind.conf | grep -v '^#' | grep -v '^$'
[Login]
HandleLidSwitchExternalPower=ignore
LidSwitchIgnoreInhibited=no
```

If you make this change, you'll need to reboot for it to take effect.

---

Next: [Bringing up a New Machine][2]

[1]: _bin/bootstrap-ssh-bare-metal.sh
[2]: 05-new-machine.md
[3]: https://ubuntu.com/download/desktop/thank-you?version=20.04.3&architecture=amd64
[4]: https://askubuntu.com/a/594417/439339
