# softlayer_fw_hardware_dedicated_rules

Represents rules for Dedicated Hardware Firewall resources in SoftLayer. 
Only one `softlayer_fw_hardware_dedicated_rules` resource is allowed per 
firewall. One `softlayer_fw_hardware_dedicated_rules` can have multiple 
rules. For additional details please refer to
[Configure a Hardware Firewall (Dedicated)](https://knowledgelayer.softlayer.com/procedure/configure-hardware-firewall-dedicated).

_Please Note_: Target VLAN should have at least one subnet for rule 
configuration. To express _ANY IP addresses_ in external side, configure 
`src_ip_address` as `0.0.0.0` and `src_ip_cidr` as `0`. To express _API 
IP addresses_ in internal side, configure `dst_ip_address` as `any` and 
`src_ip_cidr` as `32`. Once `softlayer_fw_hardware_dedicated_rules` resource 
is created, it cannot be deleted. SoftLayer doesnot allow entire rule deleting. 
Firewalls should have at least one rule. If terraform destroys 
`softlayer_fw_hardware_dedicated_rules` resources, _permit from any to any
 with TCP, UDP, ICMP, GRE, PPTP, ESP, and HA_ rules will be configured. 

```hcl
resource "softlayer_fw_hardware_dedicated" "demofw" {
  ha_enabled = false
  public_vlan_id = 1234567
}

resource "softlayer_fw_hardware_dedicated_rules" "rules" {
 firewall_id = "${softlayer_fw_hardware_dedicated.demofw.id}"
 rules = {
      "action" = "permit"
      "src_ip_address"= "10.1.1.0"
      "src_ip_cidr"= 24
      "dst_ip_address"= "any"
      "dst_ip_cidr"= 32
      "dst_port_range_start"= 80
      "dst_port_range_end"= 80
      "notes"= "Permit from 10.1.1.0"
      "protocol"= "udp"
 }
  rules = {
       "action" = "deny"
       "src_ip_address"= "10.1.1.0"
       "src_ip_cidr"= 24
       "dst_ip_address"= "any"
       "dst_ip_cidr"= 32
       "dst_port_range_start"= 81
       "dst_port_range_end"= 81
       "notes"= "Permit from 10.1.1.0"
       "protocol"= "udp"
  }
}
```

## Argument Reference

The following arguments are supported:

* `firewall_id` | *int*
    * Target Hardware Firewall (Dedicated) device id.
    * **Required**
* `rules` | *array*
    * Represents firewall rules. At least one `rules` should be defined.
    * **Required**
* `rules.action` | *string*
    * "permit" or "deny" traffic matching this rule.
    * **Required**
* `rules.src_ip_address` | *string*
    * Can be either specific ip address or the network address for a specific subnet.
    * **Required**
* `rules.src_ip_cidr` | *string*
    * Indicates the standard CIDR notation for the selected source.  "32"
     will implement the rule for a single IP while, for example, "24" will
      implement the rule for 256 IPs.
    * **Required**
* `rules.dst_ip_address` | *string*
    * Can be either 'any' or a specific ip address or the network address for a specific subnet.
    * **Required**
* `rules.dst_ip_cidr` | *string*
    *  Indicates the standard CIDR notation for the selected destination.
    * **Required**
* `rules.dst_port_range_start` | *string*
    * The range of ports for TCP and UDP. 1~65535 values are allowed. 
    * **Optional**
* `rules.dst_port_range_end` | *string*
    * The range of ports for TCP and UDP. 1~65535 values are allowed. 
    * **Optional**
* `rules.notes` | *string*
    * Comments for the rule.
    * **Optional**
* `rules.protocol` | *string*
    * Protocol for the rule. _tcp/udp/icmp/gre/pptp/ah/esp_ are allowed. 
    * **Required**
    