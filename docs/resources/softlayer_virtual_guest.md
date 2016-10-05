# `softlayer_virtual_guest`

Provides a `virtual_guest` resource. This allows virtual guests to be created, updated and deleted.

```hcl
# Create a new virtual guest using image "Debian"
resource "softlayer_virtual_guest" "twc_terraform_sample" {
    name = "twc-terraform-sample-name"
    domain = "bar.example.com"
    os_reference_code = "DEBIAN_7_64"
    datacenter = "wdc01"
    network_speed = 10
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
# Create a new virtual guest using block device template and tags
resource "softlayer_virtual_guest" "terraform-sample-BDTGroup" {
   name = "terraform-sample-blockDeviceTemplateGroup"
   domain = "bar.example.com"
   datacenter = "ams01"
   public_network_speed = 10
   hourly_billing = false
   cpu = 1
   ram = 1024
   local_disk = false
   image_id = 12345
   tags = [
     "collectd"
     "mesos-master"
   ]
}
```

## Argument Reference

The following arguments are supported:

*   `name` | *string*
    * Hostname for the computing instance.
    * **Required**
*   `domain` | *string*
    * Domain for the computing instance.
    * **Required**
*   `cpu` | *int*
    * The number of CPU cores to allocate.
    * **Required**
*   `ram` | *int*
    * The amount of memory to allocate in megabytes.
    * **Required**
*   `datacenter` | *string*
    * Specifies which datacenter the instance is to be provisioned in.
    * **Required**
*   `hourly_billing` | *boolean*
    * Specifies the billing type for the instance. When true the computing instance will be billed on hourly usage, otherwise it will be billed on a monthly basis.
    * *Default*: true
    * *Optional*
*   `local_disk` | *boolean*
    * Specifies the disk type for the instance. When true the disks for the computing instance will be provisioned on the host which it runs, otherwise SAN disks will be provisioned.
    * *Default*: false
    * *Optional*
*   `dedicated_acct_host_only` | *boolean*
    * Specifies whether or not the instance must only run on hosts with instances from the same account
    * *Default*: false
    * *Optional*
*   `os_reference_code` | *string*
    * An operating system reference code that will be used to provision the computing instance. [Get a complete list of the os reference codes available](https://api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest_Block_Device_Template_Group/getVhdImportSoftwareDescriptions.json?objectMask=referenceCode) (use your api key as the password).
    * **Conflicts with** `image_id`.
*   `image_id` | *int*
    * The image template id to be used to provision the computing instance. Note this is not the global identifier (uuid), but the image template group id that should point to a valid global identifier. You can get the image template id by navigating on the portal to _Devices > Manage > Images_, clicking on the desired image, and taking note of the id number in the browser URL location.
    * **Conflicts with** `os_reference_code`.

    **Note:** Don't know the ID(s) for your image templates? [You can reference them by name, too](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_image_template.md).

*   `network_speed` | *int*
    * Specifies the connection speed (in Mbps) for the instance's network components.
    * *Default*: 100
    * *Optional*
*   `private_network_only` | *boolean*
    * Specifies whether or not the instance only has access to the private network. When true this flag specifies that a compute instance is to only have access to the private network.
    * *Default*: False
    * *Optional*
*   `front_end_vlan` | *map*
    * Public VLAN which is to be used for the public network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
    * *Optional*
*   `back_end_vlan` | *map*
    * Private VLAN which is to be used for the private network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans).
    * *Optional*
*   `front_end_subnet` | *string*
    * Public subnet which is to be used for the public network interface of the instance. Accepted values are primary public networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Optional*
*   `back_end_subnet` | *string*
    * Private subnet which is to be used for the private network interface of the instance. Accepted values are primary private networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Optional*
*   `disks` | *array* of numeric disk sizes (in GBs).
    * Block device and disk image settings for the computing instance
    * *Optional*
    * *Default*: The smallest available capacity for the primary disk will be used. If an image template is specified the disk capacity will be be provided by the template.
*   `user_data` | *string*
    * Arbitrary data to be made available to the computing instance.
    * *Optional*
*   `ssh_keys` | *array* of numbers
    * SSH key _IDs_ to install on the computing instance upon provisioning.
    * *Optional*

    **Note:** Don't know the ID(s) for your SSH keys? See [here](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_ssh_key.md) for a way to reference your SSH keys by their labels.

*   `post_install_script_uri` | *string*
    * As defined in the [SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions).
    * *Optional*
*   `tags` | *array* of strings
    * Set tags on this virtual guest. The characters permitted are A-Z, 0-9, whitespace, _ (underscore), - (hypen), . (period), and : (colon). All other characters will be stripped away.
    * *Optional*

## Attributes Reference

The following attributes are exported:

* `id` - id of the virtual guest.
* `ipv4_address` - Public IPv4 address of the virtual guest.
* `ip_address_id_private` - Unique ID for the private ID address assigned to the virtual_guest.
* `ipv4_address_private` - Private IPv4 address of the virtual guest.
* `ip_address_id` - Unique ID for the public ID address assigned to the virtual_guest.
