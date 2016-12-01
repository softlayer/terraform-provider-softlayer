# Gateway

This sample configuration will configure one VM with a public IP to become the
gateway for another VM that is on the private network only. This will enable
the private network only VM to route out the Internet using NATing at the gateway
without being on the public Internet.

## Prerequisites

* This configuration assumes an ssh keypair in `~/.ssh`. Ensure that the correct
  *ssh_key_label* for your public key is selected Note: The private key must NOT
  be password protected, as terraform does not support password protected keys at
  this time.
* Since the private network only node is provisioned without a public IP, you
  must be connected to the VPN gateway that matches the datacenter where the
  machines are provisioned, so that the post-provision script on the nodes can
  be executed over SSH.

## Variables

Set in the config files, or override via `TF_ENV_varname`

* `datacenter` | *string*
  * The SoftLayer datacenter to use to provision the systems
  * **Required**
  * *Default*: sjc01
* `ssh_key_label` | *string*
  * The label of the public ssh key you want to place in the provisioned systems
  * **Required**
* `ssh_key_path` | *string*
  * The path to the public ssh key you want to upload to SoftLayer.
  * **Required**
