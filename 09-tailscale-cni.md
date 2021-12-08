# Configure CNI Networking for Tailscale

Taking our [lead][1] from Kubernetes The Hard Way, we use `kubenet` as our
CNI and just defer to an "external party" to [handle our routes][2]:

> Kubenet is a very basic, simple network plugin, on Linux only. It does not,
> of itself, implement more advanced features like cross-node networking or
> network policy. It is typically used together with a cloud provider that sets
> up routing rules for communication between nodes, or in single-node
> environments.

## Configuration

Create the `bridge` network configuration file:

```bash
cat <<EOF | sudo tee /etc/cni/net.d/10-bridge.conf
{
    "cniVersion": "0.4.0",
    "name": "tailsk8s",
    "type": "bridge",
    "bridge": "cnio0",
    "isGateway": true,
    "ipMasq": true,
    "ipam": {
        "type": "host-local",
        "ranges": [
            [
                {
                    "subnet": "${ADVERTISE_SUBNET}"
                }
            ]
        ],
        "routes": [
            {
                "dst": "0.0.0.0/0"
            }
        ]
    }
}
EOF
```

Create the `loopback` network configuration file:

```bash
cat <<EOF | sudo tee /etc/cni/net.d/99-loopback.conf
{
    "cniVersion": "0.4.0",
    "name": "lo",
    "type": "loopback"
}
EOF
```

## In the Cloud

On AWS, GCP or another cloud, this `kubenet` configuration is insufficient for
Kubernetes networking to **actually work**. Our goal here is a bare metal
cluster on Tailscale, but first let's look to how a cluster using the
`kubenet` CNI would handle routes in a public cloud.

In order to materialize the Kubernetes routes in a [GCP network][3]:

```bash
for i in 0 1 2; do
  gcloud compute routes create kubernetes-route-10-200-${i}-0-24 \
    --network kubernetes-the-hard-way \
    --next-hop-address 10.240.0.2${i} \
    --destination-range 10.200.${i}.0/24
done
```

and in an [AWS VPC][4]:

```bash
for instance in worker-0 worker-1 worker-2; do
  # ...
  aws ec2 create-route \
    --route-table-id "${ROUTE_TABLE_ID}" \
    --destination-cidr-block "${pod_cidr}" \
    --instance-id "${instance_id}"
done
```

## Tailscale

The fact that `kubenet` doesn't handle
"more advanced features like cross-node networking or network policy" is
perfectly fine. That's why we have Tailscale! In order to make this work,
we make each Kubernetes node a [subnet router][5] and accept the advertised
subnets from the other devices in the Tailnet.

<p align="center">
  <img src="./_images/tailscale-all-subnets-route-disabled.png?raw=true" />
</p>

Using the `tailscale` CLI to both advertise a subnet and to accept other
advertised subnets would look like:

```bash
tailscale up --accept-routes --advertise-routes '10.100.0.0/24'
```

However newly advertised routes must be accepted by a Tailnet admin:

<p align="center">
  <img src="./_images/tailscale-subnet-route-disabled.png?raw=true" />
</p>

<p align="center">
  <img src="./_images/tailscale-subnet-route-enabled.png?raw=true" />
</p>

Luckily the Tailscale cloud [API][6] has support for updating enabled routes.
We can combine the changes made via the local API (`tailscale up`) with
the cloud API changes:

```bash
sudo tailscale-advertise \
  --debug \
  --api-key "file:${TAILSCALE_API_KEY_FILENAME}" \
  --cidr "${ADVERTISE_SUBNET}"
```

When a node's pod subnet is advertised, Kubernetes and Tailscale will
collaborate to route packets to the right place:

```
dhermes@pedantic-yonath:~$ kubectl get pods --namespace httpbin --output wide
NAME                      READY   STATUS    RESTARTS   AGE    IP           NODE              NOMINATED NODE   READINESS GATES
httpbin-6698c4cbc-bqmpx   1/1     Running   0          5m9s   10.100.3.9   relaxed-bouman    <none>           <none>
httpbin-6698c4cbc-g4c82   1/1     Running   0          5m9s   10.100.2.2   nice-mcclintock   <none>           <none>
httpbin-6698c4cbc-pcj56   1/1     Running   0          5m9s   10.100.2.3   nice-mcclintock   <none>           <none>
dhermes@pedantic-yonath:~$ curl --max-time 5 http://10.100.2.2:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.2.2",
    "User-Agent": "curl/7.68.0"
  }
}
```

However, if we were to withdraw the routes for one of the devices in the
Tailnet:

```
dhermes@nice-mcclintock:~$ sudo tailscale-withdraw \
>   --api-key "file:${TAILSCALE_API_KEY_FILENAME}" \
>   --cidr "${ADVERTISE_SUBNET}"
Reading Tailscale API key from: /var/data/tailsk8s-bootstrap/tailscale-api-key
Inferring Tailnet from magic DNS suffix: dhermes.github.beta.tailscale.net
Using hostname: nice-mcclintock
Enabled routes for device 23563742208244416:
- 10.100.2.0/24
Disabling route 10.100.2.0/24 for device 23563742208244416...
Disabled route 10.100.2.0/24 for device 23563742208244416
```

then the other Kubernetes nodes can no longer "reach" the subnet:

```
dhermes@pedantic-yonath:~$ curl --max-time 5 http://10.100.2.2:80/headers
curl: (28) Connection timed out after 5000 milliseconds
```

## Extra Credit: Debug Mode for Advertise and Withdraw

Running `tailscale-advertise` with `--debug` gives a rough idea of the
equivalent `curl` commands corresponding to the changes made:

```
Reading Tailscale API key from: /var/data/tailsk8s-bootstrap/tailscale-api-key
[DEBUG] Calling "status without peers" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/status?peers=false
Inferring Tailnet from magic DNS suffix: dhermes.github.beta.tailscale.net
[DEBUG] Calling "get prefs" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/prefs
[DEBUG] Calling "edit prefs" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --request PATCH \
[DEBUG] >   --data-binary '{...}' \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/prefs
[DEBUG] --- /tmp/3364665985/before.json 2021-12-08 05:32:21.421986395 +0000
[DEBUG] +++ /tmp/3364665985/after.json  2021-12-08 05:32:21.421986395 +0000
[DEBUG] @@ -12,7 +12,9 @@
[DEBUG]      "AdvertiseTags": null,
[DEBUG]      "Hostname": "",
[DEBUG]      "NotepadURLs": false,
[DEBUG] -    "AdvertiseRoutes": [],
[DEBUG] +    "AdvertiseRoutes": [
[DEBUG] +        "10.100.2.0/24"
[DEBUG] +    ],
[DEBUG]      "NoSNAT": false,
[DEBUG]      "NetfilterMode": 2,
[DEBUG]      "Config": {
Using hostname: nice-mcclintock
[DEBUG] Calling "get devices in Tailnet" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   https://api.tailscale.com/api/v2/tailnet/dhermes.github/devices
[DEBUG] Matched device:
[DEBUG] > {
[DEBUG] >     "addresses": [
[DEBUG] >         "100.70.213.118",
[DEBUG] >         "fd7a:115c:a1e0:ab12:4843:cd96:6246:d576"
[DEBUG] >     ],
[DEBUG] >     "authorized": true,
[DEBUG] >     "hostname": "nice-mcclintock",
[DEBUG] >     "id": "23563742208244416",
[DEBUG] >     "name": "nice-mcclintock.dhermes.github"
[DEBUG] > }
[DEBUG] Calling "get routes" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   https://api.tailscale.com/api/v2/device/23563742208244416/routes
Advertised routes for device 23563742208244416:
- 10.100.2.0/24
Enabling route 10.100.2.0/24 for device 23563742208244416...
[DEBUG] Calling "set routes" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   --data-binary '{"routes":["10.100.2.0/24"]}'
[DEBUG] >   https://api.tailscale.com/api/v2/device/23563742208244416/routes
Enabled route 10.100.2.0/24 for device 23563742208244416
```

Similarly, running `tailscale-withdraw` with `--debug` does the same:

```
Reading Tailscale API key from: /var/data/tailsk8s-bootstrap/tailscale-api-key
[DEBUG] Calling "status without peers" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/status?peers=false
Inferring Tailnet from magic DNS suffix: dhermes.github.beta.tailscale.net
[DEBUG] Calling "get prefs" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/prefs
[DEBUG] Calling "edit prefs" local API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --request PATCH \
[DEBUG] >   --data-binary '{...}' \
[DEBUG] >   --unix-socket /var/run/tailscale/tailscaled.sock \
[DEBUG] >   http://no-op-host.invalid/localapi/v0/prefs
[DEBUG] --- /tmp/3627152075/before.json 2021-12-08 05:30:45.761018466 +0000
[DEBUG] +++ /tmp/3627152075/after.json  2021-12-08 05:30:45.761018466 +0000
[DEBUG] @@ -12,9 +12,7 @@
[DEBUG]      "AdvertiseTags": null,
[DEBUG]      "Hostname": "",
[DEBUG]      "NotepadURLs": false,
[DEBUG] -    "AdvertiseRoutes": [
[DEBUG] -        "10.100.2.0/24"
[DEBUG] -    ],
[DEBUG] +    "AdvertiseRoutes": [],
[DEBUG]      "NoSNAT": false,
[DEBUG]      "NetfilterMode": 2,
[DEBUG]      "Config": {
Using hostname: nice-mcclintock
[DEBUG] Calling "get devices in Tailnet" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   https://api.tailscale.com/api/v2/tailnet/dhermes.github/devices
[DEBUG] Matched device:
[DEBUG] > {
[DEBUG] >     "addresses": [
[DEBUG] >         "100.70.213.118",
[DEBUG] >         "fd7a:115c:a1e0:ab12:4843:cd96:6246:d576"
[DEBUG] >     ],
[DEBUG] >     "authorized": true,
[DEBUG] >     "hostname": "nice-mcclintock",
[DEBUG] >     "id": "23563742208244416",
[DEBUG] >     "name": "nice-mcclintock.dhermes.github"
[DEBUG] > }
[DEBUG] Calling "get routes" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   https://api.tailscale.com/api/v2/device/23563742208244416/routes
Enabled routes for device 23563742208244416:
- 10.100.2.0/24
Disabling route 10.100.2.0/24 for device 23563742208244416...
[DEBUG] Calling "set routes" cloud API route:
[DEBUG] > curl \
[DEBUG] >   --include \
[DEBUG] >   --user "...redacted API Key...:" \
[DEBUG] >   --data-binary '{"routes":[]}'
[DEBUG] >   https://api.tailscale.com/api/v2/device/23563742208244416/routes
Disabled route 10.100.2.0/24 for device 23563742208244416
```

[1]: https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/79a3f79b27bd28f82f071bb877a266c2e62ee506/docs/09-bootstrapping-kubernetes-workers.md#configure-cni-networking
[2]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/#kubenet
[3]: https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/79a3f79b27bd28f82f071bb877a266c2e62ee506/docs/11-pod-network-routes.md#routes
[4]: https://github.com/prabhatsharma/kubernetes-the-hard-way-aws/blob/c4872b83989562a35e9aba98ff92526a0f1498ca/docs/11-pod-network-routes.md#the-routing-table-and-routes
[5]: https://tailscale.com/kb/1019/subnets/
[6]: https://github.com/tailscale/tailscale/blob/v1.18.1/api.md
