# `softlayer_image_template`

Use this data source to import the details of an *existing* image template as a read-only data source.

## Example Usage

```hcl
data "softlayer_image_template" "img_tpl" {
    name = "jumpbox"
}
```

The fields of the data source can then be referenced by other resources within the
same configuration using interpolation syntax. For example, when specifying an image
template id in a softlayer_virtual_guest resource configuration,
the numeric "IDs" are often unknown. Using the above data source as an example, it would be possible to
reference the `id` property in a softlayer_virtual_guest resource:

```hcl
resource "softlayer_virtual_guest" "vm1" {
    ...
    image_id = "${data.softlayer_image_template.img_tpl.id}"
    ...
}
```

## Argument Reference

* `name` - (Required) The name of the image template as it was defined in SoftLayer. These names can be found from the SoftLayer portal, navigating to _Devices > Manage > Images_.

## Attributes Reference

`id` is set to the ID of the image template.  In addition, the following attributes are exported:

* `global_id` - The global identifier for the image template.
