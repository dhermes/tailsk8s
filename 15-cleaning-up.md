# Cleaning Up

## Bare Metal

When a **single** node is leaving the Kubernetes cluster, the process looks
exactly the same for workers and control plane nodes. This process is encoded
in the `k8s-node-down.sh` [script][2]. When the **last** node leaves the
cluster, i.e. the entire cluster is being torn down, there is a modified
`k8s-final-down.sh` [script][3]. In addition to Kubernetes nodes, we also
brought up a bare metal load balance, so there is a `k8s-load-balancer-down.sh`
[script][4] to turn that down as well.

### Most Nodes

From the jump host make sure the teardown scripts are present and SSH onto
the Kubernetes node to complete the task:

```bash
TAILSCALE_DEVICE_NAME=relaxed-bouman
SSH_TARGET="dhermes@${TAILSCALE_DEVICE_NAME}"

scp \
  _bin/k8s-node-down.sh \
  _bin/k8s-uninstall.sh \
  _bin/tailscale-withdraw-linux-amd64-* \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

On the Kubernetes node, set up the teardown scripts and run them:

```bash
sudo mv tailscale-withdraw-linux-amd64-* /usr/local/bin/tailscale-withdraw

# Configuration **before** teardown
ls -1 /var/data/tailsk8s-bootstrap/
kubectl get nodes

./k8s-node-down.sh
rm --force ./k8s-node-down.sh

# Configuration **before** teardown
ls -1 /var/data/tailsk8s-bootstrap/
kubectl get nodes
```

If the machine will stay in service but has no need to be part of the
Kubernetes cluster again:

```bash
./k8s-uninstall.sh
rm --force ./k8s-uninstall.sh
```

Once back on the jump host, confirm the node was removed from the cluster:

```bash
kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes --output wide
```

### **Last** Nodes

From the jump host make sure the teardown scripts are present and SSH onto
the Kubernetes node to complete the task:

```bash
TAILSCALE_DEVICE_NAME=pedantic-yonath
SSH_TARGET="dhermes@${TAILSCALE_DEVICE_NAME}"

scp \
  _bin/k8s-final-down.sh \
  _bin/tailscale-withdraw-linux-amd64-* \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

On the Kubernetes node, set up the teardown scripts and run them:

```bash
sudo mv tailscale-withdraw-linux-amd64-* /usr/local/bin/tailscale-withdraw
./k8s-final-down.sh
rm --force ./k8s-final-down.sh
```

### Load Balancer

From the jump host make sure the teardown scripts are present and SSH onto
the Kubernetes node to complete the task:

```bash
TAILSCALE_DEVICE_NAME=nice-mcclintock
SSH_TARGET="dhermes@${TAILSCALE_DEVICE_NAME}"

scp \
  _bin/k8s-load-balancer-down.sh \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

On the load balancer machine, set up the teardown scripts and run them:

```bash
./k8s-load-balancer-down.sh
rm --force ./k8s-load-balancer-down.sh
```

## AWS EC2 VM

Since the VM in this demo is intended to be ephemeral, we can just destroy it
from the jump host. However, before doing it, we need to gracefully leave the
cluster and withdraw the pod CIDR routes from the Tailnet. From the jump host
make sure the teardown scripts are present and SSH onto the EC2 VM to complete
the task:

```bash
TAILSCALE_DEVICE_NAME=interesting-jang
SSH_TARGET="ubuntu@${TAILSCALE_DEVICE_NAME}"

scp \
  _bin/k8s-node-down.sh \
  _bin/tailscale-withdraw-linux-amd64-* \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

On the Kubernetes node, set up the teardown scripts and run them:

```bash
sudo mv tailscale-withdraw-linux-amd64-* /usr/local/bin/tailscale-withdraw
./k8s-node-down.sh
```

Back on the jump host, confirm the node was removed from the cluster:

```bash
kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes --output wide
```

Finally, tear down the EC2 instance and all other resources:

```bash
source .ec2-env

aws ec2 terminate-instances --instance-ids "${INSTANCE_ID}"
aws ec2 wait instance-terminated --instance-ids "${INSTANCE_ID}"

aws ec2 delete-key-pair --key-name tailsk8s
rm --force ./tailsk8s.id_rsa ./tailsk8s.id_rsa.pub

aws ec2 delete-security-group --group-id "${SECURITY_GROUP_ID}"
ROUTE_TABLE_ASSOCIATION_ID="$(aws ec2 describe-route-tables \
  --route-table-ids "${ROUTE_TABLE_ID}" \
  --output text --query 'RouteTables[].Associations[].RouteTableAssociationId')"
aws ec2 disassociate-route-table --association-id "${ROUTE_TABLE_ASSOCIATION_ID}"
aws ec2 delete-route-table --route-table-id "${ROUTE_TABLE_ID}"

aws ec2 detach-internet-gateway \
  --internet-gateway-id "${INTERNET_GATEWAY_ID}" \
  --vpc-id "${VPC_ID}"
aws ec2 delete-internet-gateway --internet-gateway-id "${INTERNET_GATEWAY_ID}"

aws ec2 delete-subnet --subnet-id "${SUBNET_ID}"

aws ec2 delete-vpc --vpc-id "${VPC_ID}"
```

After doing this, manually **remove** the `${TAILSCALE_DEVICE_NAME}` from
the Tailnet in the Tailscale UI. (It's probably worth adding a
`tailscale-remove` command to this project.)

## GCP GCE Instance

Since the instance in this demo is intended to be ephemeral, we can just
destroy it from the jump host. However, before doing it, we need to gracefully
leave the cluster and withdraw the pod CIDR routes from the Tailnet. From the
jump host make sure the teardown scripts are present and SSH onto the GCE VM to
complete the task:

```bash
TAILSCALE_DEVICE_NAME=agitated-feistel
SSH_TARGET="ubuntu@${TAILSCALE_DEVICE_NAME}"

scp \
  _bin/k8s-node-down.sh \
  _bin/tailscale-withdraw-linux-amd64-* \
  "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

On the Kubernetes node, set up the teardown scripts and run them:

```bash
sudo mv tailscale-withdraw-linux-amd64-* /usr/local/bin/tailscale-withdraw
./k8s-node-down.sh
```

Back on the jump host, confirm the node was removed from the cluster:

```bash
kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes --output wide
```

Finally, tear down the GCE instance and all other resources:

```bash
gcloud --quiet compute instances delete \
  "${TAILSCALE_DEVICE_NAME}" \
  --zone "$(gcloud config get-value compute/zone)"
# Attempt to delete `tailsk8s-allow-external`; it should already have been
# deleted after the GCE VM joined the Tailnet
gcloud --quiet compute firewall-rules delete tailsk8s-allow-external
gcloud --quiet compute networks subnets delete tailsk8s
gcloud --quiet compute networks delete tailsk8s
```

After doing this, manually **remove** the `${TAILSCALE_DEVICE_NAME}` from
the Tailnet in the Tailscale UI. (It's probably worth adding a
`tailscale-remove` command to this project.)

[1]: https://github.com/prabhatsharma/kubernetes-the-hard-way-aws/blob/c4872b83989562a35e9aba98ff92526a0f1498ca/docs/14-cleanup.md
[2]: _bin/k8s-node-down.sh
[3]: _bin/k8s-final-down.sh
[4]: _bin/k8s-load-balancer-down.sh
