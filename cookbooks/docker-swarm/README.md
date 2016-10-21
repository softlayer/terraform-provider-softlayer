# Docker swarm

Terraform config to stand up a Docker swarm.

## Prerequisites

* This configuration assumes an ssh keypair in `~/.ssh`. SSH will use the default private keys there. Ensure that the correct *ssh_key_label* for your public key is selected (must be pre-registered in SoftLayer) corresponding to your default private key. Note: The private key must NOT be password protected, as terraform does not support password protected keys at this time.

## Variables

Set in the config files, or override via `TF_ENV_varname`

* `worker_count` | *int*
    * The number of docker swarm worker nodes
    * *Optional*
    * *Default*: 5
* `datacenter` | *string*
    * The SoftLayer datacenter to use to provision the systems
    * **Required**
    * *Default*: wdc01
* `ssh_key_label` | *int*
    * The vlan number on which the nodes will be provisioned.
    * **Required**
