# `softlayer_dns_secondary`

The `softlayer_dns_secondary` resource represents a single secondary DNS zone managed on SoftLayer. Each record created within the secondary DNS service defines which zone is transferred, what server it is transferred from, and the frequency that zone transfers occur at. Zone transfers are performed automatically based on the transfer frequency set on the secondary DNS record.

```hcl
resource "softlayer_dns_secondary" "dns-secondary-zone-1" {
    zoneName = "secondary-zone1.com"
    transferFrequency = 10
    masterIpAddress = "172.16.0.1"
}
```

## Argument Reference

The following arguments are supported:

* `zoneName` | *string* - (Required) The name of the zone that is transferred.
* `transferFrequency` | *int* - (Required) Signifies how often a secondary DNS zone should be transferred in minutes.
* `masterIpAddress` | *string* - (Required) The IP address of the master name server where a secondary DNS zone is transferred from.

## Attributes Reference

The following attributes are exported

* `id` - A secondary zone's internal identifier.
* `statusId` - The current status of a secondary DNS record.
* `statusText` - The textual representation of a secondary DNS zone's status.
