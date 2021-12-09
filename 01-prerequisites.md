# Prerequisites

We'll be coordinating everything from a **jump host**. The jump host is the
primary computer used to administer all of the machines in the Kubernetes
cluster. We'll lock down all of our Kubernetes nodes so that they are **only**
reachable via the Tailnet, so the jump host also needs to be in the Tailnet.
I'll be using Ubuntu 20.04.3 both for the jump host and all of the nodes.
(Technically it's WSL Ubuntu, a VM inside of Windows.)

## Install

For the orchestration actions we'll be doing, the following applications and
CLIs will be used:

- SSH client
- [Tailscale][1]
- `kubectl` [binary][2]
- `gcloud` [CLI][3]
- `aws` [CLI][4]
- **Optional**: `aws-vault` [CLI][5]

## Configure

In order to actually use these tools, we'll need to get some basic
configuration out of the way.

### Tailscale

```bash
tailscale up
```

### Amazon Web Services (AWS)

```bash
AWS_REGION=us-east-2
AWS_PROFILE_NAME=dhermes

aws configure set default.region "${AWS_REGION}"
# Below is optional
aws-vault add "${AWS_PROFILE_NAME}"
```

### Google Cloud Platform (GCP)

```bash
GCP_REGION=us-central1
GCP_ZONE=us-central1-a

gcloud init
gcloud auth login
gcloud config set compute/region"${GCP_REGION}"
gcloud config set compute/zone "${GCP_ZONE}"
```

### SSH

Make sure to generate an SSH key pair you can use to SSH into new Tailscale
devices. Additionally, create a `.extra_authorized_keys` file that we'll use
to append to `~/.ssh/authorized_keys` on new devices to allow connections
from our jump host and any other machine we like:

```bash
cp ~/.ssh/id_ed25519.pub .extra_authorized_keys
```

---

Next: [Prepare Tailscale Keys][6]

[1]: https://tailscale.com/download/linux/ubuntu-2004
[2]: https://kubernetes.io/docs/tasks/tools/
[3]: https://cloud.google.com/sdk/docs/install
[4]: https://aws.amazon.com/cli/
[5]: https://github.com/99designs/aws-vault
[6]: 02-prepare-tailscale-keys.md
