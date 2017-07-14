# `softlayer_bare_metal`

Provides a `bare_metal` resource. This allows bare metals to be created, updated and deleted. `softlayer_bare_metal` resource supports both monthly bare metal servers and hourly bare metal servers.
 For more detail on bare metal seves, refer to the [link](https://www.ibm.com/cloud-computing/bluemix/bare-metal-servers)

## Hourly bare metal server
If the `softlayer_bare_metal` resource definition has a `fixed_config_preset` attribute, terraform will create an hourly
bare metal server. The following example creates a new hourly bare metal server. Hardware specifications 
are already defined in the `fixed_config_preset` attribute and cannot be modified.

### Example of an hourly bare metal server
```hcl
resource "softlayer_bare_metal" "hourly-bm1" {
    hostname = "hourly-bm1"
    domain = "example.com"
    os_reference_code = "UBUNTU_16_64"
    datacenter = "dal01"
    network_speed = 100 # Optional
    hourly_billing = true # Optional
    private_network_only = false # Optional
    fixed_config_preset = "S1270_8GB_2X1TBSATA_NORAID"
}
```

In addition, users can use configure optional attributes such as `user_metadata`, `tags`, and `notes` attributes as follows:

### Example of additional attributes for the hourly bare metal server
```hcl
resource "softlayer_bare_metal" "hourly-bm1" {
    hostname = "hourly-bm1"
    domain = "example.com"
    os_reference_code = "UBUNTU_16_64"
    datacenter = "dal01"
    network_speed = 100 # Optional
    hourly_billing = true # Optional
    private_network_only = false # Optional
    fixed_config_preset = "S1270_8GB_2X1TBSATA_NORAID"
    
    user_metadata = "{\"value\":\"newvalue\"}" # Optional
    tags = [
      "collectd",
      "mesos-master"
    ]
    notes = "note test"
}
```

## Monthly bare metal server
If the `fixed_config_preset` attribute is not configured, terraform will consider it as a monthly bare metal server resource. It provides 
options to configure process, memory, network, disk, and RAID. Users also can assign VLANs and subnets for the target monthly bare metal server. To configure the monthly bare 
metal server, you need to provide additional attributes such as `package_key_name`, `proecss_key_name`, `disk_key_names`, and `os_key_name`. The following example describes a basic configuration
 of the monthly bare metal server.

### Example of a monthly bare metal server
```hcl
resource "softlayer_bare_metal" "monthly_bm1" {
    package_key_name = "DUAL_E52600_V4_12_DRIVES"
    process_key_name = "INTEL_INTEL_XEON_E52620_V4_2_10"
    memory = 64
    os_key_name = "OS_WINDOWS_2012_R2_FULL_DC_64_BIT_2"
    hostname = "cust-bm"
    domain = "ms.com"
    datacenter = "wdc04"
    network_speed = 100
    public_bandwidth = 500
    disk_key_names = [ "HARD_DRIVE_800GB_SSD", "HARD_DRIVE_800GB_SSD", "HARD_DRIVE_800GB_SSD" ]
    hourly_billing = false
}
```

Users can configure additional options. The following example configures target VLANs, subnets, and a RAID controller. `storage_groups` 
configures RAIDs and disk partitioning.
### Example of a monthly bare metal server with additional options
```hcl
resource "softlayer_bare_metal" "monthly_bm1" {

# Mandatory attributes
    package_key_name = "DUAL_E52600_V4_12_DRIVES"
    process_key_name = "INTEL_INTEL_XEON_E52620_V4_2_10"
    memory = 64
    os_key_name = "OS_WINDOWS_2012_R2_FULL_DC_64_BIT_2"
    hostname = "cust-bm"
    domain = "ms.com"
    datacenter = "wdc04"
    network_speed = 100
    public_bandwidth = 500
    disk_key_names = [ "HARD_DRIVE_800GB_SSD", "HARD_DRIVE_800GB_SSD", "HARD_DRIVE_800GB_SSD" ]
    hourly_billing = false

# Optional attributes
    private_network_only = false
    unbonded_network = true
    user_metadata = "{\"value\":\"newvalue\"}"
    public_vlan_id = 12345678
    private_vlan_id = 87654321
    public_subnet = "50.97.46.160/28"
    private_subnet = "10.56.109.128/26"
    tags = [
      "collectd",
      "mesos-master"
    ]
    redundant_power_supply = true
    storage_groups = {
       # RAID 5
       array_type_id = 3
       # Use three disks
       hard_drives = [ 0, 1, 2]
       array_size = 1600
       # Basic partition template for windows
       partition_template_id = 17
    }
}
```

_Please Note_: Monthly bare metal servers does not support `immediate cancellation`. If the monthly bare metal server is deleted by terraform, `anniversary date cancellation` option will be used. 

## Create a bare metal server using quote ID
If users already have a quote id for the bare metal server, they can create a new bare metal server with the quote id. The following example describes a basic configuration for a bare metal server with 
 quote_id.
  
### Example of a quote based ordering
```hcl
# Create a new bare metal
resource "softlayer_bare_metal" "quote_test" {
    hostname = "quote-bm-test"
    domain = "example.com"
    quote_id = 2209349 
}
```

Users can use additional options when they create a new bare metal server with `quote_id`. The following example defines target VLANs, subnets, 
 user metadata, and tags additionally. 
 
### Example of a quote based ordering with additional options
```hcl
# Create a new bare metal
resource "softlayer_bare_metal" "quote_test" {

# Mandatory attributes
    hostname = "quote-bm-test"
    domain = "example.com"
    quote_id = 2209349

# Optional attributes
    user_metadata = "{\"value\":\"newvalue\"}"
    public_vlan_id = 12345678
    private_vlan_id = 87654321
    public_subnet = "50.97.46.160/28"
    private_subnet = "10.56.109.128/26"
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
    * **Optional**
* `domain` | *string*
    * Domain for the computing instance.
    * **Required**
* `user_metadata` | *string*
    * Arbitrary data to be made available to the computing instance.
    * *Optional*
* `notes` | *string*
    * A note of up to 1000 characters about the server.
    * *Optional*
* `ssh_key_ids` | *array* of numbers
    * SSH key _IDs_ to install on the computing instance upon provisioning.
    * *Optional*

    **Note:** Don't know the ID(s) for your SSH keys? See [here](https://github.com/softlayer/terraform-provider-softlayer/blob/master/docs/datasources/softlayer_ssh_key.md) for a way to reference your SSH keys by their labels.

* `post_install_script_uri` | *string*
    * As defined in the [SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest_SupplementalCreateObjectOptions).
    * *Optional*
* `tags` | *array* of strings
    * Set tags on this bare metal server. The characters permitted are A-Z, 0-9, whitespace, _ (underscore), - (hyphen), . (period), and : (colon). All other characters will be stripped away.
    * *Optional*

**Monthly/Hourly bare metal server attributes**

* `datacenter` | *string*
    * Specifies which datacenter the instance is to be provisioned in.
    * It is a mandatory attribute for monthly and hourly bare metal servers.
    * *Optional*
* `hourly_billing` | *boolean*
    * Specifies the billing type for the instance. When true the computing instance will be billed on hourly usage, otherwise it will be billed on a monthly basis.
    * *Default*: true
    * *Optional*
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

**Hourly bare metal server only attributes**

* `fixed_config_preset` | *string*
    * The configuration preset that the hourly bare metal server will be provisioned with. This governs the type of cpu, number of cores, amount of ram, and hard drives which the bare metal server will have. [Take a look at the available presets](https://api.softlayer.com/rest/v3/SoftLayer_Hardware/getCreateObjectOptions.json) (use your api key as the password), and find the key called _fixedConfigurationPresets_. Under that, the presets will be identified by the *keyName*s.
    * It is a mandatory attribute for hourly bare metal server provisioning.
    * *Optional*
* `os_reference_code` | *string*
    * An operating system reference code that will be used to provision the computing instance. 
    * Hourly bare metal server : [Get a complete list of the os reference codes available for hourly bare metal servers](https://api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest_Block_Device_Template_Group/getVhdImportSoftwareDescriptions.json?objectMask=referenceCode) (use your api key as the password).
    * *Optional*
    * **Conflicts with** `image_template_id`.
    
**Monthly / Quote based bare metal server provisioning attributes**

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

**Monthly bare metal server only attributes**

* `package_key_name` | *string*
    * Monthly bare metal server's package key name. This attribute is only used when a new monthly bare metal server is created.
    * You can find available key names in the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/getAllObjects?objectFilter={"type":{"keyName":{"operation":"BARE_METAL_CPU"}}}). You need your username and api_key to access to the page. 
    * *Optional*
* `process_key_name` | *string*
    * Monthly bare metal server's process key name. This attribute is only used when a new monthly bare metal server is created.
    * Note the package key ID from the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/getAllObjects?objectFilter={"type":{"keyName":{"operation":"BARE_METAL_CPU"}}}) and replace **PACKAGE_ID** in the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/PACKAGE_ID/getItems?objectMask=mask[prices[id,categories[id,name,categoryCode],capacityRestrictionType,capacityRestrictionMinimum,capacityRestrictionMaximum,locationGroupId]]) to your package key ID. Select a process key name from available process key names.
    * *Optional*
* `disk_key_names` | *list*
    * Array of internal disk key names. This attribute is only used when a new monthly bare metal server is created.
    * Note the package key ID from the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/getAllObjects?objectFilter={"type":{"keyName":{"operation":"BARE_METAL_CPU"}}}) and replace **PACKAGE_ID** in the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/PACKAGE_ID/getItems?objectMask=mask[prices[id,categories[id,name,categoryCode],capacityRestrictionType,capacityRestrictionMinimum,capacityRestrictionMaximum,locationGroupId]]) to your package key ID. Select a disk key name from available disk key names.
    * *Optional*
* `os_key_name` | *string*
    * An operating system key name that will be used to provision the computing instance. 
    * Note the package key ID from the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/getAllObjects?objectFilter={"type":{"keyName":{"operation":"BARE_METAL_CPU"}}}) and replace **PACKAGE_ID** in the [link](https://api.softlayer.com/rest/v3/SoftLayer_Product_Package/PACKAGE_ID/getItems?objectMask=mask[prices[id,categories[id,name,categoryCode],capacityRestrictionType,capacityRestrictionMinimum,capacityRestrictionMaximum,locationGroupId]]) to your package key ID. Select a OS key name from available OS key names.
    * *Optional*
* `redundant_network` | *boolean*
    * If `redundant_network` is `true`, two physical network interfaces will be provided with a bonding configuration. 
    * *Default*: False
    * *Optional*
* `unbonded_network` | *boolean*
    * If `unbonded_network` is `true`, two physical network interfaces will be provided.
    * *Default*: False
    * *Optional*
* `public_bandwidth` | *int*
    * Allowed public network traffic(GB) per month. 
    * `public_bandwidth` can be greater than 0 when `private_network_only` is `false` and the server is a monthly based server.
    * *Optional*
* `memory` | *int*
    * An amount of memory(GB) for the server.
    * *Optional*
* `storage_groups` | *array of storage group objects*
    * RAID and partition configuration. Refer to the [link](https://sldn.softlayer.com/blog/hansKristian/Ordering-RAID-through-API) to configure `storage_groups`.
    * *Optional*
    * Each storage group object has the following sub-attributes:
    * `array_type_id` | *int*
    * It provides RAID type. You can find `array_type_id` from the [link](https://api.softlayer.com/rest/v3/SoftLayer_Configuration_Storage_Group_Array_Type/getAllObjects). 
    * *Required*
    * `hard_drives` | *array of int*
    * Index of hard drives for RAID configuration. The index starts from 0. For example, if you want to use first two hard drives, you will use the following expression: [0,1]
    * *Required*
    * `array_size` | *int*
    * Target RAID disk size in GB unit. 
    * *Optional*
    * `partition_template_id` | *int*
    * Partition template id for OS disk. The templates are different based on the target OS. Check your OS with the [link](https://api.softlayer.com/rest/v3/SoftLayer_Hardware_Component_Partition_OperatingSystem/getAllObjects ). Note the id of the OS and 
    check available partition templates using the URL : https://api.softlayer.com/rest/v3/SoftLayer_Hardware_Component_Partition_OperatingSystem/OS_ID/getPartitionTemplates . Replace `OS_ID` to your OS ID from the URL and find your template id.  
    * *Optional*
    
* `redundant_power_supply` | *boolean*
    * If `redundant_power_supply` is true, an additional power supply will be provided. 
    * *Optional*
* `tcp_monitoring` | *boolean*
    * If `tcp_monitoring` is `false`, ping monitoring service will be provided. If `tcp_monitoring` is `true`, ping and tcp monitoring service will be provided.
    * *Optional*
**Quote based provisioning only attributes**

* `quote_id` | *int*
    * Create a bare metal server using the quote. 
    * If quote_id is defined, the terraform uses specifications in the quote to create a bare metal server.
    * You can find the quote id by navigating on the portal to _Account > Sales > Quotes_.
    * *Optional*

## Attributes Reference

The following attributes are exported:

* `id` - id of the bare metal.
* `public_ipv4_address` - Public IPv4 address of the bare metal server.
* `private_ipv4_address` - Private IPv4 address of the bare metal server.