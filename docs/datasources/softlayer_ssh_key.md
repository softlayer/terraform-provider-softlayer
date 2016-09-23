# `softlayer_ssh_key`

Use this data source to import the details of an *existing* SSH key as a read-only data source.

## Example Usage

```hcl
data "softlayer_ssh_key" "public_key" {
    label = "Terraform Public Key"
}
```

The fields of the data source can then be referenced by other resources within the same configuration using
interpolation syntax. For example, when specifying SSH keys in a softlayer_virtual_guest resource configuration,
the numeric "IDs" are often unknown. Using the above data source as an example, it would be possible to
reference the `id` property in a softlayer_virtual_guest resource:

```hcl
resource "softlayer_virtual_guest" "vm1" {
    ...
    ssh_keys = ["${data.softlayer_ssh_key.public_key.id}"]
    ...
}
```

## Argument Reference

* `label` - (Required) The label of the SSH key, as it was defined in SoftLayer
* `most_recent` - (Optional) If more than SSH key matches the label, use the most recent key

NOTE: If more or less than a single match is returned by the search, Terraform will fail.
Ensure that your search is specific enough to return a single SSH key only,
or use most_recent to choose the most recent one.

## Attributes Reference

`id` is set to the ID of the SSH key.  In addition, the following attributes are exported:

* `fingerprint` - sequence of bytes to authenticate or lookup a longer ssh key
* `public_key` - the public key contents
* `notes` - notes stored with the SSH key
