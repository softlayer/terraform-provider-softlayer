# Elasticsearch/Kibana cluster

This sample configuration will stand up a cluster of elasticsearch/kibana nodes, running on virtual guests, behind an HAProxy load balancer. The URL of the load balancer is provided as an output.

## Prerequisites

* This configuration assumes an ssh keypair in `~/.ssh`. SSH will use the default private keys there. Ensure that the correct *ssh_key_label* for your public key is selected (must be pre-registered in SoftLayer) corresponding to your default private key. Note: The private key must NOT be password protected, as terraform does not support password protected keys at this time.
* Since the elasticsearch nodes are provisioned without a public IP, you must be connected to the VPN gateway that matches the datacenter where the machines are provisioned, so that the post-provision script on the nodes can be executed over SSH.
* [Download the kibana package](https://download.elastic.co/kibana/kibana/kibana-4.1.11-linux-x64.tar.gz) and put it in the same directory as this readme file.

## Variables

Set in the config files, or override via `TF_ENV_varname`

* `node_count` | *int*
    * The number of elasticsearch nodes
    * *Optional*
    * *Default*: 2
* `datacenter` | *string*
    * The SoftLayer datacenter to use to provision the systems
    * **Required**
    * *Default*: dal06
* `private_vlan_number` | *int*
    * The vlan number on which the nodes will be provisioned.
    * **Required**
* `private_router_hostname` | *string*
    * The hostname of the primary router on which the nodes will be provisioned
    * **Required**
* `ssh_key_label` | *string*
    * The label of the ssh key you want to place in the provisioned systems
    * **Required**
* `kibana_package` | *string*
    * The relative path to the kibana package to send to the cluster nodes. Kibana does not exist in the Ubuntu package repositories, and with the nodes being on the private network without access to the public Internet, this is used to push the package to them (over the required VPN).
    * **Required**
