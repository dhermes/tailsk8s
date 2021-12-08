# Smoke Test

## Interact with Kubernetes API from Jump Host

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get nodes --output wide
NAME              STATUS   ROLES                  AGE     VERSION   INTERNAL-IP       EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
eager-jennings    Ready    control-plane,master   19m     v1.22.4   100.109.83.23     <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
nice-mcclintock   Ready    <none>                 10m     v1.22.4   100.70.213.118    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
pedantic-yonath   Ready    control-plane,master   36m     v1.22.4   100.110.217.104   <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
relaxed-bouman    Ready    <none>                 4m16s   v1.22.4   100.122.162.98    <none>        Ubuntu 20.04.3 LTS   5.11.0-41-generic   docker://20.10.11
```

## Apply Manifest

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml apply --filename _templates/httpbin.manifest.yaml
namespace/httpbin created
serviceaccount/httpbin created
deployment.apps/httpbin created
service/httpbin created
$
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get pods --all-namespaces --output wide
NAMESPACE     NAME                                      READY   STATUS    RESTARTS       AGE     IP                NODE              NOMINATED NODE   READINESS GATES
httpbin       httpbin-6698c4cbc-4d7hj                   1/1     Running   0              23s     10.100.2.3        nice-mcclintock   <none>           <none>
httpbin       httpbin-6698c4cbc-8svnc                   1/1     Running   0              23s     10.100.2.2        nice-mcclintock   <none>           <none>
httpbin       httpbin-6698c4cbc-fm94t                   1/1     Running   0              23s     10.100.3.2        relaxed-bouman    <none>           <none>
kube-system   coredns-78fcd69978-2mxgw                  1/1     Running   0              39m     10.100.0.3        pedantic-yonath   <none>           <none>
kube-system   coredns-78fcd69978-7rvml                  1/1     Running   0              39m     10.100.0.2        pedantic-yonath   <none>           <none>
kube-system   etcd-eager-jennings                       1/1     Running   10             22m     100.109.83.23     eager-jennings    <none>           <none>
kube-system   etcd-pedantic-yonath                      1/1     Running   29             39m     100.110.217.104   pedantic-yonath   <none>           <none>
kube-system   kube-apiserver-eager-jennings             1/1     Running   6 (22m ago)    22m     100.109.83.23     eager-jennings    <none>           <none>
kube-system   kube-apiserver-pedantic-yonath            1/1     Running   8              39m     100.110.217.104   pedantic-yonath   <none>           <none>
kube-system   kube-controller-manager-eager-jennings    1/1     Running   5              21m     100.109.83.23     eager-jennings    <none>           <none>
kube-system   kube-controller-manager-pedantic-yonath   1/1     Running   10 (22m ago)   39m     100.110.217.104   pedantic-yonath   <none>           <none>
kube-system   kube-proxy-62gj2                          1/1     Running   0              22m     100.109.83.23     eager-jennings    <none>           <none>
kube-system   kube-proxy-kw87d                          1/1     Running   0              39m     100.110.217.104   pedantic-yonath   <none>           <none>
kube-system   kube-proxy-t8tvf                          1/1     Running   0              7m58s   100.122.162.98    relaxed-bouman    <none>           <none>
kube-system   kube-proxy-wv52v                          1/1     Running   0              13m     100.70.213.118    nice-mcclintock   <none>           <none>
kube-system   kube-scheduler-eager-jennings             1/1     Running   11             22m     100.109.83.23     eager-jennings    <none>           <none>
kube-system   kube-scheduler-pedantic-yonath            1/1     Running   43 (22m ago)   39m     100.110.217.104   pedantic-yonath   <none>           <none>
$
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml get services --all-namespaces --output wide
NAMESPACE     NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                  AGE   SELECTOR
default       kubernetes   ClusterIP   10.101.0.1       <none>        443/TCP                  40m   <none>
httpbin       httpbin      ClusterIP   10.101.177.244   <none>        8000/TCP                 38s   app=httpbin
kube-system   kube-dns     ClusterIP   10.101.0.10      <none>        53/UDP,53/TCP,9153/TCP   40m   k8s-app=kube-dns
```

We can reach **pods** from the jump host because

```
$ curl http://10.100.2.3:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.2.3",
    "User-Agent": "curl/7.68.0"
  }
}
$ curl http://10.100.2.2:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.2.2",
    "User-Agent": "curl/7.68.0"
  }
}
$ curl http://10.100.3.2:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.3.2",
    "User-Agent": "curl/7.68.0"
  }
}
```

but cannot reach the service (this is expected)

```
$ curl --max-time 5 http://10.101.177.244:8000/headers
curl: (28) Connection timed out after 5001 milliseconds
```

**Within** the cluster (where `kubelet` can write custom `iptables` rules),
we can reach the service as well:

```
$ ssh dhermes@pedantic-yonath
dhermes@pedantic-yonath:~$ curl http://10.101.177.244:8000/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.101.177.244:8000",
    "User-Agent": "curl/7.68.0"
  }
}
dhermes@pedantic-yonath:~$ curl http://10.100.2.3:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.2.3",
    "User-Agent": "curl/7.68.0"
  }
}
dhermes@pedantic-yonath:~$ curl http://10.100.2.2:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.2.2",
    "User-Agent": "curl/7.68.0"
  }
}
dhermes@pedantic-yonath:~$ curl http://10.100.3.2:80/headers
{
  "headers": {
    "Accept": "*/*",
    "Host": "10.100.3.2",
    "User-Agent": "curl/7.68.0"
  }
}
```

Check out Kubernetes DNS as well:

```
dhermes@pedantic-yonath:~$ nslookup httpbin.httpbin.svc.cluster.local. 10.101.0.10
Server:         10.101.0.10
Address:        10.101.0.10#53

Name:   httpbin.httpbin.svc.cluster.local
Address: 10.101.177.244
```

Clean Up:

```
$ kubectl --kubeconfig k8s-bootstrap-shared/kube-config.yaml delete --filename _templates/httpbin.manifest.yaml
namespace "httpbin" deleted
serviceaccount "httpbin" deleted
deployment.apps "httpbin" deleted
service "httpbin" deleted
```

---

Next: [Add an AWS EC2 VM to the Kubernetes Cluster][1]

[1]: 13-add-vm-aws.md
