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
   image_id = "****-****-****-****-****"
}
```

## Argument Reference

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
* `os_reference_code` | *string*
    * An operating system reference code that will be used to provision the computing instance.
    * **Conflicts with** `image_id`.
* `image_id` | *string*
    * A global identifier for the image template to be used to provision the computing instance.
    * **Conflicts with** `os_reference_code`.
* `network_speed` | *int*
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
* `ssh_keys` | *array* of numbers
    * SSH keys to install on the computing instance upon provisioning.
    * *Default*: nil
    * *Optional*
    * *Conflicts with* `ssh_key_labels`
* `ssh_key_labels` | *array* of strings
    * SSH key labels to install on the computing instance upon provisioning. **Be warned** that if duplicate ssh key labels exist, the code will use the first one it finds that matches.
    * *Optional*
    * *Conflicts with* `ssh_keys`
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

## Attributes Reference

The following attributes are exported:

* `id` - id of the virtual guest.
