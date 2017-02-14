# `softlayer_block_storage`

Provides a `softlayer_block_storage` resource. This allows Iscsi-based [Endurance](https://knowledgelayer.softlayer.com/topic/endurance-storage)
 and [Performance](https://knowledgelayer.softlayer.com/topic/performance-storage) block storage to be created, updated and deleted.
Block storage is accessed and mounted through a Multipath I/O (MPIO) Internet Small Computer System Interface (iSCSI) connection.
Procedure to access block storage is described [here](https://knowledgelayer.softlayer.com/procedure/block-storage-linux) and [here](https://knowledgelayer.softlayer.com/procedure/accessing-block-storage-microsoft-windows) for linux and windows respectively.

## Example Usage

```hcl
# Create 20G endurance block storage with 10G snapshot capacity and 0.25 IOPS/GB option.
resource "softlayer_block_storage" "test1" {
        type = "Endurance"
        datacenter = "dal05"
        capacity = 20
        iops = 0.25
        os_format_type = "Linux"

        # Optional fields
        allowed_virtual_guest_ids = [ 27699397 ]
        allowed_ip_addresses = ["10.40.98.193", "10.40.98.200"]
        snapshot_capacity = 10
}

# Create 20G performance block storage and 100 IOPS option.
resource "softlayer_block_storage" "test2" {
        type = "Performance"
        datacenter = "dal05"
        capacity = 20
        iops = 100
        os_format_type = "Linux"

        # Optional fields
        allowed_virtual_guest_ids = [ 27699397 ]
        allowed_ip_addresses = ["10.40.98.193", "10.40.98.200"]
}
```

## Argument Reference

The following arguments are supported:

* `type` | *string*
    * Specifies the type of the storage. Accepted values are `Endurance` and `Performance`.
    * **Required**
* `datacenter` | *string*
    * Specifies which datacenter the instance is to be provisioned in.
    * **Required**
* `capacity` | *int*
    * The amount of storage capacity to allocate in gigabytes.
    * **Required**
* `iops` | *float*
    * Specifies IOPS value for the storage. Please find available values for endurance storage in the [link](https://knowledgelayer.softlayer.com/learning/introduction-endurance-storage).
    * **Required**
* `os_format_type` | *string*
    * Specifies which OS Type to be used when formatting the storage space. This should match the OS type that will be connecting to the LUN.
    * **Required**
* `snapshot_capacity` | *int*
    * The amount of snapshot capacity to allocate in gigabytes. Only `Endurance` storage supports snapshot.
    * **Optional**
* `allowed_virtual_guest_ids` | *array of int*
    * Specifies allowed virtual guests. Virtual guests should be in the same data center.
    * **Optional**
* `allowed_hardware_ids` | *array of int*
    * Specifies allowed baremetal servers. Baremetal servers should be in the same data center.
    * **Optional**    
* `allowed_ip_addresses` | *array of string*
    * Specifies allowed IP addresses. IP addresses should be in the same data center.
    * **Optional**    


## Attributes Reference

The following attributes are exported:

* `id` - id of the storage.
* `hostname` - The fully qualified domain name of the storage.
* `volumename` - The name of the storage volume.
* `allowed_virtual_guest_info` - Contains username, password and hostIQN of the virtual guests with access to the storage.
* `allowed_hardware_info` - Contains username, password and hostIQN of the bare metal servers with access to the storage.
