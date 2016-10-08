# Elasticsearch cluster

This sample configuration will stand up a cluster of elasticsearch nodes, running on virtual guests, behind a local load balancer. The URL of the cluster is provided as an output

Note: the elasticsearch nodes in this example cluster listen on their public IP address, which is generally not recommended for production deployments. This is done mainly to demonstrate how to align VMs behind a local load balancer in SoftLayer.

## Prerequisites

* This configuration assumes an ssh keypair in `~/.ssh` named `es_id_rsa` and `es_id_rsa.pub`. This will be used to connect to the VMs to configure the cluster. Note: The private key must NOT be password protected, as terraform does not support password protected keys at this time.

## Variables

Set in the config files, or override via `TF_ENV_varname`

*   `node_count` | *int*
    * The number of elasticsearch nodes
    * *Optional*
    * *Default*: 2
*   `port` | *int*
    * The port on which the load balancer will listen
    * *Optional*
    * *Default*: 9200
*   `backend_subnet` | *string*
    * The subnet on which the nodes will be provisioned. In order for the nodes to discover one another and form a cluster, they must be provisioned on the same subnet. The subnet must be valid for the provided vlan number and primary router hostname. E.g., "10.56.58.0/26"
    * **Required**
*   `backend_vlan_number` | *int*
    * The vlan number on which the nodes will be provisioned.
    * **Required**
*   `backend_primary_router_hostname` | *string*
    * The hostname of the primary router on which the nodes will be provisioned
    * **Required**
