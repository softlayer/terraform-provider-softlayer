package softlayer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/helpers/network"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/sl"
	"regexp"
	"strings"
)

func resourceSoftLayerBlockStorage() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerBlockStorageCreate,
		Read:     resourceSoftLayerBlockStorageRead,
		Update:   resourceSoftLayerBlockStorageUpdate,
		Delete:   resourceSoftLayerBlockStorageDelete,
		Exists:   resourceSoftLayerBlockStorageExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"datacenter": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"iops": {
				Type:     schema.TypeFloat,
				Required: true,
				ForceNew: true,
			},

			"volumename": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"os_format_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allowed_virtual_guest_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set: func(v interface{}) int {
					return v.(int)
				},
			},

			"allowed_virtual_guest_info": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostIQN": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					virtualGuest := v.(map[string]interface{})
					return virtualGuest["id"].(int)
				},
			},

			"allowed_hardware_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set: func(v interface{}) int {
					return v.(int)
				},
			},

			"allowed_hardware_info": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostIQN": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					baremetal := v.(map[string]interface{})
					return baremetal["id"].(int)
				},
			},

			"allowed_ip_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSoftLayerBlockStorageCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	storageType := d.Get("type").(string)
	iops := d.Get("iops").(float64)
	datacenter := d.Get("datacenter").(string)
	capacity := d.Get("capacity").(int)
	snapshotCapacity := d.Get("snapshot_capacity").(int)
	osFormatType := d.Get("os_format_type").(string)
	osType, err := network.GetOsTypeByName(sess, osFormatType)

	if err != nil {
		return err
	}

	storageOrderContainer, err := buildStorageProductOrderContainer(sess, storageType, iops, capacity, snapshotCapacity, blockStorage, datacenter)
	if err != nil {
		return fmt.Errorf("Error while creating storage:%s", err)
	}

	log.Println("[INFO] Creating storage")

	var receipt datatypes.Container_Product_Order_Receipt

	switch storageType {
	case enduranceType:
		receipt, err = services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_Storage_Enterprise{
				Container_Product_Order: storageOrderContainer,
				OsFormatType: &datatypes.Network_Storage_Iscsi_OS_Type{
					Id:      osType.Id,
					KeyName: osType.KeyName,
				},
			}, sl.Bool(false))
	case performanceType:
		receipt, err = services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_PerformanceStorage_Iscsi{
				Container_Product_Order_Network_PerformanceStorage: datatypes.Container_Product_Order_Network_PerformanceStorage{
					Container_Product_Order: storageOrderContainer,
				},
				OsFormatType: &datatypes.Network_Storage_Iscsi_OS_Type{
					Id:      osType.Id,
					KeyName: osType.KeyName,
				},
			}, sl.Bool(false))
	default:
		return fmt.Errorf("Error during creation of storage: Invalid storageType %s", storageType)
	}

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}

	// Find the storage device
	blockStorage, err := findStorageByOrderId(sess, *receipt.OrderId)

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", *blockStorage.Id))

	// Wait for storage availability
	_, err = WaitForStorageAvailable(d, meta)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for storage (%s) to become ready: %s", d.Id(), err)
	}

	// SoftLayer changes the device ID after completion of provisioning. It is necessary to refresh device ID.
	blockStorage, err = findStorageByOrderId(sess, *receipt.OrderId)

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", *blockStorage.Id))

	log.Printf("[INFO] Storage ID: %s", d.Id())

	return resourceSoftLayerBlockStorageUpdate(d, meta)
}

func resourceSoftLayerBlockStorageRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	storageId, _ := strconv.Atoi(d.Id())

	storage, err := services.GetNetworkStorageService(sess).
		Id(storageId).
		Mask(storageDetailMask).
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving storage information: %s", err)
	}

	storageType := strings.Fields(*storage.StorageType.Description)[0]

	// Calculate IOPS
	iops, err := getIops(storage, storageType)
	if err != nil {
		return fmt.Errorf("Error retrieving storage information: %s", err)
	}

	d.Set("type", storageType)
	d.Set("capacity", *storage.CapacityGb)
	d.Set("volumename", *storage.Username)
	d.Set("hostname", *storage.ServiceResourceBackendIpAddress)
	d.Set("iops", iops)
	if storage.SnapshotCapacityGb != nil {
		snapshotCapacity, _ := strconv.Atoi(*storage.SnapshotCapacityGb)
		d.Set("snapshot_capacity", snapshotCapacity)
	}

	// Parse data center short name from ServiceResourceName. For example,
	// if SoftLayer API returns "'serviceResourceName': 'PerfStor Aggr aggr_staasdal0601_p01'",
	// the data center short name is "dal06".
	r, _ := regexp.Compile("[a-zA-Z]{3}[0-9]{2}")
	d.Set("datacenter", r.FindString(*storage.ServiceResourceName))

	// Read allowed_ip_addresses
	allowedIpaddressesList := make([]string, 0, len(storage.AllowedIpAddresses))
	for _, allowedIpaddress := range storage.AllowedIpAddresses {
		allowedIpaddressesList = append(allowedIpaddressesList, *allowedIpaddress.IpAddress)
	}
	d.Set("allowed_ip_addresses", allowedIpaddressesList)

	// Read allowed_virtual_guest_ids and allowed_virtual_guest_info
	allowedVirtualGuestInfoList := make([]map[string]interface{}, 0)
	allowedVirtualGuestIdsList := make([]int, 0, len(storage.AllowedVirtualGuests))

	for _, allowedVirtualGuest := range storage.AllowedVirtualGuests {
		singleVirtualGuest := make(map[string]interface{})
		singleVirtualGuest["id"] = *allowedVirtualGuest.Id
		singleVirtualGuest["username"] = *allowedVirtualGuest.AllowedHost.Credential.Username
		singleVirtualGuest["password"] = *allowedVirtualGuest.AllowedHost.Credential.Password
		singleVirtualGuest["hostIQN"] = *allowedVirtualGuest.AllowedHost.Name
		allowedVirtualGuestInfoList = append(allowedVirtualGuestInfoList, singleVirtualGuest)
		allowedVirtualGuestIdsList = append(allowedVirtualGuestIdsList, *allowedVirtualGuest.Id)
	}
	d.Set("allowed_virtual_guest_ids", allowedVirtualGuestIdsList)
	d.Set("allowed_virtual_guest_info", allowedVirtualGuestInfoList)

	// Read allowed_hardware_ids and allowed_hardware_info
	allowedHardwareInfoList := make([]map[string]interface{}, 0)
	allowedHardwareIdsList := make([]int, 0, len(storage.AllowedHardware))
	for _, allowedHW := range storage.AllowedHardware {
		singleHardware := make(map[string]interface{})
		singleHardware["id"] = *allowedHW.Id
		singleHardware["username"] = *allowedHW.AllowedHost.Credential.Username
		singleHardware["password"] = *allowedHW.AllowedHost.Credential.Password
		singleHardware["hostIQN"] = *allowedHW.AllowedHost.Name
		allowedHardwareInfoList = append(allowedHardwareInfoList, singleHardware)
		allowedHardwareIdsList = append(allowedHardwareIdsList, *allowedHW.Id)
	}
	d.Set("allowed_hardware_ids", allowedHardwareIdsList)
	d.Set("allowed_hardware_info", allowedHardwareInfoList)

	if storage.OsType != nil {
		d.Set("os_type", *storage.OsType.Name)
	}

	return nil
}

func resourceSoftLayerBlockStorageUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	storage, err := services.GetNetworkStorageService(sess).
		Id(id).
		Mask(storageDetailMask).
		GetObject()

	if err != nil {
		return fmt.Errorf("Error updating storage information: %s", err)
	}

	// Update allowed_ip_addresses
	if d.HasChange("allowed_ip_addresses") {
		err := updateAllowedIpAddresses(d, sess, storage)
		if err != nil {
			return fmt.Errorf("Error updating storage information: %s", err)
		}
	}

	// Update allowed_subnets
	if d.HasChange("allowed_subnets") {
		err := updateAllowedSubnets(d, sess, storage)
		if err != nil {
			return fmt.Errorf("Error updating storage information: %s", err)
		}
	}

	// Update allowed_virtual_guest_ids
	if d.HasChange("allowed_virtual_guest_ids") {
		err := updateAllowedVirtualGuestIds(d, sess, storage)
		if err != nil {
			return fmt.Errorf("Error updating storage information: %s", err)
		}
	}

	// Update allowed_hardware_ids
	if d.HasChange("allowed_hardware_ids") {
		err := updateAllowedHardwareIds(d, sess, storage)
		if err != nil {
			return fmt.Errorf("Error updating storage information: %s", err)
		}
	}

	return resourceSoftLayerBlockStorageRead(d, meta)
}

func resourceSoftLayerBlockStorageDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceSoftLayerFileStorageDelete(d, meta)
}

func resourceSoftLayerBlockStorageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return resourceSoftLayerFileStorageExists(d, meta)
}
