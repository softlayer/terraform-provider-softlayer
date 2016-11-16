# `softlayer_lb_vpx_vip`

Create, update, and delete Softlayer VPX Load Balancer Virtual IP Addresses. For additional details please refer to the [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_VirtualIpAddress).

_Please Note_: If Netscaler VPX 10.5 is used, terraform uses Netscaler's REST API([NITRO API](https://docs.citrix.com/en-us/netscaler/11/nitro-api.html)) for softlayer_lb_vpx_vip resource management. The NITRO API is only accessable in SoftLayer private network, so it is necessary to execute terraform in SoftLayer's private network when you deploy Netscaler VPX 10.5 devices. SoftLayer [SSL VPN](http://www.softlayer.com/VPN-Access) also can be used for private network connection. 

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

Netscaler VPX 10.5 supports SSL Offload. If `type` is "SSL" and `security_certificate_id` is configured, the `virtual_ip_address` provides `HTTPS` protocol. The following examples describes SSL-offload configuration.
```hcl
# Create a new Netscaler VPX 10.5
resource "softlayer_lb_vpx" "test" {
    datacenter = "lon02"
    speed = 10
    version = "10.5"
    plan = "Standard"
    ip_count = 2
}

resource "softlayer_lb_vpx_vip" "test_vip1" {
    name = "test_vip1"
    nad_controller_id = "${softlayer_lb_vpx.test.id}"
    load_balancing_method = "rr"
    source_port = 443
# SSL type provides SSL offload
    type = "SSL"
    virtual_ip_address = "${softlayer_lb_vpx.test.vip_pool[0]}"
# Use a security certificated in SoftLayer portal
    security_certificate_id = 80347
}

resource "softlayer_lb_vpx_service" "testacc_service1" {
  name = "test_load_balancer_service1"
  vip_id = "${softlayer_lb_vpx_vip.test_vip1.id}"
# 10.6.218.166 should provides HTTP service with port 80
  destination_ip_address = "10.66.218.166"
  destination_port = 80
  weight = 100
  connection_limit = 4294967294
  health_check = "ICMP"
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
    * (Required) The connection type for the VPX Load Balancer Virtual IP Address. Accepted values are *HTTP*, *FTP*, *TCP*, *UDP*, *DNS*, and *SSL*. If *SSL* is configured, `security_certificate_id` will be used as a certification for SSL offload services.
* `security_certificate_id` | *int*
    * (Netscaler VPX 10.5 only)
    * (Optional) Provides a security certification for SSL offload. For additional information, refer to [softlayer_security_certificate](./softlayer_security_certificate.md)

## Attributes Reference

* `id` - The VPX Load Balancer Virtual IPs unique identifier.
