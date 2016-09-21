# `softlayer_lb_local`

Provides a `lb_local` resource. This allows local load balancers to be created, updated and deleted.

## Example Usage

```hcl
# Create a new local load balancer
resource "softlayer_lb_local" "test_lb_local" {
    connections = 1500
    datacenter = "tok02"
    ha_enabled = false
    dedicated = false
}
```

## Argument Reference

The following arguments are supported:

* `connections` | *int*
    * Set the number of connections for the local load balancer.
    * **Required**
* `datacenter` | *string*
    * Set the data center for the local load balancer.
    * **Required**
* `ha_enabled` | *boolean*
    * Set if the local load balancer needs to be HA enabled or not.
    * **Required**
* `security_certificate_id` | *int*
    * Set the Id of the security certificate associated with the local load balancer.
    * **Optional**
* `dedicated` | *boolean*
    * Set if the local load balancer needs to be shared or dedicated.
    * **Optional**

## Attributes Reference

The following attributes are exported:

* `id` - id of the local load balancer.
* `ip_address` - The IP Address of the local load balancer.
* `subnet_id` - The Id of the subnet associated with the local load balancer.
* `ssl_enabled` - If the local load balancer provides ssl capability or not.
