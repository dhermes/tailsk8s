# Initialize Cluster

## Bring Up Primary Control Plane Node

```
$ scp _bin/k8s-primary-init.sh dhermes@pedantic-yonath:~/
k8s-primary-init.sh                                100% 6060   234.5KB/s   00:00
$ scp \
>   k8s-bootstrap-shared/tailscale-api-key.txt \
>   dhermes@pedantic-yonath:/var/data/tailsk8s-bootstrap/
tailscale-api-key.txt                              100%   40     2.8KB/s   00:00
$ scp _templates/kubeadm* dhermes@pedantic-yonath:/var/data/tailsk8s-bootstrap/
kubeadm-control-plane-join-config.yaml             100% 1000    66.0KB/s   00:00
kubeadm-init-config.yaml                           100% 1280    68.2KB/s   00:00
kubeadm-worker-join-config.yaml                    100%  877    82.4KB/s   00:00
$ scp _bin/tailscale-advertise-linux-amd64-* dhermes@pedantic-yonath:~/
tailscale-advertise-linux-amd64-v1.20211203.1      100% 2427KB   3.5MB/s   00:00
$
$ ssh dhermes@pedantic-yonath
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise
dhermes@pedantic-yonath:~$ tailscale-advertise --help
Advertise to the Tailnet that the local node handles a given CIDR range
...
dhermes@pedantic-yonath:~$ ls -1 /var/data/tailsk8s-bootstrap/
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
tailscale-api-key.txt
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ ./k8s-primary-init.sh \
>   stoic-pike \
>   '10.100.0.0/16' \
>   '10.101.0.0/16' \
>   '10.100.0.0/24' \
>   '100.70.213.118'
+ '[' 5 -ne 5 ']'
+ CLUSTER_NAME=stoic-pike
...
+ rm --force /var/data/tailsk8s-bootstrap/kube-config.yaml
+ cp /home/dhermes/.kube/config /var/data/tailsk8s-bootstrap/kube-config.yaml
+ chmod 444 /var/data/tailsk8s-bootstrap/kube-config.yaml
+ rm --force /home/dhermes/kubeadm-init-config.yaml
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ kubectl get nodes
NAME              STATUS   ROLES                  AGE   VERSION
pedantic-yonath   Ready    control-plane,master   25s   v1.22.4
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ rm k8s-primary-init.sh
dhermes@pedantic-yonath:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
cluster-name.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key.txt
dhermes@pedantic-yonath:~$ exit
logout
Connection to pedantic-yonath closed.
$
$
$ scp dhermes@pedantic-yonath:/var/data/tailsk8s-bootstrap/*.txt k8s-bootstrap-shared/
ca-cert-hash.txt                                   100%   65     9.7KB/s   00:00
certificate-key.txt                                100%   65     6.9KB/s   00:00
cluster-name.txt                                   100%   11     1.3KB/s   00:00
control-plane-load-balancer.txt                    100%   15     1.9KB/s   00:00
join-token.txt                                     100%   24     3.6KB/s   00:00
k8s-bootstrap-shared//tailscale-api-key.txt: Permission denied
$ scp dhermes@pedantic-yonath:/var/data/tailsk8s-bootstrap/kube-config.yaml k8s-bootstrap-shared/
kube-config.yaml                                   100% 5642   368.1KB/s   00:00
```

## Add a Second Control Plane Node

```
$ scp _bin/k8s-control-plane-join.sh dhermes@eager-jennings:~/
k8s-control-plane-join.sh                          100% 3735   485.9KB/s   00:00
$ scp k8s-bootstrap-shared/* dhermes@eager-jennings:/var/data/tailsk8s-bootstrap/
ca-cert-hash.txt                                   100%   65     6.1KB/s   00:00
certificate-key.txt                                100%   65     8.8KB/s   00:00
cluster-name.txt                                   100%   11     1.6KB/s   00:00
control-plane-load-balancer.txt                    100%   15     3.0KB/s   00:00
join-token.txt                                     100%   24     2.9KB/s   00:00
kube-config.yaml                                   100% 5642   996.5KB/s   00:00
tailscale-api-key.txt                              100%   40     3.0KB/s   00:00
$ scp _templates/kubeadm* dhermes@eager-jennings:/var/data/tailsk8s-bootstrap/
kubeadm-control-plane-join-config.yaml             100%  996   104.9KB/s   00:00
kubeadm-init-config.yaml                           100% 1276   242.1KB/s   00:00
kubeadm-worker-join-config.yaml                    100%  873   152.1KB/s   00:00
$ scp _bin/tailscale-advertise-linux-amd64-* dhermes@eager-jennings:~/
tailscale-advertise-linux-amd64-v1.20211203.1      100% 2427KB   3.2MB/s   00:00
$
$ ssh dhermes@eager-jennings
dhermes@eager-jennings:~$ sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise
dhermes@eager-jennings:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
cluster-name.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key.txt
dhermes@eager-jennings:~$
dhermes@eager-jennings:~$ ./k8s-control-plane-join.sh '10.100.1.0/24'
+ '[' 1 -ne 1 ']'
+ ADVERTISE_SUBNET=10.100.1.0/24
++ hostname
+ CURRENT_HOSTNAME=eager-jennings
++ tailscale ip -4
+ HOST_IP=100.109.83.23
+ K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap
++ cat /var/data/tailsk8s-bootstrap/ca-cert-hash.txt
...
Run 'kubectl get nodes' to see this node join the cluster.

+ rm --force /home/dhermes/kubeadm-join-config.yaml
dhermes@eager-jennings:~$
dhermes@eager-jennings:~$ kubectl get nodes --output wide
NAME              STATUS   ROLES                  AGE     VERSION   INTERNAL-IP       EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
eager-jennings    Ready    control-plane,master   2m34s   v1.22.4   100.109.83.23     <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
pedantic-yonath   Ready    control-plane,master   19m     v1.22.4   100.110.217.104   <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
dhermes@eager-jennings:~$ rm k8s-control-plane-join.sh
dhermes@eager-jennings:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
cluster-name.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key.txt
dhermes@eager-jennings:~$ exit
logout
Connection to eager-jennings closed.
```
