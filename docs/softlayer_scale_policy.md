# `softlayer_scale_policy`

Provides a `scale_policy` resource. This allows scale policies for auto scale groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Scale_Policy).

## Example Usage

```hcl
# Create a new scale policy
resource "softlayer_scale_policy" "test_scale_policy" {
    name = "test_scale_policy_name"
    scale_type = "RELATIVE"         # or ABSOLUTE, PERCENT
    scale_amount = 1
    cooldown = 30                   # if not provided, the scale_group cooldown applies
    scale_group_id = "${softlayer_scale_group.sample-http-cluster.id}"
    triggers = {
        type = "RESOURCE_USE"
        watches = {
                    metric = "host.cpu.percent"
                    operator = ">"
                    value = "80"
                    period = 120
        }
    }
    triggers = {
        type = "ONE_TIME"
        date = "2016-07-30T23:55:00-00:00"
    }
    triggers = {
        type = "REPEATING"
        schedule = "0 1 ? * MON,WED *"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` | *string*
    * Name of the scale policy.
    * **Required**
* `scale_type` | *string*
    * Set the scale type for the scale policy. Accepted values are ABSOLUTE, RELATIVE and PERCENT
    * **Required**
* `scale_amount` | *int*
    * A count of the scale actions to perform upon any trigger hit.
    * **Required**
* `cooldown` | *int*
    * The number of seconds this policy will wait after last action date on group before performing another action.
    * **Optional**
* `scale_group_id` | *int*
    * Specifies the id of the scale group this policy is on.
    * **Required**
* `triggers` | *array of map of ints and strings*
    * The triggers to check for this group.
    * **Optional**

## Attributes Reference

The following attributes are exported:

* `id` - id of the scale policy.
