# Provision Load Balancer

We'll use an HAProxy load balancer to make the Kubernetes control plane
(`kube-apiserver`) available at a single address while still allowing the
number and identity of nodes in the control plane to be **dynamic**.
See the [High Availability Considerations][2] document linked to from
[Creating Highly Available clusters with `kubeadm`][3].

This load balancer **should not** run on one of the Kubernetes nodes. However,
my cluster doesn't currently have very many machines (just four) so I put the
load balancer on one of the worker nodes.

From the jump host, copy over the `k8s-load-balancer-proxy.sh` [script][1] to
the new machine:

```bash
SSH_TARGET=dhermes@nice-mcclintock

scp _bin/k8s-load-balancer-proxy.sh "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

Then on the new machine:

```bash
TAILSCALE_HOST1=eager-jennings
TAILSCALE_HOST2=pedantic-yonath

./k8s-load-balancer-proxy.sh "${TAILSCALE_HOST1}" "${TAILSCALE_HOST2}"
rm --force ./k8s-load-balancer-proxy.sh
```

Below, let's dive into what `k8s-load-balancer-proxy.sh` does.

## Install `haproxy`

```bash
sudo apt-get update
sudo apt-get --yes upgrade
sudo apt-get install --yes haproxy
```

## Enable Non-local IPv4 Bind for HAProxy

```bash
if sudo test -f /etc/sysctl.d/haproxy.conf; then
    echo "/etc/sysctl.d/haproxy.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/haproxy.conf
fi

cat <<EOF | sudo tee /etc/sysctl.d/haproxy.conf
net.ipv4.ip_nonlocal_bind = 1
EOF

sudo sysctl --system
```

## Configure HAProxy

This uses approximately the same `haproxy.cfg` ini file from
[High Availability Considerations][2]

```bash
# Backup
sudo mv /etc/haproxy/haproxy.cfg /etc/haproxy/haproxy.cfg.backup

cat <<EOF | sudo tee /etc/haproxy/haproxy.cfg
#---------------------------------------------------------------------
# Global settings
#---------------------------------------------------------------------
global
     log /dev/log local0
     log /dev/log local1 notice
     daemon
     user haproxy
     group haproxy
...
EOF

for TAILSCALE_HOST in "${@}"
do
  TAILSCALE_IP="$(tailscale status | grep "${TAILSCALE_HOST}" | cut -f 1 -d ' ')"
  if test -z "${TAILSCALE_IP}"
  then
    echo "Could not determine Tailscale IP for host ${TAILSCALE_HOST}" >&2
    exit 1
  fi
  echo "     server ${TAILSCALE_HOST} ${TAILSCALE_IP}:6443 check fall 3 rise 2" | sudo tee --append /etc/haproxy/haproxy.cfg
done
```

One primary difference is that the arguments form an array of `TAILSCALE_HOST`,
which each get one `server ...` line at the end of `haproxy.cfg`.

For each `${TAILSCALE_HOST}` hostname, the HAProxy load balancer also needs the
Tailscale IP. To determine the `${TAILSCALE_IP}`, the
`k8s-load-balancer-proxy.sh` script uses `tailscale status`, e.g.

```
$ tailscale status | grep '\(eager-jennings\|pedantic-yonath\)'
100.109.83.23   eager-jennings       dhermes@     linux   -
100.110.217.104 pedantic-yonath      dhermes@     linux   -
```

# Ensure HAProxy is Running

Ensure the `systemd` unit is enabled and restart to reload new configuration:

```bash
sudo systemctl enable haproxy --now
sudo systemctl restart haproxy
```

---

Next: [Initialize Cluster][4]

[1]: _bin/k8s-load-balancer-proxy.sh
[2]: https://github.com/kubernetes/kubeadm/blob/e55c2a2b8e0b4e3079fd6a3586baf6472700428b/docs/ha-considerations.md#haproxy-configuration
[3]: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/high-availability/#create-load-balancer-for-kube-apiserver
[4]: 08-initialize-cluster.md
