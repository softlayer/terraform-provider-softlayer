# `softlayer_global_ip`

Provides a `global_ip` resource. This allows Global Ip's to be created, updated, and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/services/SoftLayer_Network_Subnet_IpAddress_Global) and [Global IP Address Overview](https://knowledgelayer.softlayer.com/learning/global-ip-addresses).

## Example Usage

```hcl
# Create a global IPv4 address
resource "softlayer_global_ip" "test_global_ip " {
    routes_to = "119.81.82.163"
}
```

```hcl
# Create a global IPv6 address
resource "softlayer_global_ip" "test_global_ip " {
    routes_to = "2401:c900:1501:0032:0000:0000:0000:0003"
}
```

## Argument Reference

The following arguments are supported:

* `routes_to` | *string*
     * Destination ip address which the global IP route traffic through. The destination ip address can be a public ip address of SoftLayer resources in the same account such as a public ip address of virtual_guests and public virtual ip address of netscaler VPXs. 
     * **Required**

## Attributes Reference

The following attributes are exported:

* `id` - id of the global ip
* `ip_address` - ip address of the global ip
