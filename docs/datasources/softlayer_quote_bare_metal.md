# `softlayer_quote_bare_metal`

Use this data source to import the name of an *existing* custom bare metal quote as a read-only data source.

## Example Usage

```hcl
data softlayer_quote_bare_metal quote_test{
  name = "quote_test"
}
```

It imports the quote of the custom bare metal server and shows detailed attributes.


## Argument Reference

`name` - (Required) The name of the quote, as it was defined in SoftLayer

## Attributes Reference

`id` - Set to the ID of the quote.
`datacenter` - It specifies which datacenter the instance is to be provisioned in.
`os_reference_code` - Target OS key name.
`network_speed` - Specifies the connection speed (in Mbps) for the instance's network components.
`private_network_only` - Specifies whether or not the instance only has access to the private network.
`package_key_name` - Custom bare metal server's package key name.
`process_key_name` - Custom bare metal server's process key name.
`disk_key_names` - Array of internal disk key names. 
`redundant_network` - If `redundant_network` is `true`, two physical network interfaces will be provided with a bonding configuration. 
`unbonded_network` - If `unbonded_network` is `true`, two physical network interfaces will be provided.
`public_bandwidth` - Allowed public network traffic(GB) per month. 
`memory` - An amount of memory(GB) for the server.
`storage_groups` - RAID and partition configuration.
`redundant_power_supply`- If `redundant_power_supply` is true, an additional power supply will be provided. 
`tcp_monitoring` - If `tcp_monitoring` is `true`, ping and tcp monitoring service will be provided.