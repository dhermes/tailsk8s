# `tailsk8s`: Bare Metal Kubernetes with Tailscale

> Pronounced "Tail Skates"

[![Go Reference][1]][2]

## Terms and Definitions

- **Bare Metal**: this is primarily intended to mean "**not** virtualized",
  i.e. not running in someone else's data center. The expectation is that the
  reader has physical possession of a bare metal machine and can manually
  administer if e.g. networking or power completely fail

## References, Documentation and Motivation

Over the **many** days of getting this off the ground, I leaned heavily on
the work and writings of others. I could not have made nearly as much progress
without example projects, blog posts and great documentation. Some that were
particularly helpful:

- [kelseyhightower/kubernetes-the-hard-way][3] from Kelsey Hightower
- [Deploying Kubernetes on Bare Metal][4] by [Layachi Khodja][8]
- [rmb938/tailscale-cni][5] from Ryan Belgrave
- [`kubeadm init/join` and ExternalIP vs InternalIP][6] from
  [Alasdair Lumsden][7]

## Development

```
$ make  # Or `make help`
Makefile for the `tailsk8s` project

Usage:
   make tailscale-advertise-linux-amd64    Build static `tailscale-advertise` binary for linux/amd64
   make tailscale-authorize-linux-amd64    Build static `tailscale-authorize` binary for linux/amd64
   make tailscale-withdraw-linux-amd64     Build static `tailscale-withdraw` binary for linux/amd64
   make release                            Build all static binaries

```

[1]: https://pkg.go.dev/badge/github.com/dhermes/tailsk8s.svg
[2]: https://pkg.go.dev/github.com/dhermes/tailsk8s
[3]: https://github.com/kelseyhightower/kubernetes-the-hard-way/tree/79a3f79b27bd28f82f071bb877a266c2e62ee506
[4]: https://www.inap.com/blog/deploying-kubernetes-on-bare-metal/
[5]: https://github.com/rmb938/tailscale-cni/tree/dba6992227958e61ac85b3168dbcae4ff10dde57
[6]: https://medium.com/@aleverycity/kubeadm-init-join-and-externalip-vs-internalip-519519ddff89
[7]: https://github.com/alaslums
[8]: https://linkedin.com/in/layachi-khodja-38428a1
