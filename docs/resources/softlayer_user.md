# softlayer_user

Represents the SoftLayer's user login resource. You can get, create,
update and delete this resource. For additional details please refer to
[SoftLayer API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_User_Customer).

Also see additional notes below.

```hcl
resource "softlayer_user" "joe" {
    address1     = "12345 Any Street"
    address2     = "Suite #99"
    city         = "Atlanta"
    company_name = "Comp Inc"
    country      = "US"
    email        = "joe@doe.com"
    first_name   = "Joe"
    has_api_key  = false
    last_name    = "Doe"
    password     = "Change3Me!"
    permissions  = [
        "ACCESS_ALL_GUEST",
        "ACCESS_ALL_HARDWARE",
        "SERVER_ADD",
        "SERVER_CANCEL",
        "RESET_PORTAL_PASSWORD"
    ]
    state        = "GA"
    timezone     = "EST"
}
```

## Argument Reference

The following arguments are supported:

* `address1` | *string*
    * User's street address first line.
    * **Required**
* `address2` | *string*
    * User's street address second line.
    * *Default*: ""
    * *Optional*
* `city` | *string*
    * User's street address city.
    * **Required**
* `company_name` | *string*
    * User's company name.
    * **Required**
* `country` | *string*
    * User's street address country.
    * **Required**
* `email` | *string*
    * User's email address associated with this login userid.
    * **Required**
* `first_name` | *string*
    * User's first name.
    * **Required**
* `has_api_key` | *boolean*
    * This flag when true specifies that a new SoftLayer API key
      be created for this user. They key is returned back in the
      `api_key` computed attribute.
    * *Default*: False
    * *Optional* - When false, it will delete any api key that was
      previously created.
    * **Required**
* `last_name` | *string*
    * User's last name.
    * **Required**
* `password` | *string*
    * Initial password for this new user login. This string value must
      conform to SoftLayer's portal password to avoid failures. You can
      find the password policies in your SoftLayer portal profile page.
      At the time of this writing, valid passwords must be 8 to 20 characters
      in length with a combination of UPPER and lower case characters, at
      least one number, and at least one of the following special
      characters: `_-|@.,?/!~#$%^&*(){}[]=`. The password specified here
      is 'hashed' and 'encoded' before it is stored in the Terraform
      state file.
    * **Required**
* `permissions` | *array of strings*
    * Permissions assigned to this user. This is a set of zero or more
      string values. See [SoftLayer_User_Customer_CustomerPermission_Permission](http://sldn.softlayer.com/reference/datatypes/SoftLayer_User_Customer_CustomerPermission_Permission).
    * *Default*: []
    * *Optional*
* `state` | *string*
    * User's street address state.
    * **Required**
* `timezone` | *string*
    * User's timezone (shortname, e.g., "EST")
      Value is one of [SoftLayer_Locale_Timezone](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Locale_Timezone).
    * **Required**
* `user_status` | *string*
    * User's login status. Value is one of
      [SoftLayer_User_Customer_Status](http://sldn.softlayer.com/reference/datatypes/SoftLayer_User_Customer_Status).
    * *Optional*
    * *Default*: "ACTIVE"

## Attributes Reference

The following computed attributes are returned:

* `api_key` | *string*
    * SoftLayer API key that was created for this new user.
    * *Computed*
* `id` | *string*
    * Unique SoftLayer id for this new user.
    * *Computed*
* `username` | *string*
    * A name that uniquely identifies a user globally across all SoftLayer
      logins. It is also the login userid. Once a user login is created,
      it cannot be changed.
    * *Computed*

## Additional notes

In SoftLayer, when user logins are deleted, there is a delay when that
login actually gets deleted in the SoftLayer backend systems. SoftLayer
successfully acknowledges the delete request and immediately updates the
user status to CANCEL_PENDING. Actual deletion of happens at some
unspecified amount of time in the future. This delay may be significant
especially during your projects testing phase. If you create a new user
login, and then delete it, and then create it again, you may receive an
error, as SoftLayer backend has not completely processed the previous delete
operation. If you do want to run through this create-delete-create-again
cycle again, you will have to specify a new globally unique username value
in your subsequent requests.
