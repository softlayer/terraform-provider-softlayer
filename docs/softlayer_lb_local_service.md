# `softlayer_lb_local_service`

Provides a `lb_local_service` resource. This allows local load balancer service to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service).

## Example Usage

```hcl
# Create a new local load balancer service
resource "softlayer_lb_local_service" "test_lb_local_service" {
    port = 80
    enabled = true
    service_group_id = "${softlayer_lb_local_service_group.test_service_group.service_group_id}"
    weight = 1
    health_check_type = "DNS"
    ip_address_id = "${softlayer_virtual_guest.test_server.ip_address_id}"
}

```

## Argument Reference

The following arguments are supported:

* `service_group_id` | *int*
    * Set the Id of the local load balancer service group.
    * **Required**
* `ip_address_id` | *int*
    * Set the Id of the virtual server.
    * **Required**
* `port` | *int*
    * Set the port for the local load balancer service.
    * **Required**
* `enabled` | *boolean*
    * Set the enabled field for the load balancer service. Accepted values are true, false.
    * **Required**
* `health_check_type` | *string*
    * Set the health check type for the load balancer service.
    * **Required**
* `weight` | *int*
    * Set the weight for the load balancer service.
    * **Required**
