# Add Workers

## First Worker

```
$ scp _bin/k8s-worker-join.sh dhermes@nice-mcclintock:~/
k8s-worker-join.sh                                 100% 3532   385.3KB/s   00:00
$ scp k8s-bootstrap-shared/* dhermes@nice-mcclintock:/var/data/tailsk8s-bootstrap/
ca-cert-hash.txt                                   100%   65    14.1KB/s   00:00
certificate-key.txt                                100%   65    23.8KB/s   00:00
control-plane-load-balancer.txt                    100%   15     0.9KB/s   00:00
join-token.txt                                     100%   24     8.4KB/s   00:00
kube-config.yaml                                   100% 5642   275.1KB/s   00:00
tailscale-api-key                                  100%   40    13.4KB/s   00:00
$ scp _templates/kubeadm* dhermes@nice-mcclintock:/var/data/tailsk8s-bootstrap/
kubeadm-control-plane-join-config.yaml             100%  996   353.9KB/s   00:00
kubeadm-init-config.yaml                           100% 1276   163.3KB/s   00:00
kubeadm-worker-join-config.yaml                    100%  873   206.2KB/s   00:00
$ scp _bin/tailscale-advertise-linux-amd64-* dhermes@nice-mcclintock:~/
tailscale-advertise-linux-amd64-v1.20211203.1      100% 2427KB   6.2MB/s   00:00
$
$ ssh dhermes@nice-mcclintock
dhermes@nice-mcclintock:~$ sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise
[sudo] password for dhermes:
dhermes@nice-mcclintock:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key
dhermes@nice-mcclintock:~$
dhermes@nice-mcclintock:~$ ./k8s-worker-join.sh '10.100.2.0/24'
+ '[' 1 -ne 1 ']'
+ ADVERTISE_SUBNET=10.100.2.0/24
++ hostname
+ CURRENT_HOSTNAME=nice-mcclintock
++ tailscale ip -4
+ HOST_IP=100.70.213.118
+ K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap
++ cat /var/data/tailsk8s-bootstrap/ca-cert-hash.txt
...
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.

+ rm --force /home/dhermes/kubeadm-join-config.yaml
dhermes@nice-mcclintock:~$
dhermes@nice-mcclintock:~$ kubectl get nodes --output wide
NAME              STATUS   ROLES                  AGE   VERSION   INTERNAL-IP       EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
eager-jennings    Ready    control-plane,master   10m   v1.22.4   100.109.83.23     <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
nice-mcclintock   Ready    <none>                 61s   v1.22.4   100.70.213.118    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
pedantic-yonath   Ready    control-plane,master   27m   v1.22.4   100.110.217.104   <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
dhermes@nice-mcclintock:~$ rm k8s-worker-join.sh
dhermes@nice-mcclintock:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key
dhermes@nice-mcclintock:~$ exit
logout
Connection to nice-mcclintock closed.
```

## Second Worker

```
$ scp _bin/k8s-worker-join.sh dhermes@relaxed-bouman:~/
k8s-worker-join.sh                                 100% 3532   179.6KB/s   00:00
$ scp k8s-bootstrap-shared/* dhermes@relaxed-bouman:/var/data/tailsk8s-bootstrap/
ca-cert-hash.txt                                   100%   65     2.8KB/s   00:00
certificate-key.txt                                100%   65     8.3KB/s   00:00
control-plane-load-balancer.txt                    100%   15     1.3KB/s   00:00
join-token.txt                                     100%   24     0.8KB/s   00:00
kube-config.yaml                                   100% 5642   461.5KB/s   00:00
tailscale-api-key                                  100%   40     5.3KB/s   00:00
$ scp _templates/kubeadm* dhermes@relaxed-bouman:/var/data/tailsk8s-bootstrap/
kubeadm-control-plane-join-config.yaml             100%  996    85.9KB/s   00:00
kubeadm-init-config.yaml                           100% 1276   104.5KB/s   00:00
kubeadm-worker-join-config.yaml                    100%  873   115.5KB/s   00:00
$ scp _bin/tailscale-advertise-linux-amd64-* dhermes@relaxed-bouman:~/
tailscale-advertise-linux-amd64-v1.20211203.1      100% 2427KB   2.2MB/s   00:01
$
$ ssh dhermes@relaxed-bouman
dhermes@relaxed-bouman:~$ sudo mv tailscale-advertise-linux-amd64-* /usr/local/bin/tailscale-advertise
[sudo] password for dhermes:
dhermes@relaxed-bouman:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key
dhermes@relaxed-bouman:~$
dhermes@relaxed-bouman:~$ ./k8s-worker-join.sh '10.100.3.0/24'
+ '[' 1 -ne 1 ']'
+ ADVERTISE_SUBNET=10.100.3.0/24
++ hostname
+ CURRENT_HOSTNAME=relaxed-bouman
++ tailscale ip -4
+ HOST_IP=100.122.162.98
+ K8S_BOOTSTRAP_DIR=/var/data/tailsk8s-bootstrap
++ cat /var/data/tailsk8s-bootstrap/ca-cert-hash.txt
...
* Certificate signing request was sent to apiserver and a response was received.
* The Kubelet was informed of the new secure connection details.

Run 'kubectl get nodes' on the control-plane to see this node join the cluster.

+ rm --force /home/dhermes/kubeadm-join-config.yaml
dhermes@relaxed-bouman:~$
dhermes@relaxed-bouman:~$ kubectl get nodes --output wide
NAME              STATUS   ROLES                  AGE     VERSION   INTERNAL-IP       EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
eager-jennings    Ready    control-plane,master   15m     v1.22.4   100.109.83.23     <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
nice-mcclintock   Ready    <none>                 6m41s   v1.22.4   100.70.213.118    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
pedantic-yonath   Ready    control-plane,master   32m     v1.22.4   100.110.217.104   <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
relaxed-bouman    Ready    <none>                 56s     v1.22.4   100.122.162.98    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
dhermes@relaxed-bouman:~$ rm k8s-worker-join.sh
dhermes@relaxed-bouman:~$ ls -1 /var/data/tailsk8s-bootstrap/
ca-cert-hash.txt
certificate-key.txt
control-plane-load-balancer.txt
join-token.txt
kubeadm-control-plane-join-config.yaml
kubeadm-init-config.yaml
kubeadm-worker-join-config.yaml
kube-config.yaml
tailscale-api-key
dhermes@relaxed-bouman:~$ exit
logout
Connection to relaxed-bouman closed.
```
