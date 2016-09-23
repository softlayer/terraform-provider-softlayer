# `softlayer_lb_local_service_group`

Provides a `lb_local_service_group` resource. This allows local load balancer service groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service_Group).

## Example Usage

```hcl
# Create a new local load balancer service group
resource "softlayer_lb_local_service_group" "test_service_group" {
    port = 82
    routing_method = "CONSISTENT_HASH_IP"
    routing_type = "HTTP"
    load_balancer_id = "${softlayer_lb_local.test_lb_local.id}"
    allocation = 100
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` | *int*
    * Set the Id of the local load balancer.
    * **Required**
* `allocation` | *int*
    * Set the allocation field for the load balancer service group.
    * **Required**
* `port` | *int*
    * Set the port for the local load balancer service group.
    * **Required**
* `routing_method` | *string*
    * Set the routing method for the load balancer group. For example CONSISTENT_HASH_IP
    * **Required**
* `routing_type` | *string*
    * Set the routing type for the group.
    * **Required**

## Attributes Reference

The following attributes are exported:

* `virtual_server_id` - id of the virtual server.
* `service_group_id` - id of the load balancer service group.
