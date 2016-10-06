# `softlayer_bare_metal`

Provides a `bare_metal` resource. This allows bare metals to be created, updated and deleted.

```hcl
# Create a new bare metal
resource "softlayer_bare_metal" "twc_terraform_sample" {
    hostname = "twc-terraform-sample-name"
    domain = "bar.example.com"
    os_reference_code = "UBUNTU_16_64"
    datacenter = "dal01"
    network_speed = 100 # Optional
    hourly_billing = true # Optional
    private_network_only = false # Optional
    user_metadata = "{\"value\":\"newvalue\"}" # Optional
    public_vlan_id = 12345678 # Optional
    private_vlan_id = 87654321 # Optional
    public_subnet = "50.97.46.160/28" # Optional
    private_subnet = "10.56.109.128/26" # Optional
    fixed_config_preset = "S1270_8GB_2X1TBSATA_NORAID"
    image_template_id = 12345 # Optional
    tags = [
      "collectd",
      "mesos-master"
    ]
}
```

## Argument Reference

The following arguments are supported:

* `hostname` | *string*
    * Hostname for the computing instance.
    * **Required**
* `domain` | *string*
    * Domain for the computing instance.
    * **Required**
* `datacenter` | *string*
    * Specifies which datacenter the instance is to be provisioned in.
    * **Required**
* `fixed_config_preset` | *string*
    * The configuration preset that the bare metal server will be provisioned with. This governs the type of cpu, number of cores, amount of ram, and hard drives which the bare metal server will have. [Take a look at the available presets](https://api.softlayer.com/rest/v3/SoftLayer_Hardware/getCreateObjectOptions.json) (use your api key as the password), and find the key called _fixedConfigurationPresets_. Under that, the presets will be identified by the *keyName*s.
    * **Required**
* `hourly_billing` | *boolean*
    * Specifies the billing type for the instance. When true the computing instance will be billed on hourly usage, otherwise it will be billed on a monthly basis.
    * *Default*: true
    * *Optional*
* `os_reference_code` | *string*
    * An operating system reference code that will be used to provision the computing instance. [Get a complete list of the os reference codes available](https://api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest_Block_Device_Template_Group/getVhdImportSoftwareDescriptions.json?objectMask=referenceCode) (use your api key as the password).
    * *Optional*
    * **Conflicts with** `image_template_id`.
* `image_template_id` | *int*
    * The image template id to be used to provision the computing instance. Note this is not the global identifier (uuid), but the image template group id that should point to a valid global identifier. You can get the image template id by navigating on the portal to _Devices > Manage > Images_, clicking on the desired image, and taking note of the id number in the browser URL location.
    * *Optional*
    * **Conflicts with** `os_reference_code`.

    **Note:** Don't know the ID(s) for your image templates? [You can reference them by name, too](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_image_template.md).

* `network_speed` | *int*
    * Specifies the connection speed (in Mbps) for the instance's network components.
    * *Default*: 100
    * *Optional*
* `private_network_only` | *boolean*
    * Specifies whether or not the instance only has access to the private network. When true this flag specifies that a compute instance is to only have access to the private network.
    * *Default*: False
    * *Optional*
* `public_vlan_id` | *int*
    * Public VLAN which is to be used for the public network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans). Click on the desired VLAN and note the id number in the URL.
    * *Optional*
* `private_vlan_id` | *int*
    * Private VLAN which is to be used for the private network interface of the instance. Accepted values can be found [here](https://control.softlayer.com/network/vlans). Click on the desired VLAN and note the id number in the URL.
    * *Optional*
* `public_subnet` | *string*
    * Public subnet which is to be used for the public network interface of the instance. Accepted values are primary public networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Optional*
* `private_subnet` | *string*
    * Private subnet which is to be used for the private network interface of the instance. Accepted values are primary private networks and can be found [here](https://control.softlayer.com/network/subnets).
    * *Optional*
* `user_metadata` | *string*
    * Arbitrary data to be made available to the computing instance.
    * *Optional*
* `ssh_key_ids` | *array* of numbers
    * SSH key _IDs_ to install on the computing instance upon provisioning.
    * *Optional*

    **Note:** Don't know the ID(s) for your SSH keys? See [here](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_ssh_key.md) for a way to reference your SSH keys by their labels.

* `post_install_script_uri` | *string*
    * As defined in the [SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions).
    * *Optional*
*   `tags` | *array* of strings
    * Set tags on this bare metal server. The characters permitted are A-Z, 0-9, whitespace, _ (underscore), - (hypen), . (period), and : (colon). All other characters will be stripped away.
    * *Optional*

## Attributes Reference

The following attributes are exported:

* `id` - id of the bare metal.
* `public_ipv4_address` - Public IPv4 address of the bare metal server.
* `private_ipv4_address` - Private IPv4 address of the bare metal server.
