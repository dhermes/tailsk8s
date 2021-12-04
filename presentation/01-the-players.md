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

In order to "randomize" the machines on my Tailnet, I turned to the Moby
[name generator][2]. (Most Docker users will have unknowingly used this name
generator when invoking `docker run` without the `--name` flag.) I didn't
do anything special to seed the random name generation other than
`rand.Seed(time.Now().UnixNano())`, yet when it came time to generate the name
for the Kubernetes cluster I got none other than [Rob Pike][3] himself.

<p align="center">
  <img src="../_images/stoic-pike.jpg?raw=true" />
</p>

This is incredibly appropriate! Rob is one of the three creators of the Go
programming language. Over the last [twelve years][4], Go has been a primary
driver of systems and cloud Open Source innovation. The two projects most
responsible for `tailsk8s` (Kubernetes and Tailscale) are essentially 100%
written in Go.

It was fate; what a great cluster name!

## Bare Metal

### `pedantic-yonath`

Lurking around the corner a now-defunct MacBook charging block can be
seen sticking out of the wall:

<p align="center">
  <img src="../_images/pedantic-yonath-power-supply.jpg?raw=true" />
</p>

It leads to an entire Early 2011 [MacBook Pro][12] under the bed!

<p align="center">
  <img src="../_images/pedantic-yonath.jpg?raw=true" />
</p>

Though it is on the older end of the devices in my home cluster, it's still
Apple hardware (running Ubuntu 20.04.3). In fact, it has the highest end
CPU, disk and RAM, so it will be the lynchpin of the cluster. (The disk is
an [SSD][13] I had to swap in when OS X performance dropped off a cliff on
HDDs in summer 2015.)

### `eager-jennings`

In my office you can find `pedantic-yonath` under a bed; one room over you
can find `eager-jennings` under a dresser:

<p align="center">
  <img src="../_images/eager-jennings.jpg?raw=true" />
</p>

Other than the jump host (which is not in the cluster, just in the Tailnet)
this is the newest machine I own. As a result, this machine will be one of the
control plane nodes. This machine is a 2017 [ASUS ZenBook][5] running
Ubuntu 20.04.3.

### `nice-mcclintock`

Under a bed, under a dresser, now head downstairs and look under a chair and
you'll find `nice-mcclintock`:

<p align="center">
  <img src="../_images/nice-mcclintock.jpg?raw=true" />
</p>

This laptop is reliable, but old and underpowered. The WiFi radio is 2.4GHz
only and somehow the machine consistently has issues with Eero switching
between 2.4GHz and 5GHz at will (this is likely a software problem in the OS,
but not one I plan to solve). Instead of dealing with this, it's just hardwired
into the Eero. Since it is underpowered, it is a worker node. This machine is
a 2014 [Lenovo ThinkPad Edge][6] running Ubuntu 20.04.3.

### `relaxed-bouman`

In a somewhat more conventional place, the last tower I ever bought (as of
this writing) can be found under the desk in my office:

<p align="center">
  <img src="../_images/relaxed-bouman.jpg?raw=true" />
</p>

This machine is very old (2011) and hadn't been turned on since 2016 and so
it definitely will be a worker node. This machine is a 2011
[Lenovo Essential H405][7] running Ubuntu 20.04.3. In 2011, I was still
buying hardware on TigerDirect!

## Cloud Provider VMs

Whether it's

> My Other Computer Is a Data Center

or

> The Cloud Is Just Someone Else's Computer

for the purposes of the demonstration, it really is true. We'll grab two VMs
from **different** cloud providers and easily add them to our Tailnet and
Kubernetes cluster as if they were sitting under some furniture in my house
like the rest of the nodes.

### `interesting-jang`

When "someone else's computer" is [AWS EC2][8] we use a `t3.micro`
[instance][9] since it's just meant to be a toy and we'll tear it down
shortly.

<p align="center">
  <img src="../_images/interesting-jang.jpg?raw=true" />
</p>

`t3.micro`

### `agitated-feistel`

When "someone else's computer" is [GCP GCE][10] we use a `e2-micro`
[instance][11] (similar comments on instance size and lifetime).

<p align="center">
  <img src="../_images/agitated-feistel.png?raw=true" />
</p>

`e2-micro`

## Load Balancer

- `nice-mcclintock`

## Control Plane

- `pedantic-yonath`
- `eager-jennings`
- `interesting-jang`

## Workers

- `nice-mcclintock`
- `relaxed-bouman`
- `agitated-feistel`

[1]: https://www.amazon.com/gp/product/B01G1RUQHW/
[2]: https://github.com/moby/moby/blob/v20.10.11/pkg/namesgenerator/names-generator.go
[3]: https://en.wikipedia.org/wiki/Rob_Pike
[4]: https://go.dev/blog/12years
[5]: https://www.amazon.com/gp/product/B01CQRNBJG/
[6]: https://www.amazon.com/gp/product/B00D5TPT4A/
[7]: https://www.tigerdirect.com/applications/searchtools/item-details.asp?EdpNo=417541
[8]: https://aws.amazon.com/ec2/
[9]: https://aws.amazon.com/ec2/instance-types/t3/
[10]: https://cloud.google.com/compute
[11]: https://cloud.google.com/compute/docs/general-purpose-machines
[12]: https://support.apple.com/kb/SP619
[13]: https://www.amazon.com/gp/product/B00OAJ412U/
