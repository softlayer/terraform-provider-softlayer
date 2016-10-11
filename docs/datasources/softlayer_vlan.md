# `softlayer_vlan`

Use this data source to import the details of an *existing* VLAN as a read-only data source.

## Example Usage

```hcl
data "softlayer_vlan" "vlan_foo" {
    name = "FOO"
}
```

The fields of the data source can then be referenced by other resources within the
same configuration using interpolation syntax. For example, when specifying a VLAN id
in a *softlayer_bare_metal* resource configuration,
the numeric "IDs" are often unknown. Using the above data source as an example, it would be possible to
reference the `id` property in a *softlayer_bare_metal* resource:

```hcl
resource "softlayer_bare_metal" "bm1" {
    ...
    public_vlan_id = "${data.softlayer_vlan.vlan_foo.id}"
    ...
}
```

## Argument Reference

* `name` - (Required if number nor router hostname are provided) The name of the VLAN as it was defined in SoftLayer. These names can be found from the SoftLayer portal, navigating to [Network > IP Management > VLANs](https://control.softlayer.com/network/vlans).
* `number` - (Required if name is not provided) The VLAN number as seen on the [SoftLayer portal](https://control.softlayer.com/network/vlans).
* `router_hostname` - (Required if name is not provided) The primary VLAN router hostname as seen on the [SoftLayer portal](https://control.softlayer.com/network/vlans).

## Attributes Reference

`id` is set to the ID of the image template
