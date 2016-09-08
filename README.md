# terraform-provider-softlayer

## Install

```
$ go get github.com/softlayer/terraform-provider-softlayer
```

Create or edit this file to specify the location of the terraform softlayer provider binary:

```
# ~/.terraformrc
providers {
    softlayer = "/path/to/bin/terraform-provider-softlayer"
}
```

## Documentation

### Provider

Here is an example that will setup the following:
+ An SSH key resource.
+ A virtual server resource that uses an existing SSH key.
+ A virtual server resource using an existing SSH key and a Terraform managed SSH key (created as "test_key_1" in the example below).

(create this as sl.tf and run terraform commands from this directory):

```hcl
provider "softlayer" {
    username = ""
    api_key = ""
}

# This will create a new SSH key that will show up under the \
# Devices>Manage>SSH Keys in the SoftLayer console.
resource "softlayer_ssh_key" "testkey1" {
    name = "testkey1"
    public_key = "${file(\"~/.ssh/id_rsa_test_key_1.pub\")}"
    # Windows Example:
    # public_key = "${file(\"C:\ssh\keys\path\id_rsa_test_key_1.pub\")}"
}

# Virtual Server created with existing SSH Key already in SoftLayer \
# inventory and not created using this Terraform template.
resource "softlayer_virtual_guest" "host-a" {
    name = "host-a.example.com"
    domain = "example.com"
    ssh_keys = ["123456"]
    image = "DEBIAN_7_64"
    datacenter = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
}

# Virtual Server created with a mix of previously existing and \
# Terraform created/managed resources.
resource "softlayer_virtual_guest" "host-b" {
    name = "host-b.example.com"
    domain = "example.com"
    ssh_keys = ["123456", "${softlayer_ssh_key.test_key_1.id}"]
    image = "CENTOS_6_64"
    datacenter = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
}
```

You'll need to provide your SoftLayer username and API key,
so that Terraform can connect. If you don't want to put
credentials in your configuration file, you can leave them
out:

```
provider "softlayer" {}
```

...and instead set these environment variables:

- **SOFTLAYER_USERNAME**: Your SoftLayer username
- **SOFTLAYER_API_KEY**: Your API key

You can also put credentials in _~/.softlayer_. See the [softlayer api python client docs](http://softlayer-python.readthedocs.io/en/latest/config_file.html) for details on this configuration file.

### Resources

#### `softlayer_virtual_guest`

Provides a `virtual_guest` resource. This allows virtual guests to be created, updated and deleted.

```hcl
# Create a new virtual guest using image "Debian"
resource "softlayer_virtual_guest" "twc_terraform_sample" {
    name = "twc-terraform-sample-name"
    domain = "bar.example.com"
    image = "DEBIAN_7_64"
    datacenter = "wdc01"
    public_network_speed = 10
    hourly_billing = true
    private_network_only = false
    cpu = 1
    ram = 1024
    disks = [25, 10, 20]
    user_data = "{\"value\":\"newvalue\"}"
    dedicated_acct_host_only = true
    local_disk = false
    front_end_vlan {
       vlan_number = 1144
       primary_router_hostname = "fcr03a.wdc01"
    }
    back_end_vlan {
       vlan_number = 978
       primary_router_hostname = "bcr03a.wdc01"
    }
    front_end_subnet = "50.97.46.160/28"
    back_end_subnet = "10.56.109.128/26"
}
```

```hcl
# Create a new virtual guest using block device template
resource "softlayer_virtual_guest" "terraform-sample-BDTGroup" {
   name = "terraform-sample-blockDeviceTemplateGroup"
   domain = "bar.example.com"
   datacenter = "ams01"
   public_network_speed = 10
   hourly_billing = false
   cpu = 1
   ram = 1024
   local_disk = false
   block_device_template_group_gid = "****-****-****-****-****"
}
```

##### Argument Reference

The following arguments are supported:

* `name` | *string*
    * Hostname for the computing instance.
    * **Required**
* `domain` | *string*
    * Domain for the computing instance.
    * **Required**
* `cpu` | *int*
    * The number of CPU cores to allocate.
    * **Required**
* `ram` | *int*
    * The amount of memory to allocate in megabytes.
    * **Required**
* `datacenter` | *string*
    * Specifies which datacenter the instance is to be provisioned in.
    * **Required**
* `hourly_billing` | *boolean*
    * Specifies the billing type for the instance. When true the computing instance will be billed on hourly usage, otherwise it will be billed on a monthly basis.
    * **Required**
* `local_disk` | *boolean*
    * Specifies the disk type for the instance. When true the disks for the computing instance will be provisioned on the host which it runs, otherwise SAN disks will be provisioned.
    * **Required**
* `dedicated_acct_host_only` | *boolean*
    * Specifies whether or not the instance must only run on hosts with instances from the same account
    * *Default*: nil
    * *Optional*
* `image` | *string*
    * An identifier for the operating system to provision the computing instance with.
    * **Conditionally required**    - Disallowed when blockDeviceTemplateGroup.globalIdentifier is provided, as the template will specify the operating system.
* `block_device_template_group_gid` | *string*
    * A global identifier for the template to be used to provision the computing instance.
    * **Conditionally required**    - Disallowed when operatingSystemReferenceCode is provided, as the template will specify the operating system.
* `public_network_speed` | *int*
    * Specifies the connection speed for the instance's network components.
    * *Default*: 10
    * *Optional*
* `private_network_only` | *boolean*
    * Specifies whether or not the instance only has access to the private network. When true this flag specifies that a compute instance is to only have access to the private network.
    * *Default*: False
    * *Optional*
* `front_end_vlan` | *map*
    * Public VLAN which is to be used for the public network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
    * *Default*: nil
    * *Optional*
* `back_end_vlan` | *map*
    * Private VLAN which is to be used for the private network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
    * *Default*: nil
    * *Optional*
* `front_end_subnet` | *string*
    * Public subnet which is to be used for the public network interface of the instance. Accepted values are primary public networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Default*: nil
    * *Optional*
* `back_end_subnet` | *string*
    * Public subnet which is to be used for the private network interface of the instance. Accepted values are primary private networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Default*: nil
    * *Optional*
* `disks` | *array*
    * Block device and disk image settings for the computing instance
    * *Optional*
    * *Default*: The smallest available capacity for the primary disk will be used. If an image template is specified the disk capacity will be be provided by the template.
* `user_data` | *string*
    * Arbitrary data to be made available to the computing instance.
    * *Default*: nil
    * *Optional*
* `ssh_keys` | *array*
    * SSH keys to install on the computing instance upon provisioning.
    * *Default*: nil
    * *Optional*
* `ipv4_address` | *string*
    * Uses editObject call, template data [defined here](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest).
    * *Default*: nil
    * *Optional*
* `ipv4_address_private` | *string*
    * Uses editObject call, template data [defined here](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest).
    * *Default*: nil
    * *Optional*
* `post_install_script_uri` | *string*
    * As defined in the [SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions).
    * *Default*: nil
    * *Optional*

##### Attributes Reference

The following attributes are exported:

* `id` - id of the virtual guest.

#### `softlayer_user`

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
    timezone     = 114
    user_status  = 1001
    username     = "joedoe"
}
```

##### Argument Reference

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
* `timezone` | *int*
    * User's timezone id (no validation checks with the street address).
      Value is one of [SoftLayer_Locale_Timezone](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Locale_Timezone).
    * **Required**
* `user_status` | *int*
    * Status id of user's login status. Value is one of
      [SoftLayer_User_Customer_Status](http://sldn.softlayer.com/reference/datatypes/SoftLayer_User_Customer_Status).
    * *Optional*
    * *Default*: 1001 // means 'active'
* `username` | *string*
    * A name that uniquely identifies a user globally across all SoftLayer
      logins. It is also the login userid. Once a user login is created,
      it cannot be changed.
    * **Required**

All fields except `username` are editable.

##### Attributes Reference

The following computed attributes are returned:

* `api_key` | *string*
    * SoftLayer API key that was created for this new user.
    * *Computed*
* `id` | *string*
    * Unique SoftLayer id for this new user.
    * *Computed*


##### Additional notes

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

#### `softlayer_ssh_key`

Provides SSK keys. This allows SSH keys to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Security_Ssh_Key).

##### Example Usage

```hcl
resource "softlayer_ssh_key" "test_ssh_key" {
    name = "test_ssh_key_name"
    notes = "test_ssh_key_notes"
    public_key = "ssh-rsa <rsa_public_key>"
}
```

##### Argument Reference

The following arguments are supported:

* `name` - (Required) A descriptive name used to identify a ssh key.
* `public_key` - (Required) The public ssh key.
* `notes` - (Optional) A small note about a ssh key to use at your discretion.

Fields `name` and `notes` are editable.

##### Attributes Reference

The following attributes are exported:

* `id` - id of the new ssh key
* `fingerprint` - sequence of bytes to authenticate or lookup a longer ssh key.

#### `softlayer_security_certificate`

Create, update, and destroy [SoftLayer Security Certificates](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Security_Certificate).

**Using certs on file:**

```hcl
resource "softlayer_security_certificate" "test_cert" {
  certificate = "${file("cert.pem")}"
  private_key = "${file("key.pem")}"
}
```

**Example with cert in-line:**

```hcl
resource "softlayer_security_certificate" "test_cert" {
    certificate = <<EOF
[......] # cert contents
-----END CERTIFICATE-----
    EOF

    private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
[......] # cert contents
-----END RSA PRIVATE KEY-----
    EOF
}
```

##### Argument Reference

* `certificate` | *string*
    * (Required) The certificate provided publicly to clients requesting identity credentials.
* `intermediate_certificate` | *string*
    * (Optional) The intermediate certificate authorities certificate that completes the certificate chain for the issued certificate. Required when clients will only trust the root certificate.
* `private_key` | *string*
    * (Required) The private key in the key/certificate pair.

##### Attributes Reference

* `common_name` - The common name (usually a domain name) encoded within the certificate.
* `create_date` - The date the certificate record was created.
* `id` - The ID of the certificate record.
* `key_size` - The size (number of bits) of the public key represented by the certificate.
* `modify_date` - The date the certificate record was last modified.
* `organization_name` - The organizational name encoded in the certificate.
* `validity_begin` - The UTC timestamp representing the beginning of the certificate's validity.
* `validity_days` - The number of days remaining in the validity period for the certificate.
* `validity_end` - The UTC timestamp representing the end of the certificate's validity period.

#### `softlayer_objectstorage_account`

**Note:** For managing SoftLayer object storage *containers* and *objects*, please see the [Swift provider](/docs/providers/swift/index.html), since SoftLayer's object storage is an implementation of Swift object storage.

Ensures there is an existing object storage account within your SoftLayer account. If there is an existing object storage, it will learn its account name and keep it as its ID for future usage. If there is no object storage account, it will order one for you and remember the account name. It is not meant to be used for managing the life cycle of an object storage account in SoftLayer (e.g. update, delete) at this time.

```hcl
resource "softlayer_objectstorage_account" "foo" {
}
```

##### Argument Reference

No additional arguments needed.

##### Computed Fields

* `id` - The object storage account name, which you can later use with [Swift resources](/docs/providers/swift/index.html).

#### `softlayer_dns_domain`

The `softLayer_dns_domain` data type represents a single DNS domain record hosted on the SoftLayer nameservers. Domains contain general information about the domain name such as name and serial. Individual records such as `A`, `AAAA`, `CTYPE`, and `MX` records are stored in the domain's associated resource records using the  [`softlayer_dns_domain_record`](/docs/providers/softlayer/r/dns_records.html) resource.

```hcl
resource "softlayer_dns_domain" "dns-domain-test" {
    name = "dns-domain-test.com"
}
```

##### Argument Reference

The following arguments are supported:

* `name` | *string*
     * (Required) A domain's name including top-level domain, for example "example.com". _Name_ is the only field that needs to be set for `softlayer_dns_domain`. During creation the `NS` and `SOA` resource records are created automatically.

##### Attributes Reference

The following attributes are exported

* `id` - A domain record's internal identifier.
* `serial` - A unique number denoting the latest revision of a domain.
* `update_date` - The date that this domain record was last updated.

#### `softlayer_dns_domain_record`

The `softlayer_dns_domain_record` data type represents a single resource record entry in a [`softlayer_dns_domain`](/docs/providers/softlayer/r/dns.html). Each resource record contains a `host` and `record_data` property, defining a resource's name and it's target data.

We are using [SoftLayer_Dns_Domain_ResourceRecord](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord)
SL's object for most of CRUD operations. Only for SRV record type we are using [SoftLayer_Dns_Domain_ResourceRecord_SrvType](https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord_SrvType) SL's object.

Currently we can CRUD almost all record types except _SOA_ type which is initially created on DNS create action.

##### `A` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_AType)

```hcl
resource "softlayer_dns_domain" "main" {
    name = "main.example.com"
}

resource "softlayer_dns_domain_record" "www" {
    record_data = "123.123.123.123"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "www.example.com"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "a"
}
```

##### `AAAA` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_AaaaType)

```hcl
resource "softlayer_dns_domain_record" "aaaa" {
    record_data = "FE80:0000:0000:0000:0202:B3FF:FE1E:8329"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "www.example.com"
    contact_email = "user@softlayer.com"
    ttl = 1000
    record_type = "aaaa"
}
```

##### `CNAME` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_CnameType)

```hcl
resource "softlayer_dns_domain_record" "cname" {
    record_data = "real-host.example.com."
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "alias.example.com"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "cname"
}
```

##### `MX` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_MxType)

```hcl
resource "softlayer_dns_domain_record" "recordMX-1" {
    record_data = "mail-1"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "@"
    mx_priority = "10"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "mx"
}
```

##### `NS` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_NsType)

```hcl
resource "softlayer_dns_domain_record" "recordNS" {
    record_data = "ns1.example.org"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "@"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "ns"
}
```

##### `SPF` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_SpfType)

```hcl
resource "softlayer_dns_domain_record" "recordSPF" {
    record_data = "v=spf1 mx:mail.example.org ~all"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "mail-1"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "spf"
}
```

##### `TXT` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_TxtType/)

```hcl
resource "softlayer_dns_domain_record" "recordTXT" {
    record_data = "host"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "A SPF test host"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "txt"
}
```

##### `SRV` Record | [SLDN](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_SrvType)

```hcl
resource "softlayer_dns_domain_record" "recordSRV" {
    record_data = "ns1.example.org"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "hosta-srv.com"
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "srv"
    port = 8080
    priority = 3
    protocol = "_tcp"
    weight = 3
    service = "_mail"
}
```

##### `PTR` Record
######  _A note on creating `PTR` records:_

There are a lot of things that make the `PTR` record work properly, please review the [SLDN documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Dns_Domain_ResourceRecord_PtrType/) regarding how they are to be implemented.

```hcl
resource "softlayer_dns_domain_record" "recordPTR" {
    record_data = "ptr.example.com"
    domain_id = "${softlayer_dns_domain.main.id}"
    host = "45"  # <- this is the last octet of IPAddress in the range of the subnet
    contact_email = "user@softlayer.com"
    ttl = 900
    record_type = "ptr"
}
```

##### Argument Reference

* `record_data` | *string*
    * (Required) The value of a domain's resource record. This can be an IP address or a hostname. Fully qualified host and domain name data must end with the "." character.
* `domain_id` | *int*
    * (Required) An identifier belonging to the domain that a resource record is associated with.
* `expire` | *int*
    * The amount of time in seconds that a secondary name server (or servers) will hold a zone before it is no longer considered authoritative.
* `host` | *string*
    * (Required) The host defined by a resource record. A value of `"@"` denotes a wildcard.
* `minimum_ttl` | *int*
    * The amount of time in seconds that a domain's resource records are valid. This is also known as a minimum TTL, and can be overridden by an individual resource record's TTL.
* `mx_priority` | *int*
    * Useful in cases where a domain has more than one mail exchanger, the priority property is the priority of the MTA that delivers mail for a domain. A lower number denotes a higher priority, and mail will attempt to deliver through that MTA before moving to lower priority mail servers. Priority is defaulted to 10 upon resource record creation.
* `refresh` | *int*
    * The amount of time in seconds that a secondary name server should wait to check for a new copy of a DNS zone from the domain's primary name server. If a zone file has changed then the secondary DNS server will update it's copy of the zone to match the primary DNS server's zone.
* `contact_email` | *string*
    * (Required) The email address of the person responsible for a domain, with the "@" replaced with a `.`. For instance, if root@example.org is responsible for example.org, then example.org's SOA responsibility is `root.example.org.`.
* `retry` | *int*
    * The amount of time in seconds that a domain's primary name server (or servers) should wait if an attempt to refresh by a secondary name server failed before attempting to refresh a domain's zone with that secondary name server again.
* `ttl` | *int*
    * (Required) The Time To Live value of a resource record, measured in seconds. TTL is used by a name server to determine how long to cache a resource record. An SOA record's TTL value defines the domain's overall TTL.
* `record_type` | *string* - (Required) A domain resource record's type, valid types are:
    * `a` for address records
    * `aaaa` for address records
    * `cname` for canonical name records
    * `mx` for mail exchanger records
    * `ns` for name server records
    * `ptr` for pointer records in reverse domains
    * `soa` for a domain's start of authority record
    * `spf` for sender policy framework records
    * `srv` for service records
* `txt` | *string*
    * for text records
* `service` | *string*
    * The symbolic name of the desired service
* `protocol` | *string*
    * The protocol of the desired service; this is usually either TCP or UDP.
* `port` | *int*
    * The TCP or UDP port on which the service is to be found.
* `priority` | *int*
    * The priority of the target host, lower value means more preferred.
* `weight` | *int*
    * A relative weight for records with the same priority.

##### Attributes Reference

* `id` - A domain resource record's internal identifier.

#### `softlayer_lb_vpx`

Create, update, and destroy SoftLayer VPX Load Balancers.

_Please Note_: SoftLayer VPX Load Balancer consists of Citrix Netscaler VPX devices (virtual), these are currently priced on a per-month basis, so please use caution when creating the resource as the cost for an entire month is incurred immediately upon creation. For more information on pricing please see this [link](http://www.softlayer.com/network-appliances), under the Citrix log, click "see more pricing" for a current price matrix.

You can also use this REST URL to get a listing of VPX choices along with version numbers, speed and plan type:

```
https://{{userName}}:{{apiKey}}@api.softlayer.com/rest/v3/SoftLayer_Product_Package/192/getItems.json?objectMask=id;capacity;description;units;keyName;prices.id;prices.categories.id;prices.categories.name
```

[SLDN reference](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Application_Delivery_Controller)

```hcl
resource "softlayer_lb_vpx" "test_vpx" {
    datacenter = "dal06"
    speed = 10
    version = "10.1"
    plan = "Standard"
    ip_count = 2
    front_end_vlan {
      vlan_number = 1251
      primary_router_hostname = "fcr01a.dal06"
    }
    back_end_vlan {
       vlan_number = 1540
       primary_router_hostname = "bcr01a.dal06"
    }
    front_end_subnet = "23.246.226.248/29"
    back_end_subnet = "10.107.180.0/26"
}
```

##### Argument Reference

* `datacenter` | *string*
    * (Required) Specifies which datacenter the VPX Load Balancer is to be provisioned in. Accepted values can be found [here](http://www.softlayer.com/data-centers).
* `speed` | *int*
    * (Required) The speed in Mbps. Accepted values are `10`, `200`, and `1000`.
* `version` | *string*
    * (Required) The VPX Load Balancer version. Accepted values are `10.1` and `10.5`.
* `plan` | *string*
    * (Required) The VPX Load Balancer plan. Accepted values are `Standard` and `Platinum`.
* `ip_count` | *int*
    * (Required) The number of static public IP addresses assigned to the VPX Load Balancer. Accepted values are `2`, `4`, `8`, and `16`.
* `front_end_vlan` | *map*
    * (Optional) Public VLAN which is to be used for the public network interface of the VPX Load Balancer. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
* `back_end_vlan` | *map*
    * (Optional) Private VLAN which is to be used for the private network interface of the VPX Load Balancer. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
* `front_end_subnet` | *string*
    * (Optional) Public subnet which is to be used for the public network interface of the VPX Load Balancer. Accepted values are primary public networks and can be found [here](https://control.softlayer.com/network/subnets).
* `back_end_subnet` | *string*
    * (Optional) Public subnet which is to be used for the private network interface of the VPX Load Balancer. Accepted values are primary private networks and can be found [here](https://control.softlayer.com/network/subnets).

##### Attributes Reference

* `id` - A VPX Load Balancer's internal identifier.
* `name` - A VPX Load Balancer's internal name.
* `vip_pool` - List of virtual ip addresses for the VPX Load Balancer.

#### `softlayer_lb_vpx_service`

Create, update, and delete Softlayer VPX Load Balancer Services. For additional details please refer to the [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service).

```hcl
resource "softlayer_lb_vpx_service" "test_service" {
  name = "test_load_balancer_service"
  vip_id = "${softlayer_lb_vpx_vip.testacc_vip.id}"
  destination_ip_address = "${softlayer_virtual_guest.terraform-acceptance-test-2.ipv4_address}"
  destination_port = 89
  weight = 55
  connection_limit = 5000
  health_check = "HTTP"
}
```

##### Argument Reference

* `name` | *string*
    * (Required) The unique identifier for the VPX Load Balancer Service.
* `vip_id` | *string*
    * (Required) The ID of the VPX Load Balancer Virtual IP Address that the VPX Load Balancer Service is assigned to.
* `destination_ip_address` | *string*
    * (Required) The IP address of the server traffic will be directed to.
* `destination_port` | *int*
    * (Required) The destination port of the server traffic will be directed to.
* `weight` | *int*
    * (Required) Set the weight of this VPX Load Balancer service. Affects the choices the VPX Load Balancer makes between the various services. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service) for details.
* `connection_limit` | *int*
    * (Required) Set the connection limit for this service.
* `health_check` | *string*
    * (Required) Set the health check for the VPX Load Balancer Service. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_Service) for details.

##### Attributes Reference

* `id` - The VPX Load Balancer Service unique identifier.

#### `softlayer_lb_vpx_vip`

Create, update, and delete Softlayer VPX Load Balancer Virtual IP Addresses. For additional details please refer to the [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_VirtualIpAddress).

```hcl
resource "softlayer_lb_vpx_vip" "testacc_vip" {
    name = "test_load_balancer_vip"
    nad_controller_id = "${softlayer_lb_vpx.testacc_foobar_vpx.id}"
    load_balancing_method = "lc"
    source_port = 80
    virtual_ip_address = "${softlayer_virtual_guest.terraform-acceptance-test-1.ipv4_address}"
    type = "HTTP"
}
```

##### Argument Reference

* `name` | *string*
    * (Required) The unique identifier for the VPX Load Balancer Virtual IP Address.
* `nad_controller_id` | *int*
    * (Required) The ID of the VPX Load Balancer that the VPX Load Balancer Virtual IP Address will be assigned to.
* `load_balancing_method` | *string*
    * (Required) The VPX Load Balancer load balancing method. See [the documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_LoadBalancer_VirtualIpAddress) for details.
* `virtual_ip_address` | *string*
    * (Required) The public facing IP address for the VPX Load Balancer Virtual IP.
* `source_port` | *int*
    * (Required) The source port for the VPX Load Balancer Virtual IP Address.
* `type` | *string*
    * (Required) The connection type for the VPX Load Balancer Virtual IP Address. Accepted values are `HTTP`, `FTP`, `TCP`, `UDP`, and `DNS`.
* `security_certificate_id` | *int*
    * (Optional) The id of the Security Certificate to be used when SSL is enabled.

##### Attributes Reference

* `id` - The VPX Load Balancer Virtual IPs unique identifier.
* `connection_limit` - The sum of the connection limit values of the VPX Load Balancer Services associated with this VPX Load Balancer Virtual IP Address.
* `modify_date` - The most recent time that the VPX Load Balancer Virtual IP Address was modified.

#### `softlayer_provisioning_hook`

Provides Provisioning Hooks containing all the information needed to add a hook into a server/Virtual provision and os reload. This allows Provisioning Hooks to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Provisioning_Hook).

##### Example Usage

```hcl
resource "softlayer_provisioning_hook" "test_provisioning_hook" {
    name = "test_provisioning_hook_name"
    uri = "https://raw.githubusercontent.com/test/slvm/master/test-script.sh"
}
```

##### Argument Reference

The following arguments are supported:

* `name` | *string* - (Required) A descriptive name used to identify a provisioning hook.
* `uri` | *string* - (Required) The endpoint that the script will be downloaded/downloaded and executed from .

Fields `name` and `uri` are editable.

##### Attributes Reference

The following attributes are exported:

* `id` - id of the new provisioning hook

#### `softlayer_scale_group`

Provides a `scale_group` resource. This allows auto scale groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Scale_Group).

##### Example Usage

```hcl
# Create a new scale group using image "Debian"
resource "softlayer_scale_group" "test_scale_group" {
    name = "test_scale_group_name"
    regional_group = "as-sgp-central-1"
    minimum_member_count = 1
    maximum_member_count = 10
    cooldown = 30
    termination_policy = "CLOSEST_TO_NEXT_CHARGE"
    virtual_server_id = 267513
    port = 8080
    health_check = {
      type = "HTTP"
    }
    virtual_guest_member_template = {
      name = "test_virtual_guest_name"
      domain = "example.com"
      cpu = 1
      ram = 4096
      public_network_speed = 1000
      hourly_billing = true
      block_device_template_group_gid = "07beadaa-1e11-476e-a188-3f7795feb9fb"
      image = "DEBIAN_7_64"
      # Optional Fields for virtual guest template (SL defaults apply):
      local_disk = false
      disks = [25,100]
      region = "sng01"
      post_install_script_uri = ""
      ssh_keys = [383111]
      user_data = "#!/bin/bash ..."
    }
    # Optional Fields for scale_group:
    network_vlans = {
        vlan_number = "1928"
        primary_router_hostname = "bcr02a.sng01"
    }
}
```

##### Argument Reference

The following arguments are supported:

* `name` | *string*
    * Name of the scale group.
    * **Required**
* `regional_group` | *string*
    * Regional group for the scale group.
    * **Required**
* `minimum_member_count` | *int*
    * The fewest number of virtual guest members allowed in the scale group.
    * **Required**
* `maximum_member_count` | *int*
    * The greatest number of virtual guest members that are allowed in the scale group.
    * **Required**
* `cooldown` | *int*
    * Specifies the number of seconds this group will wait before performing another action.
    * **Required**
* `termination_policy` | *string*
    * Specifies the termination policy for the scaling group.
    * **Required**
* `virtual_server_id` | *int*
    * Specifies the id of a virtual server .
    * **Required**
* `port` | *int*
    * Specifies the port number. For example 8080
    * **Required**
* `health_check` | *map*
    * Specifies the type of health check. For example HTTP. Also used to specify custom HTTP methods.
    * **Required**  
* `virtual_guest_member_template` | *array*
    * This is the template to create guest memebers with.
    * **Required**    
* `network_vlans` | *array of map of strings*
    * Collection of VLANs for this auto scale group.
    * *Default*: nil
    * *Optional*

##### Attributes Reference

The following attributes are exported:

* `id` - id of the scale group.

#### `softlayer_scale_policy`

Provides a `scale_policy` resource. This allows scale policies for auto scale groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Scale_Policy).

##### Example Usage

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

##### Argument Reference

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

##### Attributes Reference

The following attributes are exported:

* `id` - id of the scale policy.
 
#### `softlayer_lb_local`

Provides a `lb_local` resource. This allows local load balancers to be created, updated and deleted.

##### Example Usage

```hcl
# Create a new local load balancer
resource "softlayer_lb_local" "test_lb_local" {
    connections = 1500
    datacenter = "tok02"         
    ha_enabled = false
}
```

##### Argument Reference

The following arguments are supported:

* `connections` | *int*
    * Set the number of connections for the local load balancer.
    * **Required**
* `datacenter` | *string*
    * Set the data center for the local load balancer.
    * **Required** 
* `ha_enabled` | *boolean*
    * Set if the local load balancer needs to be HA enabled or not.
    * **Required**
* `security_certificate_id` | *int*
    * Set the Id of the security certificate associated with the local load balancer.
    * **Optional**

##### Attributes Reference

The following attributes are exported:

* `id` - id of the local load balancer.
* `ip_address` - The IP Address of the local load balancer.
* `subnet_id` - The Id of the subnet associated with the local load balancer.

#### `softlayer_lb_local_service_group`

Provides a `lb_local_service_group` resource. This allows local load balancer service groups to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service_Group).

##### Example Usage

```hcl
# Create a new local load balancer service group
resource "softlayer_lb_local_service_group" "test_service_group" {
    port = 82
    routing_method = "CONSISTENT_HASH_IP"
    routing_type = "HTTP"
    load_balancer_id = "${softlayer_lb_local.test_lb_local.id}"
    allocation = 100
}
```

##### Argument Reference

The following arguments are supported:

* `load_balancer_id` | *int*
    * Set the Id of the local load balancer.
    * **Required**
* `allocation` | *int*
    * Set the allocation field for the load balancer service group.
    * **Required** 
* `port` | *int*
    * Set the port for the local load balancer service group.
    * **Required**
* `routing_method` | *string*
    * Set the routing method for the load balancer group. For example CONSISTENT_HASH_IP
    * **Required**
* `routing_type` | *string*
    * Set the routing type for the group.
    * **Required**

##### Attributes Reference

The following attributes are exported:

* `virtual_server_id` - id of the virtual server.
* `service_group_id` - id of the load balancer service group.

#### `softlayer_lb_local_service`

Provides a `lb_local_service` resource. This allows local load balancer service to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service).

##### Example Usage

```hcl
# Create a new local load balancer service 
resource "softlayer_lb_local_service" "test_lb_local_service" {
    port = 80
    enabled = true
    service_group_id = "${softlayer_lb_local_service_group.test_service_group.service_group_id}"
    weight = 1
    health_check_type = "DNS"
    ip_address_id = "${softlayer_virtual_guest.test_server.ip_address_id}"
}

```

##### Argument Reference

The following arguments are supported:

* `service_group_id` | *int*
    * Set the Id of the local load balancer service group.
    * **Required**
* `ip_address_id` | *int*
    * Set the Id of the virtual server.
    * **Required** 
* `port` | *int*
    * Set the port for the local load balancer service.
    * **Required**
* `enabled` | *boolean*
    * Set the enabled field for the load balancer service. Accepted values are true, false.
    * **Required**
* `health_check_type` | *string*
    * Set the health check type for the load balancer service.
    * **Required**
* `weight` | *int*
    * Set the weight for the load balancer service.
    * **Required** 

##### Attributes Reference

#### `softlayer_basic_monitor`

Provides a `basic_monitor` resource. This allows basic monitors to be created, updated and deleted.
For additional details please refer to [API documentation](http://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Monitor_Version1_Query_Host).

##### Example Usage

```hcl
# Create a new basic monitor
resource "softlayer_basic_monitor" "test_basic_monitor" {
    guest_id = ${softlayer_virtual_guest.test_server.id}
    ip_address = ${softlayer_virtual_guest.test_server.id.ipv4_address}
    query_type_id = 1
    response_action_id = 1
    wait_cycles = 5
    notified_users = [460547]
}
```

##### Argument Reference

The following arguments are supported:

* `guest_id` | *int*
    * Set the Id of the virtual guest being monitored.
    * **Required**
* `ip_address` | *string*
    * Set the ip address to be monitored.
    * **Optional** 
* `query_type_id` | *int*
    * Set the id of the query type.
    * **Required**
* `response_action_id` | *int*
    * Set the id of the response action to take when the monitor fails. Accepted values are 1,2
    * **Required**
* `wait_cycles` | *int*
    * Set the number of 5-minute cycles to wait before the response action is taken.
    * **Optional**
* `notified_users` | *array of ints*
    * Set the list of user id's to be notified.
    * **Optional**    

##### Attributes Reference

The following attributes are exported:

* `id` - id of the basic monitor.
* `notified_users` - the list of user id's to be notified.

## Development

### Setup

You should have the correct source in your _$GOPATH_ for both terraform and softlayer-go.

To get _softlayer-go_:

```
go get github.com/softlayer/softlayer-go
```

To get _terraform_:

```
go get github.com/hashicorp/terraform
```

### Build

```
make bin
```

### Test

```
make test
```

### Updating dependencies

We are using [govendor](https://github.com/kardianos/govendor) to manage dependencies just like Terraform. Please see its documentation for additional help.
