# Clean Up

## AWS EC2 VM

Placeholder

## GCP GCE VM

Placeholder

## Bare Metal

```
$ scp _bin/k8s-node-down.sh dhermes@relaxed-bouman:~/
k8s-node-down.sh                                   100% 2010   157.1KB/s   00:00
$ scp _bin/tailscale-withdraw-linux-amd64-* dhermes@relaxed-bouman:~/
tailscale-withdraw-linux-amd64-v1.20211203.1       100% 2428KB   3.0MB/s   00:00
$
$ ssh dhermes@relaxed-bouman
dhermes@relaxed-bouman:~$ sudo mv tailscale-withdraw-linux-amd64-* /usr/local/bin/tailscale-withdraw
[sudo] password for dhermes:
dhermes@relaxed-bouman:~$ ls -1 /var/data/tailsk8s-bootstrap/
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
dhermes@relaxed-bouman:~$
dhermes@relaxed-bouman:~$ ./k8s-node-down.sh
+ '[' 0 -ne 0 ']'
++ hostname
+ CURRENT_HOSTNAME=relaxed-bouman
...
[DEBUG] >   https://api.tailscale.com/api/v2/device/12345678901234567/routes
Disabled route 10.100.3.0/24 for device 12345678901234567
+ rm --force --recursive /home/dhermes/.kube
+ sudo rm --force --recursive /etc/cni/net.d/
+ sudo rm --force --recursive /etc/kubernetes/
+ sudo rm --force --recursive /var/data/tailsk8s-bootstrap
+ sudo mkdir --parents /var/data/tailsk8s-bootstrap
+ sudo chown 1000:1000 /var/data/tailsk8s-bootstrap
dhermes@relaxed-bouman:~$
dhermes@relaxed-bouman:~$ kubectl get nodes
The connection to the server localhost:8080 was refused - did you specify the right host or port?
dhermes@relaxed-bouman:~$ ls -1 /var/data/tailsk8s-bootstrap/
dhermes@relaxed-bouman:~$ rm k8s-node-down.sh
dhermes@relaxed-bouman:~$ exit
logout
Connection to relaxed-bouman closed.
$
$
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes --output wide
NAME              STATUS   ROLES                  AGE   VERSION   INTERNAL-IP       EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
eager-jennings    Ready    control-plane,master   9h    v1.22.4   100.109.83.23     <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
nice-mcclintock   Ready    <none>                 9h    v1.22.4   100.70.213.118    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
pedantic-yonath   Ready    control-plane,master   10h   v1.22.4   100.110.217.104   <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
```
