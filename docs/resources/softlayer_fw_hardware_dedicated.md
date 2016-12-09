# softlayer_fw_hardware_dedicated

Represents Dedicated Hardware Firewall resources in SoftLayer. One firewall 
protects one public VLAN and provides in-bound network packet filtering services. 
You can order or find firewalls in `Gateway/Firewall` column of `Network > IP Management > VLANs` menu. 
For additional details please refer to
[Configure a Hardware Firewall (Dedicated)](https://knowledgelayer.softlayer.com/procedure/configure-hardware-firewall-dedicated).

```hcl
resource "softlayer_fw_hardware_dedicated" "testfw" {
  ha_enabled = false
  public_vlan_id = 12345678
}
```

## Argument Reference

The following arguments are supported:

* `ha_enabled` | *boolean*
    * Set if the local load balancer needs to be HA enabled or not.
    * **Required**
* `public_vlan_id` | *int*
    * Target public VLAN ID which will be protected by the firewall. Accepted values can be found [here](https://control.softlayer.com/network/vlans).  Click on the desired VLAN and note the ID on the resulting URL. Or, you can also [refer to a VLAN by name using a data source](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_vlan.md).
    * **Required**
