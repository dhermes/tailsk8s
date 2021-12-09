# Add a GCP GCE Instance to the Kubernetes Cluster

In this demo `tailsk8s` cluster, we have fully virtualized the `10.100.0.0/16`
pod subnet and are letting Tailscale handle the assignment of blocks. This
means I can treat a computer under a dresser in my house and a VM in a Google
datacenter as part of the same virtualized network without any extra work other
than joining the Tailnet.

In order to be **slightly** paranoid, we'll use parts of the "20-bit block"
(`172.16.0.0/12`) in the AWS VPC to avoid any accidental collision with our
chunk of the "24-bit block" (`10.0.0.0/8`).

## Provision All GCP Resources

### VPC and Subnet

```bash
gcloud compute networks create tailsk8s --subnet-mode custom
gcloud compute networks subnets create tailsk8s \
  --network tailsk8s \
  --range '172.29.0.0/20'
```

### Firewall Rule

We need to be able to reach the new GCE instance over the public internet for
a brief period. However, once the instance has joined the Tailnet, it can be
completely unreachable from the public internet but Tailscale will still
punch our packets through! This is one of the incredible security benefits of
Tailscale, we can make the instance unreachable using AWS APIs and use
`ufw` on the instance so that the only traffic allowed is via Tailscale.

For the "first contact", we just need to add the public IP of our jump host
to a firewall rule. Once the GCE has joined the Tailnet, we can remove this
firewall rule.

```bash
JUMP_HOST_PUBLIC_IP=$(curl --silent http://checkip.amazonaws.com)
gcloud compute firewall-rules create tailsk8s-allow-external \
  --allow tcp:22 \
  --network tailsk8s \
  --source-ranges "${JUMP_HOST_PUBLIC_IP}/32"
```

### GCE Instance

Finally we have our VPC and network security in place, so the instance can
be created:

```bash
TAILSCALE_DEVICE_NAME=agitated-feistel
MACHINE_TYPE=e2-micro

gcloud compute instances create "${TAILSCALE_DEVICE_NAME}" \
  --async \
  --can-ip-forward \
  --image-family ubuntu-2004-lts \
  --image-project ubuntu-os-cloud \
  --machine-type "${MACHINE_TYPE}" \
  --scopes 'compute-rw,storage-ro,service-management,service-control,logging-write,monitoring' \
  --subnet tailsk8s \
  --tags 'tailsk8s,worker'

# To poll until completion:
gcloud compute instances list --filter="tags.items=tailsk8s"
```

> **NOTE**: Here I've chosen to use to use an `e2-micro` [instance][1], but
> this may be underpowered. Note that an `e2-standard-2` is the instance type
> [used][3] in Kelsey's Kubernetes The Hard Way. We similarly don't use the
> `--boot-disk-size 200GB` flag when invoking
> `gcloud compute instances create`.

## Validate Connection

From the jump host, use `gcloud compute ssh` to verify the firewall rule is
valid:

```bash
gcloud compute ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

Later we'll ditch `gcloud compute ssh` and just use `ssh` directly over
Tailscale.

## Join the Tailnet

From the jump host, copy over scripts, SSH `authorized_keys` and a Tailscale
one-off key for joining the Tailnet:

```bash
EXTRA_AUTHORIZED_KEYS_FILENAME=.extra_authorized_keys
TAILSCALE_AUTHKEY_FILENAME=k8s-bootstrap-shared/tailscale-one-off-key-PT

gcloud compute scp \
  "${EXTRA_AUTHORIZED_KEYS_FILENAME}" \
  "${TAILSCALE_AUTHKEY_FILENAME}" \
  _bin/bootstrap-ssh-cloud-provider.sh \
  _bin/new-machine.sh \
  ubuntu@"${TAILSCALE_DEVICE_NAME}":~/
# Once a one-off key has been used, get rid of it
rm --force "${TAILSCALE_AUTHKEY_FILENAME}"

gcloud compute ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

On the GCE instance, run the bootstrap script. This will add new SSH public
keys so we can stop using `gcloud compute ssh` (which will be necessary when
doing SSH over Tailscale):

```bash
TAILSCALE_DEVICE_NAME=agitated-feistel
EXTRA_AUTHORIZED_KEYS_FILENAME=~/.extra_authorized_keys

./bootstrap-ssh-cloud-provider.sh "${TAILSCALE_DEVICE_NAME}" "${EXTRA_AUTHORIZED_KEYS_FILENAME}"
rm --force ./bootstrap-ssh-cloud-provider.sh
```

Close the connection and ensure that the newly added SSH keys have taken
effect:

```bash
## H/T: https://cloud.google.com/compute/docs/instances/view-ip-address
PUBLIC_IP=$(gcloud compute instances describe \
  "${TAILSCALE_DEVICE_NAME}" \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

ssh ubuntu@"${PUBLIC_IP}"
```

Now, back on the GCE instance:

```bash
TAILSCALE_AUTHKEY_FILENAME=~/tailscale-one-off-key-PT

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
GCE instance is via Tailscale. To validate this, try to connect directly with
`gcloud compute ssh`, over the public IP and over Tailscale and observe which
connections work and which ones do not:

```bash
ssh -o ConnectTimeout=10 ubuntu@"${PUBLIC_IP}"
# ssh: connect to host 146.148.102.252 port 22: Connection timed out

gcloud compute ssh --ssh-flag '-o ConnectTimeout=10' ubuntu@"${TAILSCALE_DEVICE_NAME}"
# ssh: connect to host 146.148.102.252 port 22: Connection timed out
# ERROR: (gcloud.compute.ssh) [/usr/bin/ssh] exited with return code [255].

ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

To complete the process of securing the GCE instance, fully **remove** the
firewall rule that allows SSH (port 22) from `${JUMP_HOST_PUBLIC_IP}/32`:

```bash
gcloud --quiet compute firewall-rules delete tailsk8s-allow-external
```

(validate that SSH continues to work over Tailscale before moving on).

## Join the Kubernetes Cluster

For the purposes of this demo, we'll add this GCE instance as a **worker**
node. From the jump host, copy over Kubernetes install and join scripts:

```bash
scp \
  _bin/k8s-install.sh \
  _bin/k8s-worker-join.sh \
  _bin/tailscale-advertise-linux-amd64-* \
  _templates/httpbin.manifest.yaml \
  ubuntu@"${TAILSCALE_DEVICE_NAME}":~/
scp \
  k8s-bootstrap-shared/ca-cert-hash.txt \
  k8s-bootstrap-shared/control-plane-load-balancer.txt \
  k8s-bootstrap-shared/join-token.txt \
  k8s-bootstrap-shared/kube-config.yaml \
  k8s-bootstrap-shared/tailscale-api-key \
  _templates/kubeadm-worker-join-config.yaml \
  ubuntu@"${TAILSCALE_DEVICE_NAME}":/var/data/tailsk8s-bootstrap/

ssh ubuntu@"${TAILSCALE_DEVICE_NAME}"
```

In order to join, this new Kubernetes node will need to advertise a `/24`
range in the pod subnet that it is responsible for. On the GCE instance, set
up the scripts and run them:

```bash
ADVERTISE_SUBNET=10.100.5.0/24

sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise

./k8s-install.sh
rm --force ./k8s-install.sh

./k8s-worker-join.sh "${ADVERTISE_SUBNET}"
rm --force ./k8s-worker-join.sh
```

## Validate Cluster after Joining

Poke around on the GCE instance after the fact to make sure things look as
expected:

```bash
kubectl get nodes --output wide

kubectl apply --filename ./httpbin.manifest.yaml
kubectl get services --namespace httpbin --output wide
kubectl get pods --namespace httpbin --output wide
```

At this point, we should have three worker nodes and the `httpbin` deployment
uses three replicas with node anti-affinity. This means it's quite likely
one of the `httpbin` pods will land on the newly added node on the GCE
instance.

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

---

Next: [Cleaning Up][2]

[1]: https://cloud.google.com/compute/docs/general-purpose-machines#e2_machine_types
[2]: 15-cleaning-up.md
[3]: https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/79a3f79b27bd28f82f071bb877a266c2e62ee506/docs/03-compute-resources.md#kubernetes-controllers
