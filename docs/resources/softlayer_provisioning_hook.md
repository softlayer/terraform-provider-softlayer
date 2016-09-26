# `softlayer_provisioning_hook`

Provides Provisioning Hooks containing all the information needed to add a hook into a server/Virtual provision and os reload. This allows Provisioning Hooks to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Provisioning_Hook).

## Example Usage

```hcl
resource "softlayer_provisioning_hook" "test_provisioning_hook" {
    name = "test_provisioning_hook_name"
    uri = "https://raw.githubusercontent.com/test/slvm/master/test-script.sh"
}
```

## Argument Reference

The following arguments are supported:

* `name` | *string* - (Required) A descriptive name used to identify a provisioning hook.
* `uri` | *string* - (Required) The endpoint that the script will be downloaded/downloaded and executed from .

Fields `name` and `uri` are editable.

## Attributes Reference

The following attributes are exported:

* `id` - id of the new provisioning hook
