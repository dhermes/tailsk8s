# Adding a New Control Plane Node

Once the cluster has been initialized, adding **another** control plane node
follows roughly the same process (but has less work to do). The
`k8s-control-plane-join.sh` [script][2] requires **only** one argument:

- `ADVERTISE_SUBNET`: The subnet used for pods on this node.

However, this time the script assumes there are many files present (we'll copy
these from the jump host):

- `/var/data/tailsk8s-bootstrap/ca-cert-hash.txt`
- `/var/data/tailsk8s-bootstrap/certificate-key.txt`
- `/var/data/tailsk8s-bootstrap/control-plane-load-balancer.txt`
- `/var/data/tailsk8s-bootstrap/join-token.txt`
- `/var/data/tailsk8s-bootstrap/kubeadm-control-plane-join-config.yaml`
- `/var/data/tailsk8s-bootstrap/kube-config.yaml`
- `/var/data/tailsk8s-bootstrap/tailscale-api-key`
- `/usr/local/bin/tailscale-advertise`

To actually join the cluster, copy over the inputs from the jump host,
and then run the script on the control plane node:

```bash
SSH_TARGET=dhermes@eager-jennings

scp \
  _bin/k8s-control-plane-join.sh \
  _bin/tailscale-advertise-linux-amd64-* \
  "${SSH_TARGET}":~/
scp \
  k8s-bootstrap-shared/ca-cert-hash.txt \
  k8s-bootstrap-shared/certificate-key.txt \
  k8s-bootstrap-shared/control-plane-load-balancer.txt \
  k8s-bootstrap-shared/join-token.txt \
  k8s-bootstrap-shared/kube-config.yaml \
  k8s-bootstrap-shared/tailscale-api-key \
  _templates/kubeadm-control-plane-join-config.yaml \
  "${SSH_TARGET}":/var/data/tailsk8s-bootstrap

ssh "${SSH_TARGET}"
```

On the new control plane node (e.g. on `eager-jennings`):

```bash
ADVERTISE_SUBNET=10.100.1.0/24

sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise

./k8s-control-plane-join.sh "${ADVERTISE_SUBNET}"
rm --force ./k8s-control-plane-join.sh
```

Having to **manually** manage the `${ADVERTISE_SUBNET}` ranges (and making
sure we don't collide with previous nodes) is not something you'd want in
a production grade CNI. However, for our purposes (exploration) it is perfectly
fine. A polished Kubernetes `tailsk8s` CNI could use a DaemonSet to handle
updates to advertised routes and to interact with a database (e.g. `etcd`)
to handle subnet allocation.

Below, let's dive into what `k8s-control-plane-join.sh` does.

## Kubernetes Cluster Bootstrap

```bash
K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap

sudo rm --force --recursive /etc/kubernetes/

rm --force --recursive "${HOME}/.kube"
mkdir --parents "${HOME}/.kube"
cp "${K8S_BOOTSTRAP_DIR}/kube-config.yaml" "${HOME}/.kube/config"
```

## CNI via `kubenet`

See [Configure CNI Networking for Tailscale][4].

## Join the Cluster with `kubeadm`

We use the `kubeadm-control-plane-join-config.yaml` [template][3] to fully
specify the join configuration as YAML (vs. via flags).

```bash
CONFIG_TEMPLATE_FILENAME=/var/data/tailsk8s-bootstrap/kubeadm-control-plane-join-config.yaml
CA_CERT_HASH="sha256:$(cat "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt")"
CERTIFICATE_KEY="$(cat "${K8S_BOOTSTRAP_DIR}/certificate-key.txt")"
JOIN_TOKEN="$(cat "${K8S_BOOTSTRAP_DIR}/join-token.txt")"
CONTROL_PLANE_LOAD_BALANCER="$(cat "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt")"

echo "Populating \`kubeadm\` configuration via template:"
echo '================================================'
cat "${CONFIG_TEMPLATE_FILENAME}"

CA_CERT_HASH="${CA_CERT_HASH}" \
  CERTIFICATE_KEY="${CERTIFICATE_KEY}" \
  JOIN_TOKEN="${JOIN_TOKEN}" \
  CONTROL_PLANE_LOAD_BALANCER="${CONTROL_PLANE_LOAD_BALANCER}" \
  NODE_NAME="$(hostname)" \
  HOST_IP="$(tailscale ip -4)" \
  envsubst \
  < "${CONFIG_TEMPLATE_FILENAME}" \
  > "${HOME}/kubeadm-join-config.yaml"

sudo kubeadm join \
  --config "${HOME}/kubeadm-join-config.yaml"
rm --force "${HOME}/kubeadm-join-config.yaml"
```

The template contains information about the current node and specifies
known configuration values (address, join token, etc.) about the existing
cluster:

```yaml
apiVersion: kubeadm.k8s.io/v1beta3
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: ${CONTROL_PLANE_LOAD_BALANCER}:6443
    token: ${JOIN_TOKEN}
    caCertHashes:
      - ${CA_CERT_HASH}
nodeRegistration:
  name: ${NODE_NAME}
  kubeletExtraArgs:
    node-ip: ${HOST_IP}
controlPlane:
  localAPIEndpoint:
    advertiseAddress: ${HOST_IP}
    bindPort: 6443
  certificateKey: ${CERTIFICATE_KEY}
```

## Label the Newly Added Node with `tailsk8s` Label(s)

```bash
ADVERTISE_SUBNET="..."

kubectl label node \
  "$(hostname)" \
  "tailsk8s.io/advertise-subnet=${ADVERTISE_SUBNET/\//__}"
```

## Verify

From the jump host, verify the node was added:

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes
NAME              STATUS   ROLES                  AGE     VERSION
eager-jennings    Ready    control-plane,master   4m37s   v1.22.4
pedantic-yonath   Ready    control-plane,master   18m10s  v1.22.4
```

## High Availability

We're using a [stacked `etcd` topology][5] here, which means each control plane
node is also an `etcd` node as well. Since I only have four machines, I am
using two as control plane nodes and two as worker nodes. It's worth noting
that two `etcd` nodes is in some sense **worse** than one `etcd` node, because
they'll never be able to form [quorum][6] when they disagree.

---

Next: [Adding a Worker Node][1]

[1]: 11-add-worker-node.md
[2]: _bin/k8s-control-plane-join.sh
[3]: _templates/kubeadm-control-plane-join-config.yaml
[4]: 09-tailscale-cni.md
[5]: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology
[6]: https://etcd.io/docs/v3.3/faq/#why-an-odd-number-of-cluster-members
