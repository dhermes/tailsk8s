# The Players

## Jump Host

![suspicious-hawking][1]

The "jump host" is the primary computer I'll be using to administer all of the
machines in the Kubernetes cluster. We'll lock down all of our Kubernetes
nodes so that they are **only** reachable via the Tailnet, so the jump host
also needs to be in the Tailnet.

This machine is a 2021 [ASUS VivoBook][10] running Windows. It's my current
development machine and more importantly, my first time running Windows
since 2010. The WSL2 experience lets me pretend I am in Linux when coding
and for "consumer" type applications the Windows 11 experience has great
polish. (I will say WSL2 has **some** limitations, in particular on this
project having no access to a `tailscaled.sock` UDS and having no `systemd`
were both limiting.)

## Cluster

![stoic-pike][2]

## Load Balancer

![nice-mcclintock][7]

## Control Plane

### `pedantic-yonath`

![pedantic-yonath ... power supply][3]
![pedantic-yonath][4]

### `eager-jennings`

![eager-jennings][5]

### `interesting-jang`

![interesting-jang][6]

## Workers

### `nice-mcclintock`

![nice-mcclintock][7]

### `relaxed-bouman`

![relaxed-bouman][8]

### `agitated-feistel`

![agitated-feistel][9]

[1]: ../_images/suspicious-hawking.jpg
[2]: ../_images/stoic-pike.jpg
[3]: ../_images/pedantic-yonath-power-supply.jpg
[4]: ../_images/pedantic-yonath.jpg
[5]: ../_images/eager-jennings.jpg
[6]: ../_images/interesting-jang.jpg
[7]: ../_images/nice-mcclintock.jpg
[8]: ../_images/relaxed-bouman.jpg
[9]: ../_images/agitated-feistel.png
[10]: https://www.amazon.com/gp/product/B01G1RUQHW/
