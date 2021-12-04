# The Players

## Jump Host

<p align="center">
  <img src="../_images/suspicious-hawking.jpg?raw=true" />
</p>

The "jump host" is the primary computer I'll be using to administer all of the
machines in the Kubernetes cluster. We'll lock down all of our Kubernetes
nodes so that they are **only** reachable via the Tailnet, so the jump host
also needs to be in the Tailnet.

This machine is a 2021 [ASUS VivoBook][1] running Windows. It's my current
development machine and more importantly, my first time running Windows
since 2010. The WSL2 experience lets me pretend I am in Linux when coding
and for "consumer" type applications the Windows 11 experience has great
polish. (I will say WSL2 has **some** limitations, in particular on this
project having no access to a `tailscaled.sock` UDS and having no `systemd`
were both limiting.)

## Cluster

<p align="center">
  <img src="../_images/stoic-pike.jpg?raw=true" />
</p>

## Load Balancer

<p align="center">
  <img src="../_images/nice-mcclintock.jpg?raw=true" />
</p>

## Control Plane

### `pedantic-yonath`

<p align="center">
  <img src="../_images/pedantic-yonath-power-supply.jpg?raw=true" />
</p>

<p align="center">
  <img src="../_images/pedantic-yonath.jpg?raw=true" />
</p>

### `eager-jennings`

<p align="center">
  <img src="../_images/eager-jennings.jpg?raw=true" />
</p>

### `interesting-jang`

<p align="center">
  <img src="../_images/interesting-jang.jpg?raw=true" />
</p>

## Workers

### `nice-mcclintock`

<p align="center">
  <img src="../_images/nice-mcclintock.jpg?raw=true" />
</p>

### `relaxed-bouman`

<p align="center">
  <img src="../_images/relaxed-bouman.jpg?raw=true" />
</p>

### `agitated-feistel`

<p align="center">
  <img src="../_images/agitated-feistel.png?raw=true" />
</p>

[1]: https://www.amazon.com/gp/product/B01G1RUQHW/
