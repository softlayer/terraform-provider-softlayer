# `softlayer_lb_vpx_service`

Create, update, and delete Softlayer VPX Load Balancer Services. For additional details please refer to the [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service).

_Please Note_: If Netscaler VPX 10.5 is used, terraform uses Netscaler's REST API([NITRO API](https://docs.citrix.com/en-us/netscaler/11/nitro-api.html)) for softlayer_lb_vpx_service resource management. The NITRO API is only accessable in SoftLayer private network, so it is necessary to execute terraform in SoftLayer's private network. SoftLayer [SSL VPN](http://www.softlayer.com/VPN-Access) also can be used for private network connection. 
 
```hcl
resource "softlayer_lb_vpx_service" "test_service" {
  name = "test_load_balancer_service"
  vip_id = "${softlayer_lb_vpx_vip.testacc_vip.id}"
  destination_ip_address = "${softlayer_virtual_guest.test_server.ipv4_address}"
  destination_port = 80
  weight = 55
  connection_limit = 5000
  health_check = "HTTP"
}
```

## Argument Reference

* `name` | *string*
    * (Required) The unique identifier for the VPX Load Balancer Service.
* `vip_id` | *string*
    * (Required) The ID of the VPX Load Balancer Virtual IP Address that the VPX Load Balancer Service is assigned to.
* `destination_ip_address` | *string*
    * (Required) The IP address of the server traffic will be directed to. If Netscaler VPX 10.1 is used, destination_ip_address should be a public IP address in a SoftLayer account. If Netscaler VPX 10.5 is used, any IP address can be a destination_ip_address.
* `destination_port` | *int*
    * (Required) The destination port of the server traffic will be directed to.
* `weight` | *int*
    * (Required) Set the weight of this VPX Load Balancer service. Affects the choices the VPX Load Balancer makes between the various services. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service) for details.
    * In VPX 10.5, weight value is ignored. 
* `connection_limit` | *int*
    * (Required) Set the connection limit for this service. The range is 0 ~ 4294967294. See [maxClient](https://docs.citrix.com/en-us/netscaler/11/reference/netscaler-command-reference/basic/service.html) for details.
* `health_check` | *string*
    * (Required) Set the health check for the VPX Load Balancer Service. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service) for details.

## Attributes Reference

* `id` - The VPX Load Balancer Service unique identifier.
