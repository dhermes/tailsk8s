# Prepare Tailscale Keys

## Securing Tailnet

You'll likely want to enable [device authorization][1] to ensure only trusted
devices can access the Tailnet:

<p align="center">
  <img src="./_images/tailscale-require-device-authorization.png?raw=true" />
</p>

## API Key

Most Tailscale interactions will happen locally (i.e. with the Tailscale
daemon's API) but some things must be done with the Tailscale "cloud" [API][2].
We'll need to use this API to authorize new devices, enable subnet routes once
advertised by a device or disable subnet routes once withdrawn by a device.

Generate an API key and store it in `k8s-bootstrap-shared/tailscale-api-key`:

<p align="center">
  <img src="./_images/tailscale-new-api-key-01.png?raw=true" />
</p>

<p align="center">
  <img src="./_images/tailscale-new-api-key-02.png?raw=true" />
</p>

<p align="center">
  <img src="./_images/tailscale-new-api-key-03.png?raw=true" />
</p>

## Prepare For New Devices

As we bring up our cluster, we'll be added 4 bare metal machines and 2 cloud
provider VMs. In order to automate the process of joining the Tailnet, we'll
generate six [one-off keys][3]. (It's tempting to use a reusable key or even
ephemeral key here, but having better security hygiene here is not a large
cost with only 6 devices.) We'll store these keys locally, for example:

- `tailscale-one-off-key-KC`
- `tailscale-one-off-key-HW`
- `tailscale-one-off-key-NL`
- `tailscale-one-off-key-YG`
- `tailscale-one-off-key-AA`
- `tailscale-one-off-key-PT`

<p align="center">
  <img src="./_images/tailscale-create-new-key.png?raw=true" />
</p>

<p align="center">
  <img src="./_images/tailscale-new-one-off-key.png?raw=true" />
</p>

<p align="center">
  <img src="./_images/tailscale-prepared-one-off-keys.png?raw=true" />
</p>

----

Next: [The Players][4]

[1]: https://tailscale.com/kb/1099/device-authorization/
[2]: https://tailscale.com/kb/1101/api/
[3]: https://tailscale.com/kb/1085/auth-keys/
[4]: 03-the-players.md
