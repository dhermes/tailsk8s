# Server Dependencies

## Join Tailnet

```
$ scp _bin/new-machine.sh dhermes@pedantic-yonath:~/
new-machine.sh                                     100% 5649   409.6KB/s   00:00
```

## Install Kubernetes Dependencies

```
$ scp _bin/k8s-install.sh dhermes@pedantic-yonath:~/
k8s-install.sh                                     100% 3784    72.6KB/s   00:00
$ ssh dhermes@pedantic-yonath
dhermes@pedantic-yonath:~$ ./k8s-install.sh
+ '[' 0 -ne 0 ']'
+ ARCH=amd64
+ CNI_VERSION=v0.8.2
+ CRICTL_VERSION=v1.22.0
+ K8S_VERSION=v1.22.4
+ K8S_RELEASE_VERSION=v0.4.0
+ K8S_DOWNLOAD_DIR=/usr/local/bin
+ cat
+ sudo tee /etc/docker/daemon.json
[sudo] password for dhermes:
...
[config/images] Pulled k8s.gcr.io/pause:3.5
[config/images] Pulled k8s.gcr.io/etcd:3.5.0-0
[config/images] Pulled k8s.gcr.io/coredns/coredns:v1.8.4
dhermes@pedantic-yonath:~$
dhermes@pedantic-yonath:~$ rm k8s-install.sh
dhermes@pedantic-yonath:~$ exit
logout
Connection to pedantic-yonath closed.
```
