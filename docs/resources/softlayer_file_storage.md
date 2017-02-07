# `softlayer_file_storage`

Provides a `softlayer_file_storage` resource. This allows NFS-based [Endurance](https://knowledgelayer.softlayer.com/topic/endurance-storage)
 and [Performance](https://knowledgelayer.softlayer.com/topic/performance-storage) file storage to be created, updated and deleted.
File storage is mounted using the NFS protocol. For example, if the `hostname` of the file storage resource is `nfsdal0501a.service.softlayer.com`
 and the `volumename` is` IBM01SV278685_7`, `nfsdal0501a.service.softlayer.com:\IBM01SV278685_7` will be a mount point. [Links](https://knowledgelayer.softlayer.com/procedure/accessing-file-storage-linux) describes nfs 
 configuration of Linux systems. For additional details, please refer to [Knowledgelayer](https://knowledgelayer.softlayer.com/topic/file-storage) and [Introduction](http://www.softlayer.com/file-storage).

## Example Usage

```hcl
# Create 20G endurance file storage with 10G snapshot capacity and 0.25 IOPS/GB option.
resource "softlayer_file_storage" "test1" {
        type = "Endurance"
        datacenter = "dal05"
        capacity = 20
        iops = 0.25
        
        # Optional fields
        allowed_virtual_guest_ids = [ 27699397 ]
        allowed_subnets = ["10.40.98.192/26"]
        allowed_ip_addresses = ["10.40.98.193", "10.40.98.200"]
        snapshot_capacity = 10
}

# Create 20G performance file storage and 100 IOPS option.
resource "softlayer_file_storage" "test2" {
        type = "Endurance"
        datacenter = "dal05"
        capacity = 20
        iops = 100
        
        # Optional fields
        allowed_virtual_guest_ids = [ 27699397 ]
        allowed_subnets = ["10.40.98.192/26"]
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
    * Specifies IOPS value for the storage. Please find available values for endurnace storage in the [link](https://knowledgelayer.softlayer.com/learning/introduction-endurance-storage).
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
* `allowed_subnets` | *array of string*
    * Specifies allowed subnets. Subnets should be in the same data center.
    * **Optional**    
* `allowed_ip_addresses` | *array of string*
    * Specifies allowed IP addresses. IP addresses should be in the same data center.
    * **Optional**    
    

## Attributes Reference

The following attributes are exported:

* `id` - id of the storage.
* `hostname` - The fully qualified domain name of the storage. 
* `volumename` - The name of the storage volume.
