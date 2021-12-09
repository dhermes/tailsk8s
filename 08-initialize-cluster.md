# Initialize Cluster

Now all of the preliminaries are out of the way, it's time to bring up the
cluster! Adding our **first** control plane node is different than adding
any of the other nodes because we are configuring the cluster itself
(just once). The `k8s-primary-init.sh` [script][2] handles this task and
consumes 5 input arguments:

- `CLUSTER_NAME`: The human readable name of the cluster.
- `POD_SUBNET`: The subnet used to allocate (virtual) IPs to pods in
  the cluster.
- `SERVICE_SUBNET`: The subnet used to allocate (virtual) IPs to services in
  the cluster.
- `ADVERTISE_SUBNET`: The subnet used for pods on this node; this will be
  a part of `${POD_SUBNET}` exclusively owned by this node.
- `CONTROL_PLANE_LOAD_BALANCER`: The Tailscale IP of the load balancer we
  created in [Provision Load Balancer][3].

The script assumes there are some files present:

- `/var/data/tailsk8s-bootstrap/tailscale-api-key`
- `/var/data/tailsk8s-bootstrap/kubeadm-init-config.yaml`
- `/usr/local/bin/tailscale-advertise` (or anywhere else on `${PATH}`)

and it will produce re-usable **output** files in the
`/var/data/tailsk8s-bootstrap` directory that are intended to be copied back
onto the jump host:

- `join-token.txt`
- `certificate-key.txt`
- `control-plane-load-balancer.txt`
- `ca-cert-hash.txt`
- `kube-config.yaml`

To actually bring up the cluster, copy over the inputs from the jump host,
run the script on the control plane node and then copy back the outputs to
the jump host:

```bash
SSH_TARGET=dhermes@pedantic-yonath

scp \
  _bin/k8s-primary-init.sh \
  _bin/tailscale-advertise-linux-amd64-* \
  "${SSH_TARGET}":~/
scp \
  k8s-bootstrap-shared/tailscale-api-key \
  _templates/kubeadm-init-config.yaml \
  "${SSH_TARGET}":/var/data/tailsk8s-bootstrap

ssh "${SSH_TARGET}"
# ... run

REMOTE_FILES="ca-cert-hash.txt,certificate-key.txt,control-plane-load-balancer.txt,join-token.txt,kube-config.yaml"
scp \
  "${SSH_TARGET}":/var/data/tailsk8s-bootstrap/\{"${REMOTE_FILES}"\} \
  k8s-bootstrap-shared/
```

To actually carry out that `# ... run` step on the control plane node
(e.g. on `pedantic-yonath`):

```bash
CLUSTER_NAME=stoic-pike
POD_SUBNET=10.100.0.0/16
SERVICE_SUBNET=10.101.0.0/16
ADVERTISE_SUBNET=10.100.0.0/24
# NOTE: The `CONTROL_PLANE_LOAD_BALANCER` IP can be found with `tailscale status`
CONTROL_PLANE_LOAD_BALANCER=100.70.213.118

sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise

./k8s-primary-init.sh \
  "${CLUSTER_NAME}" \
  "${POD_SUBNET}" \
  "${SERVICE_SUBNET}" \
  "${ADVERTISE_SUBNET}" \
  "${CONTROL_PLANE_LOAD_BALANCER}"

rm --force ./k8s-primary-init.sh
```

Note that we take care to avoid colliding with the `100.x.y.z` CGNAT
address space [used by Tailscale][8].

Below, let's dive into what `k8s-primary-init.sh` does.

## Kubernetes Cluster Bootstrap (Before)

In order for **new** nodes to join the `kubeadm`-managed cluster, they'll
need to present a join token. Additionally, new control plane nodes will need
to access the private keys for the Kubernetes CA, the `etcd` CA and a few
other CAs. We'll utilize the `certificateKey` configuration value to pass
(encrypted) private keys from one node to another.

```bash
CONTROL_PLANE_LOAD_BALANCER="..."
K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap

rm --force "${K8S_BOOTSTRAP_DIR}/join-token.txt"
kubeadm token generate > "${K8S_BOOTSTRAP_DIR}/join-token.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/join-token.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
kubeadm certs certificate-key > "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"
chmod 400 "${K8S_BOOTSTRAP_DIR}/certificate-key.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"
echo "${CONTROL_PLANE_LOAD_BALANCER}" > "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"
chmod 444 "${K8S_BOOTSTRAP_DIR}/control-plane-load-balancer.txt"
```

## CNI via `kubenet`

This is a core part of bringing up the node (and cluster), but it is a very
involved topic. We'll save this for the next section:
[Configure CNI Networking for Tailscale][1].

## Initialize the Cluster with `kubeadm`

We use the `kubeadm-init-config.yaml` [template][4] to fully specify the
cluster configuration as YAML (vs. via flags). See the
[`kubeadm` Configuration (v1beta3)][6] documentation as well as the
[underlying Go types][5].

```bash
CONFIG_TEMPLATE_FILENAME=/var/data/tailsk8s-bootstrap/kubeadm-init-config.yaml
CERTIFICATE_KEY="$(cat "${K8S_BOOTSTRAP_DIR}/certificate-key.txt")"
JOIN_TOKEN="$(cat "${K8S_BOOTSTRAP_DIR}/join-token.txt")"
CLUSTER_NAME="..."
POD_SUBNET="..."
SERVICE_SUBNET="..."

echo "Populating \`kubeadm\` configuration via template:"
echo '================================================'
cat "${CONFIG_TEMPLATE_FILENAME}"

CERTIFICATE_KEY="${CERTIFICATE_KEY}" \
  JOIN_TOKEN="${JOIN_TOKEN}" \
  CLUSTER_NAME="${CLUSTER_NAME}" \
  POD_SUBNET="${POD_SUBNET}" \
  SERVICE_SUBNET="${SERVICE_SUBNET}" \
  CONTROL_PLANE_LOAD_BALANCER="${CONTROL_PLANE_LOAD_BALANCER}" \
  NODE_NAME="$(hostname)" \
  HOST_IP="$(tailscale ip -4)" \
  envsubst \
  < "${CONFIG_TEMPLATE_FILENAME}" \
  > "${HOME}/kubeadm-init-config.yaml"

sudo rm --force --recursive /etc/kubernetes/
sudo rm --force --recursive /var/lib/etcd/
sudo rm --force --recursive /var/lib/kubelet/
sudo kubeadm init \
  --config "${HOME}/kubeadm-init-config.yaml" \
  --upload-certs \
  --skip-token-print \
  --skip-certificate-key-print
rm --force "${HOME}/kubeadm-init-config.yaml"
```

The template contains information about the current node and contains some
of the secure tokens we generated above to allow other nodes to join the
cluster:

```yaml
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
bootstrapTokens:
  - token: ${JOIN_TOKEN}
    description: kubeadm bootstrap token
localAPIEndpoint:
  advertiseAddress: ${HOST_IP}
  bindPort: 6443
certificateKey: ${CERTIFICATE_KEY}
nodeRegistration:
  name: ${NODE_NAME}
  kubeletExtraArgs:
    node-ip: ${HOST_IP}
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
kubernetesVersion: v1.22.4
clusterName: ${CLUSTER_NAME}
controlPlaneEndpoint: ${CONTROL_PLANE_LOAD_BALANCER}:6443
networking:
  dnsDomain: cluster.local
  podSubnet: ${POD_SUBNET}
  serviceSubnet: ${SERVICE_SUBNET}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
cgroupDriver: systemd
```

Note the usage of the `--node-ip` extra `kubelet` argument. This is to ensure
the Tailscale IP is used vs. a local IP (e.g. `192.168.7.131`). See
[`kubeadm init/join` and ExternalIP vs InternalIP][9] for more details.

## Set Up Kubernetes Config

```bash
rm --force --recursive "${HOME}/.kube"
mkdir --parents "${HOME}/.kube"

sudo cp /etc/kubernetes/admin.conf "${HOME}/.kube/config"
sudo chown "$(id --user):$(id --group)" "${HOME}/.kube/config"
```

## Label the Newly Added Node with `tailsk8s` Label(s)

```bash
ADVERTISE_SUBNET="..."

kubectl label node \
  "$(hostname)" \
  "tailsk8s.io/advertise-subnet=${ADVERTISE_SUBNET/\//__}"
```

## Kubernetes Cluster Bootstrap (After)

See [Token-based discovery with CA pinning][7].

```bash
rm --force "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt \
  | openssl rsa -pubin -outform der 2>/dev/null \
  | openssl dgst -sha256 -hex \
  | sed 's/^.* //' \
  > "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"
chmod 444 "${K8S_BOOTSTRAP_DIR}/ca-cert-hash.txt"

rm --force "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
cp "${HOME}/.kube/config" "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
chmod 444 "${K8S_BOOTSTRAP_DIR}/kube-config.yaml"
```

## Verify

On the newly added control plane node:

```
$ kubectl get nodes
NAME              STATUS   ROLES                  AGE     VERSION
pedantic-yonath   Ready    control-plane,master   5m47s   v1.22.4
```

Similarly, once `kube-config.yaml` has been copied onto the jump host, we
can also verify that the jump host can query the Kubernetes API via the
load balancer:

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes
NAME              STATUS   ROLES                  AGE     VERSION
pedantic-yonath   Ready    control-plane,master   6m22s   v1.22.4
```

---

Next: [Configure CNI Networking for Tailscale][1]

[1]: 09-tailscale-cni.md
[2]: _bin/k8s-primary-init.sh
[3]: 07-provision-load-balancer.md
[4]: _templates/kubeadm-init-config.yaml
[5]: https://github.com/kubernetes/kubernetes/blob/v1.22.4/cmd/kubeadm/app/apis/kubeadm/types.go
[6]: https://kubernetes.io/docs/reference/config-api/kubeadm-config.v1beta3/
[7]: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-join/#token-based-discovery-with-ca-pinning
[8]: https://tailscale.com/kb/1015/100.x-addresses/
[9]: https://medium.com/@aleverycity/kubeadm-init-join-and-externalip-vs-internalip-519519ddff89
