#### `softlayer_network_vlan`

Provides a `network_vlan` resource. This allows public and private network vlans to be created, updated and cancelled. To create

For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Vlan).

##### Example Usage

```hcl
# Create a new network vlan
resource "softlayer_network_vlan" "test_vlan" {
   name = "test_vlan"
   datacenter = "lon02"
   type = "PUBLIC"
   primary_subnet_size = 8
   primary_router_hostname = "fcr01a.lon02"
}
```

##### Argument Reference

The following arguments are supported:

* `datacenter` | *string*
    * Set the datacenter in which the network VLAN resides.
    * **Required**
* `type` | *string*
    * Set the type of the VLAN if it is public or private. Accepted values are PRIVATE and PUBLIC.
    * **Required**
* `primary_subnet_size` | *int*
    * Set the size of the primary subnet for the VLAN. Accepted values are 8, 16, 32, and 64.
    * **Required** 
* `name` | *string*
    * Set the name for the VLAN.
    * **Optional**
* `primary_router_hostname` | *string*
    * Set the hostname of the primary router that the VLAN is associated with.
    * **Optional**

##### Attributes Reference

The following attributes are exported:

* `id` - id of the network VLAN.
* `vlan_number` - The VLAN number as recorded within the SoftLayer network. This is configured directly on SoftLayer's networking equipment.
* `softlayer_managed` - Whether the network vlan is managed by SoftLayer or not. If the vlan is created by SoftLayer automatically while other
 resources are created, `softlayer_managed` is true. If the vlan is created by user via SoftLayer API, portal, or ticket, `softlayer_managed`
 is false.
* `child_resource_count` - A count of all of the resources such as Virtual Servers and other network components that are connected to the VLAN. 
* `subnets` - Collection of subnets associated with the VLAN.
