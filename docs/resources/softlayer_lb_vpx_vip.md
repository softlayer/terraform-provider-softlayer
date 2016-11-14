# `softlayer_lb_vpx_vip`

Create, update, and delete Softlayer VPX Load Balancer Virtual IP Addresses. For additional details please refer to the [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_VirtualIpAddress).

_Please Note_: If Netscaler VPX 10.5 is used, terraform uses Netscaler's REST API([NITRO API](https://docs.citrix.com/en-us/netscaler/11/nitro-api.html)) for softlayer_lb_vpx_vip resource management. The NITRO API is only accessable in SoftLayer private network, so it is necessary to execute terraform in SoftLayer's private network. SoftLayer [SSL VPN](http://www.softlayer.com/VPN-Access) also can be used for private network connection. 

The following configuration supports Netscaler VPX 10.1 and 10.5
```hcl
resource "softlayer_lb_vpx_vip" "testacc_vip" {
    name = "test_load_balancer_vip"
    nad_controller_id = 1234567
    load_balancing_method = "lc"
    source_port = 80
    virtual_ip_address = "211.233.12.12"
    type = "HTTP"
}
```

Netscaler VPX 10.5 provides additional options for `load_balancing_method` and `persistence`. A private IP address can be used as a `virtual_ip_address`
```hcl
resource "softlayer_lb_vpx_vip" "testacc_vip" {
    name = "test_load_balancer_vip"
    nad_controller_id = "1234567"
    load_balancing_method = "DESTINATIONIPHASH"
    persistence = "SOURCEIP"
    source_port = 80
    virtual_ip_address = "10.10.2.2"
    type = "HTTP"
}
```

## Argument Reference

* `name` | *string*
    * (Required) The unique identifier for the VPX Load Balancer Virtual IP Address.
* `nad_controller_id` | *int*
    * (Required) The ID of the VPX Load Balancer that the VPX Load Balancer Virtual IP Address will be assigned to.
* `load_balancing_method` | *string*
    * (Required) The VPX Load Balancer load balancing method. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_VirtualIpAddress) for details. If Netscaler VPX 10.5 is used, additional methods can be used. See [Load Balancing Algorithms](https://docs.citrix.com/en-us/netscaler/10-5/ns-tmg-wrapper-10-con/ns-lb-wrapper-con-10/ns-lb-customizing-lbalgorithms-wrapper-con.html) for details. 
* `persistence` | *string*
    * (Netscaler VPX 10.5 only)
    * (Optional) Persistence option is available for Netscaler VPX 10.5. See [Persistence Type in Table.3](https://docs.citrix.com/en-us/netscaler/10-5/ns-tmg-wrapper-10-con/ns-lb-wrapper-con-10/ns-lb-persistence-wrapper-con/ns-lb-persistence-about-con.html) for details.  
* `virtual_ip_address` | *string*
    * (Required) The public facing IP address for the VPX Load Balancer Virtual IP.
* `source_port` | *int*
    * (Required) The source port for the VPX Load Balancer Virtual IP Address.
* `type` | *string*
    * (Required) The connection type for the VPX Load Balancer Virtual IP Address. Accepted values are `HTTP`, `FTP`, `TCP`, `UDP`, and `DNS`.

## Attributes Reference

* `id` - The VPX Load Balancer Virtual IPs unique identifier.
