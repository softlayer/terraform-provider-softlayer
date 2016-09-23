# `softlayer_ssh_key`

Use this data source to get the ID or other details of an existing SSH key, for use in other resources.

## Example Usage

```hcl
data "softlayer_ssh_key" "public_key" {
    label = "Terraform Public Key"
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
* `public_key` - (Required) The public ssh key
* `notes` - (Optional) A small note about an SSH key to use at your discretion
