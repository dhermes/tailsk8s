# Provision Load Balancer

```
$ scp _bin/k8s-load-balancer-proxy.sh dhermes@nice-mcclintock:~/
k8s-load-balancer-proxy.sh                         100% 3968   737.1KB/s   00:00
$ ssh dhermes@nice-mcclintock
dhermes@nice-mcclintock:~$
dhermes@nice-mcclintock:~$ ./k8s-load-balancer-proxy.sh \
>   'pedantic-yonath 100.110.217.104' \
>   'eager-jennings 100.109.83.23'
+ '[' 2 -eq 0 ']'
++ tailscale ip -4
+ HOST_IP=100.70.213.118
+ sudo test -f /etc/sysctl.d/haproxy.conf
[sudo] password for dhermes:
...
Executing: /lib/systemd/systemd-sysv-install enable haproxy
+ sudo systemctl restart haproxy
dhermes@nice-mcclintock:~$
dhermes@nice-mcclintock:~$
dhermes@nice-mcclintock:~$ tailscale ip -4
100.70.213.118
dhermes@nice-mcclintock:~$ netcat -v 100.70.213.118 6443
Connection to 100.70.213.118 6443 port [tcp/*] succeeded!

dhermes@nice-mcclintock:~$ rm k8s-load-balancer-proxy.sh
dhermes@nice-mcclintock:~$ exit
logout
Connection to nice-mcclintock closed.
```
