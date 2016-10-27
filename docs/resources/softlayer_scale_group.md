# `softlayer_scale_group`

Provides a `scale_group` resource. This allows auto scale groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Scale_Group).

## Example Usage

```hcl
# Create a new scale group using image "Debian"
resource "softlayer_scale_group" "test_scale_group" {
    name = "test_scale_group_name"
    regional_group = "as-sgp-central-1"
    minimum_member_count = 1
    maximum_member_count = 10
    cooldown = 30
    termination_policy = "CLOSEST_TO_NEXT_CHARGE"
    virtual_server_id = 267513
    port = 8080
    health_check = {
      type = "HTTP"
    }
    virtual_guest_member_template = {
      hostname = "test_virtual_guest_name"
      domain = "example.com"
      cores = 1
      memory = 1024
      network_speed = 100
      hourly_billing = true
      os_reference_code = "DEBIAN_7_64"
      # Optional Fields for virtual guest template (SL defaults apply):
      local_disk = false
      disks = [25]
      datacenter = "sng01"
      post_install_script_uri = ""
      ssh_key_ids = [383111]
      user_metadata = "#!/bin/bash ..."
    }
    # Optional Fields for scale_group:
    network_vlan_ids = [1234567, 7654321]
}
```

## Argument Reference

The following arguments are supported:

* `name` | *string*
    * Name of the scale group.
    * **Required**
* `regional_group` | *string*
    * Regional group for the scale group.
    * **Required**
* `minimum_member_count` | *int*
    * The fewest number of virtual guest members allowed in the scale group.
    * **Required**
* `maximum_member_count` | *int*
    * The greatest number of virtual guest members that are allowed in the scale group.
    * **Required**
* `cooldown` | *int*
    * Specifies the number of seconds this group will wait before performing another action.
    * **Required**
* `termination_policy` | *string*
    * Specifies the termination policy for the scaling group.
    * **Required**
* `virtual_guest_member_template` | *array*
    * This is the template to create guest memebers with. Only one template can be configured. Accepted values can be found [softlayer_virtual_guest](softlayer_virtual_guest.md).
    * **Required**
* `network_vlan_ids` | *array of numbers*
    * Collection of VLAN IDs for this auto scale group. Accepted values can be found [here](https://control.softlayer.com/network/vlans). Click on the desired VLAN and note the ID on the resulting URL. Or, you can also [refer to a VLAN by name using a data source](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_vlan.md).
    * *Default*: nil
    * *Optional*
* `virtual_server_id` | *int*
    * Specifies the id of a virtual server in a local load balancer. The id can be found with the following URL : `https://api.softlayer.com/rest/v3/SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress/[LOADBALANCER ID]/getObject?objectMask=virtualServers`. Replace **[LOADBALANER ID]** to an ID of a target load balacer's ID and open the URL with a web broswer. SoftLayer user name and API key are required.
    * *Optional*
* `port` | *int*
    * Specifies the port number in a local load balancer. For example 8080
    * *Optional*
* `health_check` | *map*
    * Specifies the type of health check in a local load balancer. For example HTTP. Also used to specify custom HTTP methods.
    * *Optional*

## Attributes Reference

The following attributes are exported:

* `id` - id of the scale group.
