# `softlayer_dns_domain`

The `softlayer_dns_domain` resource represents a single DNS domain managed on SoftLayer. Domains contain general information about the domain name such as name and serial. Individual records such as `A`, `AAAA`, `CTYPE`, and `MX` records are stored in the domain's associated resource records using the `softlayer_dns_domain_record` resource.

```hcl
resource "softlayer_dns_domain" "dns-domain-test" {
    name = "dns-domain-test.com"
    target = "127.0.0.10"
}
```

## Argument Reference

The following arguments are supported:

* `name` | *string* - (Required) A domain's name including top-level domain, for example "example.com". When the domain is created, proper `NS` and `SOA`  records are created automatically for it.
* `target`|*string* - (Required) The primary target IP address that the domain will resolve to. Upon creation, an `A` record will be created with a host value of `@` and a data-target value of the IP address provided which will be associated to the new domain.

## Attributes Reference

The following attributes are exported

* `id` - A domain record's internal identifier.
* `serial` - A unique number denoting the latest revision of a domain.
* `update_date` - The date that this domain record was last updated.
