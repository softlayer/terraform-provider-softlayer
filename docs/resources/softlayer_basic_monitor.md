# `softlayer_basic_monitor`

Provides a `basic_monitor` resource. This allows basic monitors to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Monitor_Version1_Query_Host).

## Example Usage

```hcl
# Create a new basic monitor
resource "softlayer_basic_monitor" "test_basic_monitor" {
    guest_id = ${softlayer_virtual_guest.test_server.id}
    ip_address = ${softlayer_virtual_guest.test_server.id.ipv4_address}
    query_type_id = 1
    response_action_id = 1
    wait_cycles = 5
    notified_users = [460547]
}
```

## Argument Reference

The following arguments are supported:

* `guest_id` | *int*
    * Set the Id of the virtual guest being monitored.
    * **Required**
* `ip_address` | *string*
    * Set the ip address to be monitored.
    * **Optional**
* `query_type_id` | *int*
    * Set the id of the query type.
    * **Required**
* `response_action_id` | *int*
    * Set the id of the response action to take when the monitor fails. Accepted values are 1,2
    * **Required**
* `wait_cycles` | *int*
    * Set the number of 5-minute cycles to wait before the response action is taken.
    * **Optional**
* `notified_users` | *array of ints*
    * Set the list of user id's to be notified.
    * **Optional**

## Attributes Reference

The following attributes are exported:

* `id` - id of the basic monitor.
* `notified_users` - the list of user id's to be notified.
