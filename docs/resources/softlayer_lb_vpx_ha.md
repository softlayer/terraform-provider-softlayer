# `softlayer_lb_vpx_ha`

Configure a High Availability Pair with two netscaler VPX devices. Two netscaler VPXs should have a version 10.5 and be located in the same subnet. A primary netscaler VPX provides load balancing services in active mode and a secondary netscaler VPX provies the services when the primary netscaler VPX fails. For additional details, please refer to the [CTX116748](https://support.citrix.com/article/CTX116748) and [Knowledgelayer](http://knowledgelayer.softlayer.com/articles/netscaler-vpx-10-high-availability-setup). 

_Please Note_: `softlayer_lb_vpx_ha` only supports Netscaler VPX 10.5 and [NITRO API](https://docs.citrix.com/en-us/netscaler/11/nitro-api.html) is used to configure HA. NITRO API is only accessable in SoftLayer private network, so it is necessary to execute terraform in SoftLayer's private network. SoftLayer [SSL VPN](http://www.softlayer.com/VPN-Access) also can be used for private network connection. 

_Please Note_: Two netscaler VPXs use same password in a high availabiliy mode. Once `softlayer_lb_vpx_ha` resource is created, terraform changes a password of the secondary netscaler VPX to a password of the primary netscaler VPX. If `softlayer_lb_vpx_ha` resource is destroyed, terraform will restore the password of the seconary netscaler VPX. 
```hcl
# Create a primary netscaler VPX
resource "softlayer_lb_vpx" "test_pri" {
    datacenter = "lon02"
    speed = 10
    version = "10.5"
    plan = "Standard"
    ip_count = 2
}

# Create a secondary netscaler VPX in same subnets.
resource "softlayer_lb_vpx" "test_sec" {
    datacenter = "lon02"
    speed = 10
    version = "10.5"
    plan = "Standard"
    ip_count = 2
    public_vlan_id = "${softlayer_lb_vpx.test_pri.public_vlan_id}"
    private_vlan_id = "${softlayer_lb_vpx.test_pri.private_vlan_id}"
    public_subnet = "${softlayer_lb_vpx.test_pri.public_subnet}"
    private_subnet = "${softlayer_lb_vpx.test_pri.private_subnet}"
}

# Configure a High Availability with the primary netscaler VPX and the secondary netscaler VPX
resource "softlayer_lb_vpx_ha" "test_ha" {
    primary_id = "${softlayer_lb_vpx.test_pri.id}"
    secondary_id = "${softlayer_lb_vpx.test_sec.id}"
    stay_secondary = false
}
```

## Argument Reference

* `primary_id` | *string*
    * (Required) The unique identifier of the primary netscaler VPX.
* `secondary_id` | *string*
    * (Required) The unique identifier of the secondary netscaler VPX.
* `stay_secondary` | *bool*
    * (Optional) If stay_secondary is _true_, the primary netscaler VPX will not take over the service to the secondary netscaler VPX even if the primary netscaler VPX fails. For additional details, please refer to [Link1](https://docs.citrix.com/en-us/netscaler/10-5/ns-system-wrapper-10-con/ns-nw-ha-intro-wrppr-con/ns-nw-ha-frcng-scndry-nd-sty-scndry-tsk.html) and [Link2](https://support.citrix.com/article/CTX116748). The default value is _false_.

## Attributes Reference

* `id` - The unique identifier of _softlayer_lb_vpx_ha_
