#### `softlayer_vlan`

Provides a `VLAN` resource. This allows public and private VLANs to be created, updated and cancelled. Default SoftLayer account does not have a permission to create a new VLAN by SoftLayer API. To create a new VLAN with 
terraform, user should have a VLAN creation permission in advance. Contact a SoftLayer sales person or open a ticket.
Existed VLANs can be managed by terraform with `terraform import` command. It requires SoftLayer VLAN ID in [VLANs](https://control.softlayer.com/network/vlans). 
Once they are imported, they provides useful information such as subnets and child_resource_count. When `terraform destroy`
is executed, the VLANs' billing item will be deleted. However, VLAN will be remained in SoftLayer until resources such as 
virtual guests, secondary subnets, and firewalls on the VLAN are deleted. 

For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Vlan).

##### Example Usage

```hcl
# Create a new vlan
resource "softlayer_vlan" "test_vlan" {
   name = "test_vlan"
   datacenter = "lon02"
   type = "PUBLIC"
   subnet_size = 8
   router_hostname = "fcr01a.lon02"
}
```

##### Argument Reference

The following arguments are supported:

* `datacenter` | *string*
    * Set the datacenter in which the VLAN resides.
    * **Required**
* `type` | *string*
    * Set the type of the VLAN if it is public or private. Accepted values are PRIVATE and PUBLIC.
    * **Required**
* `subnet_size` | *int*
    * Set the size of the primary subnet for the VLAN. Accepted values are 8, 16, 32, and 64.
    * **Required** 
* `name` | *string*
    * Set the name for the VLAN.
    * **Optional**
* `router_hostname` | *string*
    * Set the hostname of the primary router that the VLAN is associated with.
    * **Optional**

##### Attributes Reference

The following attributes are exported:

* `id` - id of the VLAN.
* `vlan_number` - The VLAN number as recorded within the SoftLayer network. This is configured directly on SoftLayer's networking equipment.
* `softlayer_managed` - Whether the VLAN is managed by SoftLayer or not. If the VLAN is created by SoftLayer automatically while other
 resources are created, `softlayer_managed` is true. If the VLAN is created by user via SoftLayer API, portal, or ticket, `softlayer_managed`
 is false.
* `child_resource_count` - A count of all of the resources such as Virtual Servers and other network components that are connected to the VLAN. 
* `subnets` - Collection of subnets associated with the VLAN.
