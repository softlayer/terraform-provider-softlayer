#### `softlayer_subnet`

Provides `Portable subnet` and `Static subnet` resource. Users are able to create `Public Portable Subnet`, `Private Portable Subnet`,
 and `Public Static Subnet` using `softlayer_subnet` resource.
 
The `Portable subnet` is created as a seconday subnet in the VLAN and IP addresses in the subnet can be assigned as a secondary IP 
 address for resources in the VLAN. `Portable subnet` is an independant subnet and has a default gateway, network IP address, and broadcast 
 IP address. For example, if a subnet is `10.0.0.0/30`, `10.0.0.0` is a network IP address, `10.0.0.1` is a default gateway, and `10.0.0.3` 
 is a broadcast IP address. Therefore, only `10.0.0.2` can be assigned to resources as a secondary IP address. Number of usuable IP addresses is
 `capacity - 3`. If capacity is 4, number of usuable IP addresses is (4-3)=1. If capacity is 8, number of usuable IP addresses is (8-3)=5. 
 For additional details please refer to [Static and Portable IP blocks](https://knowledgelayer.softlayer.com/articles/static-and-portable-ip-blocks).

The `Static subnet` provides secondary IP addresses for a primary IP address. It works as a secondary IP address for a specific resource such as 
virtual server, bare metal server, and netscaler VPX. Suppose that a virtual server requires additional secondary IP addresses. Then, users can create 
`Static subnet` for the virtual server. Unlike `Portable subnet`, `capacity` is same with a number of usuable IP address. For example, if a subnet is 
`10.0.0.0/30`, `10.0.0.0`~`10.0.0.3` are usuable IP addresses. 
 
For additional details please refer to [Subnet](https://knowledgelayer.softlayer.com/topic/subnets).

The following example will create a private portable subnet which has one available portable IP address. 
##### Example Usage of portable subnet

```hcl
# Create a new portable subnet
resource "softlayer_subnet" "portable_subnet" {
  type = "Portable"
  network = "PRIVATE"
  ip_version = 4
  capacity = 4
  vlan_id = 1234567
  notes = "portable_subnet"
}
```

The following example will create a public static subnet which has four available portable IP address.
##### Example Usage of static subnet

```hcl
# Create a new static subnet
resource "softlayer_subnet" "static_subnet" {
  type = "Static"
  network = "PUBLIC"
  ip_version = 4
  capacity = 4
  endpoint_ip="151.1.1.1"
  notes = "static_subnet_updated"
}
```
##### Argument Reference

The following arguments are supported:

* `network` | *string*
    * Set the network property of the subnet if it is public or private. Accepted values are PRIVATE and PUBLIC.
    * **Required**
* `type` | *string*
    * Set the type of the subnet. Accepted values are Portable and Static.
    * **Required**
* `ip_version` | *int*
    * Set the IP version of the subnet. Accepted values are 4 and 6.
    * **Required**
* `capacity` | *int*
    * Set the size of the subnet. Accepted values are 4, 8, 16, 32, and 64.
    * **Required** 
* `vlan_id` | *int*
    * VLAN id for portable subnet. It should be configured when the subnet is a portable subnet. Both public VLAN ID and private VLAN ID can 
    be configured. Accepted values can be found [here](https://control.softlayer.com/network/vlans). Click on the desired VLAN and note the 
    ID on the resulting URL. Or, you can also [refer to a VLAN by name using a data source](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_vlan.md). 
    * **Optional**
* `endpoint_ip` | *string*
    * Target primary IP address for static subnet. It should be configured when the subnet is a static subnet. Only public IP address can be 
    configured as a `endpoint_ip`. It can be public IP address of virtual servers, bare metal servers, and netscaler VPXs. `static subnet` will 
    be created on VLAN where `endpoint_ip` is located in.
    * **Optional**
* `notes` | *string*
    * Set comments for the subnet.
    * **Optional**

##### Attributes Reference

The following attributes are exported:

* `id` - id of the subnet.
* `subnet` - It rovides IP address/netmask format (ex. 10.10.10.10/28). It can be used to get an available IP address in `subnet`. Users can use built-in functions to get 
available IP addresses from `subnet`. For example, the following example returns first IP address in the subnet:
```
#resource "softlayer_subnet" "test" {
  type = "Static"
  network = "PUBLIC"
  ip_version = 4
  capacity = 4
  endpoint_ip="159.8.181.82"
}

output "first_ip_address" {
  value = "${cidrhost(softlayer_subnet.test.subnet,0)}"
}
```