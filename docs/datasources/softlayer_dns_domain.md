# `softlayer_dns_domain`

Use this data source to import the name of an *existing* domain as a read-only data source.

## Example Usage

```hcl
data "softlayer_dns_domain" "domain_id" {
    name = "test-domain.com"
}
```

The fields of the data source can then be referenced by other resources within the same configuration using
interpolation syntax. For example, when specifying domain in softlayer_dns_domain_record resource configuration,
the numeric "IDs" are often unknown. Using the above data source as an example, it would be possible to
reference the `id` property in a softlayer_dns_domain_record resource:

```hcl
resource "softlayer_dns_domain_record" "www" {
    ...
    domain_id = "${data.softlayer_dns_domain.domain_id.id}"
    ...
}
```

## Argument Reference

`name` - (Required) The name of the domain, as it was defined in SoftLayer

## Attributes Reference

`id` is set to the ID of the domain.
