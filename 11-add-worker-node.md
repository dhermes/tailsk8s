# Adding a Worker Node

Adding a worker node looks roughly the same as added a control plane node.
However, a worker doesn't need to **sign** certificates, so doesn't need access
to the cluster CA private keys. As with control plane nodes, the
`k8s-worker-join.sh` [script][2] requires **only** one argument:

- `ADVERTISE_SUBNET`: The subnet used for pods on this node.

and the script assumes the presence of roughly the same set of files:

- `/var/data/tailsk8s-bootstrap/ca-cert-hash.txt`
- `/var/data/tailsk8s-bootstrap/control-plane-load-balancer.txt`
- `/var/data/tailsk8s-bootstrap/join-token.txt`
- `/var/data/tailsk8s-bootstrap/kubeadm-worker-join-config.yaml`
- `/var/data/tailsk8s-bootstrap/kube-config.yaml`
- `/var/data/tailsk8s-bootstrap/tailscale-api-key`
- `/usr/local/bin/tailscale-advertise`

To actually join the cluster, copy over the inputs from the jump host,
run the script on the worker node:

```bash
SSH_TARGET=dhermes@nice-mcclintock

scp \
  _bin/k8s-worker-join.sh \
  _bin/tailscale-advertise-linux-amd64-* \
  "${SSH_TARGET}":~/
scp \
  k8s-bootstrap-shared/ca-cert-hash.txt \
  k8s-bootstrap-shared/control-plane-load-balancer.txt \
  k8s-bootstrap-shared/join-token.txt \
  k8s-bootstrap-shared/kube-config.yaml \
  k8s-bootstrap-shared/tailscale-api-key \
  _templates/kubeadm-worker-join-config.yaml \
  "${SSH_TARGET}":/var/data/tailsk8s-bootstrap

ssh "${SSH_TARGET}"
```

On the new worker node (e.g. on `nice-mcclintock`):

```bash
ADVERTISE_SUBNET=10.100.2.0/24

sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise

./k8s-worker-join.sh "${ADVERTISE_SUBNET}"
rm --force ./k8s-worker-join.sh
```

Below, let's dive into what `k8s-worker-join.sh` does.

## Kubernetes Cluster Bootstrap

```bash
K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap

sudo rm --force --recursive /etc/kubernetes/

rm --force --recursive "${HOME}/.kube"
mkdir --parents "${HOME}/.kube"
cp "${K8S_BOOTSTRAP_DIR}/kube-config.yaml" "${HOME}/.kube/config"
```

## CNI via `kubenet`

See [Configure CNI Networking for Tailscale][3].

## Join the Cluster with `kubeadm`

We use the `kubeadm-worker-join-config.yaml` [template][4] to fully
specify the join configuration as YAML (vs. via flags).

```bash
CONFIG_TEMPLATE_FILENAME=/var/data/tailsk8s-bootstrap/kubeadm-worker-join-config.yaml
CA_CERT_HASH="sha256:$(cat "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt")"
JOIN_TOKEN="$(cat "${K8S_BOOTSTRAP_DIR}/join-token.txt")"
CONTROL_PLANE_LOAD_BALANCER="$(cat "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt")"

echo "Populating \`kubeadm\` configuration via template:"
echo '================================================'
cat "${CONFIG_TEMPLATE_FILENAME}"

CA_CERT_HASH="${CA_CERT_HASH}" \
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

The template is identical to `kubeadm-control-plane-join-config.yaml` except
for the presence of the `controlPlane` top-level key:

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
eager-jennings    Ready    control-plane,master   5m42s   v1.22.4
nice-mcclintock   Ready    <none>                 1m5s    v1.22.4
pedantic-yonath   Ready    control-plane,master   19m15s  v1.22.4
```

Similarly, after adding the fourth bare metal node verify:

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes
NAME              STATUS   ROLES                  AGE     VERSION
eager-jennings    Ready    control-plane,master   7m55s   v1.22.4
nice-mcclintock   Ready    <none>                 3m18s   v1.22.4
pedantic-yonath   Ready    control-plane,master   21m28s  v1.22.4
relaxed-bouman    Ready    <none>                 2m13s   v1.22.4
```

## High Availability

We're using a [stacked `etcd` topology][5] here, which means each control plane
node is also an `etcd` node as well. Since I only have four machines, I am
using two as control plane nodes and two as worker nodes. It's worth noting
that two `etcd` nodes is in some sense **worse** than one `etcd` node, because
they'll never be able to form [quorum][6] when they disagree.

---

Next: [Smoke Test][1]

[1]: 12-smoke-test.md
[2]: _bin/k8s-worker-join.sh
[3]: 09-tailscale-cni.md
[4]: _templates/kubeadm-worker-join-config.yaml
[5]: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology
[6]: https://etcd.io/docs/v3.3/faq/#why-an-odd-number-of-cluster-members
