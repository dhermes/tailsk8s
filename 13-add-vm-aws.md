# Add an AWS EC2 VM to the Kubernetes Cluster

In this demo `tailsk8s` cluster, we have fully virtualized the `10.100.0.0/16`
pod subnet and are letting Tailscale handle the assignment of blocks. This
means I can treat a computer under a dresser in my house and a VM in an AWS
datacenter as part of the same virtualized network without any extra work other
than joining the Tailnet.

In order to be **slightly** paranoid, we'll use parts of the "20-bit block"
(`172.16.0.0/12`) in the AWS VPC to avoid any accidental collision with our
chunk of the "24-bit block" (`10.0.0.0/8`).

## Provision All AWS Resources

In order to make cleanup easier, we'll track created IDs of all AWS resources
in the `.ec2-env` file and then `source .ec2-env` as needed to recover them:

```bash
rm --force .ec2-env  # Start fresh
```

### VPC and Subnet

```bash
VPC_ID=$(aws ec2 create-vpc \
  --cidr-block '172.29.0.0/16' \
  --output text --query 'Vpc.VpcId')
echo "VPC_ID=${VPC_ID}" >> .ec2-env
aws ec2 create-tags --resources "${VPC_ID}" --tags 'Key=Name,Value=tailsk8s'

SUBNET_ID=$(aws ec2 create-subnet \
  --vpc-id "${VPC_ID}" \
  --cidr-block '172.29.0.0/20' \
  --output text --query 'Subnet.SubnetId')
echo "SUBNET_ID=${SUBNET_ID}" >> .ec2-env
aws ec2 create-tags --resources "${SUBNET_ID}" --tags 'Key=Name,Value=tailsk8s'
```

### Internet Gateway

In order to allow external traffic into the VPC, we need an internet gateway:

```bash
INTERNET_GATEWAY_ID=$(aws ec2 create-internet-gateway \
  --output text --query 'InternetGateway.InternetGatewayId')
echo "INTERNET_GATEWAY_ID=${INTERNET_GATEWAY_ID}" >> .ec2-env
aws ec2 create-tags --resources "${INTERNET_GATEWAY_ID}" --tags 'Key=Name,Value=tailsk8s'
aws ec2 attach-internet-gateway --internet-gateway-id "${INTERNET_GATEWAY_ID}" --vpc-id "${VPC_ID}"

ROUTE_TABLE_ID=$(aws ec2 create-route-table \
  --vpc-id "${VPC_ID}" \
  --output text --query 'RouteTable.RouteTableId')
echo "ROUTE_TABLE_ID=${ROUTE_TABLE_ID}" >> .ec2-env
aws ec2 create-tags --resources "${ROUTE_TABLE_ID}" --tags 'Key=Name,Value=tailsk8s'

aws ec2 associate-route-table --route-table-id "${ROUTE_TABLE_ID}" --subnet-id "${SUBNET_ID}"
ROUTE_TABLE_ASSOCIATION_ID="$(aws ec2 describe-route-tables \
  --route-table-ids "${ROUTE_TABLE_ID}" \
  --output text --query 'RouteTables[].Associations[].RouteTableAssociationId')"
echo "ROUTE_TABLE_ASSOCIATION_ID=${ROUTE_TABLE_ASSOCIATION_ID}" >> .ec2-env

aws ec2 create-route \
  --route-table-id "${ROUTE_TABLE_ID}" \
  --destination-cidr-block '0.0.0.0/0' \
  --gateway-id "${INTERNET_GATEWAY_ID}"
```

### Security Group

We need to be able to reach the new EC2 instance over the public internet for
a brief period. However, once the instance has joined the Tailnet, it can be
completely unreachable from the public internet but Tailscale will still
punch our packets through! This is one of the incredible security benefits of
Tailscale, we can make the instance unreachable using AWS APIs and use
`ufw` on the instance so that the only traffic allowed is via Tailscale.

For the "first contact", we just need to add the public IP of our jump host
to a security group. Once the EC2 has joined the Tailnet, we can remove this
security group rule.

```bash
SECURITY_GROUP_ID=$(aws ec2 create-security-group \
  --group-name tailsk8s \
  --description 'tailsk8s security group' \
  --vpc-id "${VPC_ID}" \
  --output text --query 'GroupId')
echo "SECURITY_GROUP_ID=${SECURITY_GROUP_ID}" >> .ec2-env
aws ec2 create-tags --resources "${SECURITY_GROUP_ID}" --tags 'Key=Name,Value=tailsk8s'

JUMP_HOST_PUBLIC_IP=$(curl --silent http://checkip.amazonaws.com)
SECURITY_GROUP_RULE_ID=$(aws ec2 authorize-security-group-ingress \
  --group-id "${SECURITY_GROUP_ID}" \
  --protocol tcp \
  --port 22 \
  --cidr "${JUMP_HOST_PUBLIC_IP}/32" \
  --output text --query 'SecurityGroupRules[].SecurityGroupRuleId')
echo "SECURITY_GROUP_RULE_ID=${SECURITY_GROUP_RULE_ID}" >> .ec2-env
aws ec2 create-tags --resources "${SECURITY_GROUP_RULE_ID}" --tags 'Key=Name,Value=tailsk8s'
```

One other thing needed for our "first contact" is an SSH key stored in AWS:

```bash
aws ec2 create-key-pair \
  --key-name tailsk8s \
  --output text --query 'KeyMaterial' > tailsk8s.id_rsa
chmod 400 tailsk8s.id_rsa
ssh-keygen -f ./tailsk8s.id_rsa -y > ./tailsk8s.id_rsa.pub
```

### EC2 Instance

The goal is to use **any** Ubuntu 20.04 image, so we'll grab the latest one:

```bash
IMAGE_ID=$(aws ec2 describe-images --owners 099720109477 \
  --output json \
  --filters \
  'Name=root-device-type,Values=ebs' \
  'Name=architecture,Values=x86_64' \
  'Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*' \
  | jq -r '.Images|sort_by(.Name)[-1]|.ImageId')
echo "IMAGE_ID=${IMAGE_ID}" >> .ec2-env
```

Finally we have our VPC and network security in place, so the instance can
be created:

```bash
TAILSCALE_DEVICE_NAME=interesting-jang
INSTANCE_TYPE=t3.micro

INSTANCE_ID=$(aws ec2 run-instances \
  --associate-public-ip-address \
  --image-id "${IMAGE_ID}" \
  --count 1 \
  --key-name tailsk8s \
  --security-group-ids "${SECURITY_GROUP_ID}" \
  --instance-type "${INSTANCE_TYPE}" \
  --subnet-id "${SUBNET_ID}" \
  --block-device-mappings='{"DeviceName": "/dev/sda1", "Ebs": {"VolumeSize": 8}, "NoDevice": ""}' \
  --output text --query 'Instances[].InstanceId')
echo "INSTANCE_ID=${INSTANCE_ID}" >> .ec2-env
aws ec2 modify-instance-attribute --instance-id "${INSTANCE_ID}" --no-source-dest-check
aws ec2 create-tags --resources "${INSTANCE_ID}" --tags "Key=Name,Value=${TAILSCALE_DEVICE_NAME}"

aws ec2 wait instance-running --instance-ids "${INSTANCE_ID}"
```

> **NOTE**: There may be some issues using a `t3.micro` [instance][1] with
> `kubeadm`. For example:
>
> ```
> # [preflight] Some fatal errors occurred:
> #         [ERROR Mem]: the system RAM (943 MB) is less than the minimum 1700 MB
> ```
>
> This can be avoided by using a larger instance type (e.g. `t3.medium`) or
> by using `--ignore-preflight-errors Mem` when invoking `kubeadm join` in
> `k8s-control-plane-join.sh`.

## Validate Connection

From the jump host, retrieve the public IP of the newly created instance and
use the newly created private key `tailsk8s.id_rsa` to start an SSH session:

```bash
PUBLIC_IP=$(aws ec2 describe-instances --filters \
  "Name=tag:Name,Values=${TAILSCALE_DEVICE_NAME}" \
  'Name=instance-state-name,Values=running' \
  --output text --query 'Reservations[].Instances[].PublicIpAddress')
echo "PUBLIC_IP=${PUBLIC_IP}" >> .ec2-env

ssh -i ./tailsk8s.id_rsa ubuntu@"${PUBLIC_IP}"
```

## Join the Tailnet

From the jump host, copy over scripts, SSH `authorized_keys` and a Tailscale
one-off key for joining the Tailnet:

```bash
EXTRA_AUTHORIZED_KEYS_FILENAME=.extra_authorized_keys
TAILSCALE_AUTHKEY_FILENAME=k8s-bootstrap-shared/tailscale-one-off-key-AA

scp -i ./tailsk8s.id_rsa \
  "${EXTRA_AUTHORIZED_KEYS_FILENAME}" \
  "${TAILSCALE_AUTHKEY_FILENAME}" \
  _bin/bootstrap-ssh-cloud-provider.sh \
  _bin/new-machine.sh \
  ubuntu@"${PUBLIC_IP}":~/
# Once a one-off key has been used, get rid of it
rm --force "${TAILSCALE_AUTHKEY_FILENAME}"

ssh -i ./tailsk8s.id_rsa ubuntu@"${PUBLIC_IP}"
```

On the EC2 VM, run the bootstrap script. This will update the `hostname` to
match the desired Tailscale device name and will add new SSH public keys so
we can stop using `tailsk8s.id_rsa`:

```bash
TAILSCALE_DEVICE_NAME=interesting-jang
EXTRA_AUTHORIZED_KEYS_FILENAME=~/.extra_authorized_keys

./bootstrap-ssh-cloud-provider.sh "${TAILSCALE_DEVICE_NAME}" "${EXTRA_AUTHORIZED_KEYS_FILENAME}"
rm --force ./bootstrap-ssh-cloud-provider.sh
```

Close the connection and ensure that the newly added SSH keys have taken
effect:

```bash
ssh ubuntu@"${PUBLIC_IP}"
```

Now, back on the EC2 VM:

```bash
TAILSCALE_AUTHKEY_FILENAME=~/tailscale-one-off-key-AA

./new-machine.sh "${TAILSCALE_AUTHKEY_FILENAME}"
rm --force ./new-machine.sh
```

This script will **block** when joining the Tailnet if your Tailnet
is configured to require authorization when a new device joins (this is
a good idea). If that is the case, it can be authorized from the jump host:

```bash
./_bin/tailscale-authorize-linux-amd64-v1.20211209.1 \
  --debug \
  --hostname "${TAILSCALE_DEVICE_NAME}" \
  --api-key file:./k8s-bootstrap-shared/tailscale-api-key
# In WSL2, use `./_bin/tailscale-authorize-windows-amd64-v1.20211209.1.exe`
```

After `new-machine.sh` has completed, the **only** traffic allowed into the
EC2 is via Tailscale. To validate this, try to connect directly over the public
IP and over Tailscale and observe which one works and which does not

```bash
ssh -o ConnectTimeout=10 ubuntu@"${PUBLIC_IP}"
# ssh: connect to host 3.142.184.150 port 22: Connection timed out

ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

To complete the process of securing the EC2, fully **remove** the ingress
rule that allows SSH (port 22) from `${JUMP_HOST_PUBLIC_IP}/32`:

```bash
aws ec2 revoke-security-group-ingress \
  --group-id "${SECURITY_GROUP_ID}" \
  --security-group-rule-ids "${SECURITY_GROUP_RULE_ID}"
```

(validate that SSH continues to work over Tailscale before moving on).

## Join the Kubernetes Cluster

For the purposes of this demo, we'll add this EC2 as a **control plane** node.
From the jump host, copy over Kubernetes install and join scripts:

```bash
scp \
  _bin/k8s-install.sh \
  _bin/k8s-control-plane-join.sh \
  _bin/tailscale-advertise-linux-amd64-* \
  _templates/httpbin.manifest.yaml \
  ubuntu@"${TAILSCALE_DEVICE_NAME}":~/
scp \
  k8s-bootstrap-shared/ca-cert-hash.txt \
  k8s-bootstrap-shared/certificate-key.txt \
  k8s-bootstrap-shared/control-plane-load-balancer.txt \
  k8s-bootstrap-shared/join-token.txt \
  k8s-bootstrap-shared/kube-config.yaml \
  k8s-bootstrap-shared/tailscale-api-key \
  _templates/kubeadm-control-plane-join-config.yaml \
  ubuntu@"${TAILSCALE_DEVICE_NAME}":/var/data/tailsk8s-bootstrap/

ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

In order to join, this new Kubernetes node will need to advertise a `/24`
range in the pod subnet that it is responsible for. On the EC2, set up the
scripts and run them:

```bash
ADVERTISE_SUBNET=10.100.4.0/24

sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise

./k8s-install.sh
rm --force ./k8s-install.sh

./k8s-control-plane-join.sh "${ADVERTISE_SUBNET}"
rm --force ./k8s-control-plane-join.sh
```

## Validate Cluster after Joining

Poke around on the EC2 after the fact to make sure things look as expected:

```bash
kubectl get nodes --output wide

kubectl apply --filename ./httpbin.manifest.yaml
kubectl get services --namespace httpbin --output wide
kubectl get pods --namespace httpbin --output wide
```

Ensure the cluster networking works as expected by directly sending requests
both to the `httpbin` service and to one of the pods in the `httpbin`
deployment, for example:

```bash
SERVICE_IP=$(kubectl get service \
  --namespace httpbin httpbin \
  --output go-template='{{ .spec.clusterIP }}')
curl "http://${SERVICE_IP}:8000/headers"
# {
#   "headers": {
#     "Accept": "*/*",
#     "Host": "10.101.89.48:8000",
#     "User-Agent": "curl/7.68.0"
#   }
# }

POD_IPS=$(kubectl get pods \
  --namespace httpbin \
  --selector app=httpbin \
  --output go-template='{{ range .items }}{{ .status.podIP }} {{ end }}')
for POD_IP in ${POD_IPS}
do
  curl "http://${POD_IP}:80/headers"
done

# {
#   "headers": {
#     "Accept": "*/*",
#     "Host": "10.100.3.2",
#     "User-Agent": "curl/7.68.0"
#   }
# }
# {
#   "headers": {
#     "Accept": "*/*",
#     "Host": "10.100.2.4",
#     "User-Agent": "curl/7.68.0"
#   }
# }
# {
#   "headers": {
#     "Accept": "*/*",
#     "Host": "10.100.2.5",
#     "User-Agent": "curl/7.68.0"
#   }
# }
```

Since `httpbin` was only used for validation, tear it down:

```bash
kubectl delete --filename ./httpbin.manifest.yaml
rm --force ./httpbin.manifest.yaml
```

## Update Load Balancer

Since we have a new control plane node, the load balancer needs to be updated
to reference this node. We can do it by re-running the
`k8s-load-balancer-proxy.sh` [script][3] with three arguments instead of the
original two. As in [Provision Load Balancer][4], copy the script from the jump
host onto the load balancer host:

```bash
SSH_TARGET=dhermes@nice-mcclintock

scp _bin/k8s-load-balancer-proxy.sh "${SSH_TARGET}":~/

ssh "${SSH_TARGET}"
```

then on the load balancer host:

```bash
TAILSCALE_HOST1=eager-jennings
TAILSCALE_HOST2=pedantic-yonath
TAILSCALE_HOST3=interesting-jang

./k8s-load-balancer-proxy.sh "${TAILSCALE_HOST1}" "${TAILSCALE_HOST2}" "${TAILSCALE_HOST3}"
rm --force ./k8s-load-balancer-proxy.sh
```

---

Next: [Add a GCP GCE Instance to the Kubernetes Cluster][2]

[1]: https://aws.amazon.com/ec2/instance-types/t3/
[2]: 14-add-vm-gcp.md
[3]: _bin/k8s-load-balancer-proxy.sh
[4]: 07-provision-load-balancer.md
