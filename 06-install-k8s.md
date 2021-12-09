# Installing Kubernetes Tools

In order to install and run Kubernetes, we need some tools (e.g. `kubeadm`,
`kubectl` and `kubelet`) and we need to add or modify some configurations (e.g.
setting `systemd` as the Docker `cgroup` driver). See [Installing kubeadm][1]
for more details.

From the jump host, copy over the `k8s-install.sh` [script][2] to the new
machine:

```bash
SSH_TARGET=dhermes@pedantic-yonath

scp _bin/k8s-install.sh "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

Then on the new machine:

```bash
./k8s-install.sh
rm --force ./k8s-install.sh
```

Below, let's dive into what `k8s-install.sh` does.

## Use `systemd` as cgroup driver for Docker

```bash
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
sudo systemctl enable docker
sudo systemctl daemon-reload
sudo systemctl restart docker
```

See [Container runtimes][3] for more details.

## Let `iptables` See Bridged Traffic

```bash
if sudo test -f /etc/modules-load.d/k8s.conf; then
    echo "/etc/modules-load.d/k8s.conf exists, will be overwritten."
    sudo rm --force /etc/modules-load.d/k8s.conf
fi
if sudo test -f /etc/sysctl.d/k8s.conf; then
    echo "/etc/sysctl.d/k8s.conf exists, will be overwritten."
    sudo rm --force /etc/sysctl.d/k8s.conf
fi

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

sudo modprobe br_netfilter
sudo sysctl --system
```

## Install `kubeadm`, `kubelet` and `kubectl`

Unlike previous install steps, we explicitly **pin** the version of these
binaries (as opposed to using the `apt.kubernetes.io` package repository).

```bash
ARCH=amd64
K8S_VERSION=v1.22.4
K8S_BIN_DIR=/usr/local/bin

cd "${K8S_BIN_DIR}"
sudo curl --location --remote-name-all \
  "https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/{kubeadm,kubelet,kubectl}"
sudo chmod +x {kubeadm,kubelet,kubectl}
```

## Install standard CNI plugins

```bash
ARCH=amd64
CNI_VERSION=v0.8.2

sudo rm --force --recursive /opt/cni/bin
sudo mkdir --parents /opt/cni/bin
curl --location \
  "https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" \
  | sudo tar --directory /opt/cni/bin --extract --gzip
```

## Install `crictl`

```bash
ARCH=amd64
CRICTL_VERSION=v1.22.0
K8S_BIN_DIR=/usr/local/bin

sudo mkdir --parents "${K8S_BIN_DIR}"
curl --location \
  "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" \
  | sudo tar --directory "${K8S_BIN_DIR}" --extract --gzip
```

## Configure systemd to run `kubelet` unit and support `kubeadm`

```bash
K8S_RELEASE_VERSION=v0.4.0
K8S_BIN_DIR=/usr/local/bin

curl --silent --show-error --location \
  "https://raw.githubusercontent.com/kubernetes/release/${K8S_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" \
  | sed "s:/usr/bin:${K8S_BIN_DIR}:g" \
  | sudo tee /etc/systemd/system/kubelet.service
sudo mkdir --parents /etc/systemd/system/kubelet.service.d
curl --silent --show-error --location \
  "https://raw.githubusercontent.com/kubernetes/release/${K8S_RELEASE_VERSION}/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" \
  | sed "s:/usr/bin:${K8S_BIN_DIR}:g" \
  | sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
```

## Disable Swap

Backup the file systems table `/etc/fstab` first if you'd like to roll back
changes later.

```bash
sudo swapoff --all
sudo sed -i '/ swap / s/^/#/' /etc/fstab
```

After doing this, sanity check the changes:

```bash
free --human
```

## Pre-fetch All Images Used by `kubeadm`

```bash
kubeadm config images pull
```

---

Next: [Provision Load Balancer][4]

[1]: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
[2]: _bin/k8s-install.sh
[3]: https://kubernetes.io/docs/setup/production-environment/container-runtimes/
[4]: 07-provision-load-balancer.md
