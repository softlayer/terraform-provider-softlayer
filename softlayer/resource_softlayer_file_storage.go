package softlayer

import (
	"fmt"
	"log"
	"strconv"

	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const (
	storagePerformancePackageType = "ADDITIONAL_SERVICES_PERFORMANCE_STORAGE"
	storageEndurancePackageType   = "ADDITIONAL_SERVICES_ENTERPRISE_STORAGE"
	storageMask                   = "id,billingItem.orderItem.order.id"
	storageDetailMask             = "id,capacityGb,iops,storageType,username,serviceResourceBackendIpAddress,properties[type]" +
		",serviceResourceName,allowedIpAddresses,allowedSubnets,allowedVirtualGuests[id,allowedHost[name,credential[username,password]]],allowedHardware[id,allowedHost[name,credential[username,password]]],snapshotCapacityGb,osType"
	itemMask        = "id,capacity,description,units,keyName,prices[id,categories[id,name,categoryCode],capacityRestrictionMinimum,capacityRestrictionMaximum,locationGroupId]"
	enduranceType   = "Endurance"
	performanceType = "Performance"
	fileStorage     = "FILE_STORAGE"
	blockStorage    = "BLOCK_STORAGE"
	retryTime       = 5
)

var (
	// Map IOPS value to endurance storage tier keyName in SoftLayer_Product_Item
	enduranceIopsMap = map[float64]string{
		0.25: "LOW_INTENSITY_TIER",
		2:    "READHEAVY_TIER",
		4:    "WRITEHEAVY_TIER",
		10:   "10_IOPS_PER_GB",
	}

	// Map IOPS value to endurance storage tier capacityRestrictionMaximum/capacityRestrictionMinimum in SoftLayer_Product_Item
	enduranceCapacityRestrictionMap = map[float64]int{
		0.25: 100,
		2:    200,
		4:    300,
		10:   1000,
	}

	// storagePackageType is a storage package keyName for SoftLayer_Product_Package. It is used to filter storage package.
	// iopsCategoryCode is a storage IOPS categoryCode for SoftLayer_Product_Item. It is used to filter storage IOPS price.
	// storageProtocolCategoryCode is a storage protocol categoryCode for SoftLayer_Product_Item. It is used to filter storage protocol price.
	storagePackageMap = map[string](map[string](map[string]string)){
		fileStorage: {
			performanceType: {
				"storagePackageType":          storagePerformancePackageType,
				"iopsCategoryCode":            "performance_storage_iops",
				"storageProtocolCategoryCode": "performance_storage_nfs",
			},
			enduranceType: {
				"storagePackageType":          storageEndurancePackageType,
				"iopsCategoryCode":            "storage_tier_level",
				"storageProtocolCategoryCode": "storage_file",
			},
		},
		blockStorage: {
			performanceType: {
				"storagePackageType":          storagePerformancePackageType,
				"iopsCategoryCode":            "performance_storage_iops",
				"storageProtocolCategoryCode": "performance_storage_iscsi",
			},
			enduranceType: {
				"storagePackageType":          storageEndurancePackageType,
				"iopsCategoryCode":            "storage_tier_level",
				"storageProtocolCategoryCode": "storage_block",
			},
		},
	}
)

func resourceSoftLayerFileStorage() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerFileStorageCreate,
		Read:     resourceSoftLayerFileStorageRead,
		Update:   resourceSoftLayerFileStorageUpdate,
		Delete:   resourceSoftLayerFileStorageDelete,
		Exists:   resourceSoftLayerFileStorageExists,
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

			"allowed_virtual_guest_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set: func(v interface{}) int {
					return v.(int)
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

			"allowed_subnets": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"allowed_ip_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSoftLayerFileStorageCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	storageType := d.Get("type").(string)
	iops := d.Get("iops").(float64)
	datacenter := d.Get("datacenter").(string)
	capacity := d.Get("capacity").(int)
	snapshotCapacity := d.Get("snapshot_capacity").(int)

	storageOrderContainer, err := buildStorageProductOrderContainer(sess, storageType, iops, capacity, snapshotCapacity, fileStorage, datacenter)
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
			}, sl.Bool(false))
	case performanceType:
		receipt, err = services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_PerformanceStorage_Nfs{
				Container_Product_Order_Network_PerformanceStorage: datatypes.Container_Product_Order_Network_PerformanceStorage{
					Container_Product_Order: storageOrderContainer,
				},
			}, sl.Bool(false))
	default:
		return fmt.Errorf("Error during creation of storage: Invalid storageType %s", storageType)
	}

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}

	// Find the storage device
	fileStorage, err := findStorageByOrderId(sess, *receipt.OrderId)

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", *fileStorage.Id))

	// Wait for storage availability
	_, err = WaitForStorageAvailable(d, meta)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for storage (%s) to become ready: %s", d.Id(), err)
	}

	// SoftLayer changes the device ID after completion of provisioning. It is necessary to refresh device ID.
	fileStorage, err = findStorageByOrderId(sess, *receipt.OrderId)

	if err != nil {
		return fmt.Errorf("Error during creation of storage: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", *fileStorage.Id))

	log.Printf("[INFO] Storage ID: %s", d.Id())

	return resourceSoftLayerFileStorageUpdate(d, meta)
}

func resourceSoftLayerFileStorageRead(d *schema.ResourceData, meta interface{}) error {
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

	// Read allowed_subnets
	allowedSubnetsList := make([]string, 0, len(storage.AllowedSubnets))
	for _, allowedSubnets := range storage.AllowedSubnets {
		allowedSubnetsList = append(allowedSubnetsList, *allowedSubnets.NetworkIdentifier+"/"+strconv.Itoa(*allowedSubnets.Cidr))
	}
	d.Set("allowed_subnets", allowedSubnetsList)

	// Read allowed_virtual_guest_ids
	allowedVirtualGuestIdsList := make([]int, 0, len(storage.AllowedVirtualGuests))
	for _, allowedVirtualGuest := range storage.AllowedVirtualGuests {
		allowedVirtualGuestIdsList = append(allowedVirtualGuestIdsList, *allowedVirtualGuest.Id)
	}
	d.Set("allowed_virtual_guest_ids", allowedVirtualGuestIdsList)

	// Read allowed_hardware_ids
	allowedHardwareIdsList := make([]int, 0, len(storage.AllowedHardware))
	for _, allowedHW := range storage.AllowedHardware {
		allowedHardwareIdsList = append(allowedHardwareIdsList, *allowedHW.Id)
	}
	d.Set("allowed_hardware_ids", allowedHardwareIdsList)

	if storage.OsType != nil {
		d.Set("os_type", *storage.OsType.Name)
	}

	return nil
}

func resourceSoftLayerFileStorageUpdate(d *schema.ResourceData, meta interface{}) error {
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

	return resourceSoftLayerFileStorageRead(d, meta)
}

func resourceSoftLayerFileStorageDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	storageService := services.GetNetworkStorageService(sess)

	storageID, _ := strconv.Atoi(d.Id())

	// Get billing item associated with the storage
	billingItem, err := storageService.Id(storageID).GetBillingItem()

	if err != nil {
		return fmt.Errorf("Error while looking up billing item associated with the storage: %s", err)
	}

	if billingItem.Id == nil {
		return fmt.Errorf("Error while looking up billing item associated with the storage: No billing item for ID:%d", storageID)
	}

	success, err := services.GetBillingItemService(sess).Id(*billingItem.Id).CancelService()
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("SoftLayer reported an unsuccessful cancellation")
	}
	return nil
}

func resourceSoftLayerFileStorageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

	storageID, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = services.GetNetworkStorageService(sess).
		Id(storageID).
		GetObject()

	if err != nil {
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving storage information: %s", err)
	}
	return true, nil
}

func buildStorageProductOrderContainer(
	sess *session.Session,
	storageType string,
	iops float64,
	capacity int,
	snapshotCapacity int,
	storageProtocol string,
	datacenter string) (datatypes.Container_Product_Order, error) {

	// Build product item filters for performance storage
	iopsKeyName, err := getIopsKeyName(iops, storageType)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}
	capacityKeyName := fmt.Sprintf("%d_GB_", capacity)
	snapshotCapacityKeyName := fmt.Sprintf("%d_GB_", snapshotCapacity)

	storagePackageType := storagePackageMap[storageProtocol][storageType]["storagePackageType"]
	iopsCategoryCode := storagePackageMap[storageProtocol][storageType]["iopsCategoryCode"]
	storageProtocolCategoryCode := storagePackageMap[storageProtocol][storageType]["storageProtocolCategoryCode"]

	// Get a package type
	pkg, err := product.GetPackageByType(sess, storagePackageType)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Get all prices
	productItems, err := product.GetPackageProducts(sess, *pkg.Id, itemMask)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Add IOPS price
	targetItemPrices := []datatypes.Product_Item_Price{}
	var iopsPrice datatypes.Product_Item_Price

	if storageType == enduranceType {
		iopsPrice, err = getPrice(productItems, iopsKeyName, iopsCategoryCode, "", 0)
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
	} else {
		iopsPrice, err = getPrice(productItems, iopsKeyName, iopsCategoryCode, "STORAGE_SPACE", capacity)
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
	}

	targetItemPrices = append(targetItemPrices, iopsPrice)

	var capacityPrice datatypes.Product_Item_Price
	// Add capacity price
	if storageType == enduranceType {
		capacityPrice, err = getPrice(productItems, capacityKeyName, "performance_storage_space", "STORAGE_TIER_LEVEL", enduranceCapacityRestrictionMap[iops])
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
	} else {
		capacityPrice, err = getPrice(productItems, capacityKeyName, "performance_storage_space", "", 0)
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
	}
	targetItemPrices = append(targetItemPrices, capacityPrice)

	// Add storageProtocol price
	storageProtocolPrice, err := getPrice(productItems, storageProtocol, storageProtocolCategoryCode, "", 0)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}
	targetItemPrices = append(targetItemPrices, storageProtocolPrice)

	// Add Endurane Storage price
	if storageType == enduranceType {
		endurancePrice, err := getPrice(productItems, "CODENAME_PRIME_STORAGE_SERVICE", "storage_service_enterprise", "", 0)
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
		targetItemPrices = append(targetItemPrices, endurancePrice)
	}

	// Add snapshot capacity price
	if storageType == enduranceType && snapshotCapacity > 0 {
		snapshotCapacityPrice, err := getPrice(productItems, snapshotCapacityKeyName, "storage_snapshot_space", "STORAGE_TIER_LEVEL", enduranceCapacityRestrictionMap[iops])
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
		targetItemPrices = append(targetItemPrices, snapshotCapacityPrice)
	}

	// Lookup the data center ID
	dc, err := location.GetDatacenterByName(sess, datacenter)
	if err != nil {
		return datatypes.Container_Product_Order{},
			fmt.Errorf("No data centers matching %s could be found", datacenter)
	}

	productOrderContainer := datatypes.Container_Product_Order{
		PackageId: pkg.Id,
		Location:  sl.String(strconv.Itoa(*dc.Id)),
		Prices:    targetItemPrices,
		Quantity:  sl.Int(1),
	}

	return productOrderContainer, nil
}

func findStorageByOrderId(sess *session.Session, orderId int) (datatypes.Network_Storage, error) {
	filterPath := "networkStorage.billingItem.orderItem.order.id"

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			storage, err := services.GetAccountService(sess).
				Filter(filter.Build(
					filter.Path(filterPath).
						Eq(strconv.Itoa(orderId)))).
				Mask(storageMask).
				GetNetworkStorage()
			if err != nil {
				return datatypes.Network_Storage{}, "", err
			}

			if len(storage) == 1 {
				return storage[0], "complete", nil
			} else if len(storage) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one Storage: %s", err)
			}
		},
		Timeout:        45 * time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     10 * time.Second,
		NotFoundChecks: 300,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Storage{}, err
	}

	var result, ok = pendingResult.(datatypes.Network_Storage)

	if ok {
		return result, nil
	}

	return datatypes.Network_Storage{},
		fmt.Errorf("Cannot find Storage with order id '%d'", orderId)
}

// Waits for storage provisioning
func WaitForStorageAvailable(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for storage (%s) to be available.", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The storage ID %s must be numeric", d.Id())
	}
	sess := meta.(ProviderConfig).SoftLayerSession()

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", "provisioning"},
		Target:  []string{"available"},
		Refresh: func() (interface{}, string, error) {
			// Check active transactions
			service := services.GetNetworkStorageService(sess)
			result, err := service.Id(id).Mask("activeTransactions").GetObject()
			if err != nil {
				if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
					return nil, "", fmt.Errorf("Error retrieving storage: %s", err)
				}
				return false, "retry", nil
			}

			log.Println("Checking active transactions.")
			if len(result.ActiveTransactions) > 0 {
				return result, "provisioning", nil
			}

			// Check volume status.
			log.Println("Checking volume status.")
			netStore, err := service.Id(id).Mask("volumeStatus").GetObject()
			if err != nil {
				return false, "retry", nil
			}

			if netStore.VolumeStatus == nil || *netStore.VolumeStatus != "PROVISION_COMPLETED" {
				return result, "provisioning", nil
			}

			return result, "available", nil
		},
		Timeout:    45 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func getIopsKeyName(iops float64, storageType string) (string, error) {
	switch storageType {
	case enduranceType:
		return enduranceIopsMap[iops], nil
	case performanceType:
		return fmt.Sprintf("%.f_IOPS", iops), nil
	}
	return "", fmt.Errorf("Invalid storageType %s.", storageType)
}

func getPrice(productItems []datatypes.Product_Item, keyName string, categoryCode string, capacityRestrictionType string, capacityRestriction int) (datatypes.Product_Item_Price, error) {
	for _, item := range productItems {
		if strings.HasPrefix(*item.KeyName, keyName) {
			for _, price := range item.Prices {
				// When price.LocationGroupId is null, xml-rpc returns <value> <string/> </value> and
				// softlayer-go returns &0 instead of nil.
				if *price.Categories[0].CategoryCode == categoryCode && (price.LocationGroupId == nil || *price.LocationGroupId == 0) {
					if capacityRestrictionType == "STORAGE_SPACE" {
						if price.CapacityRestrictionMinimum == nil ||
							price.CapacityRestrictionMaximum == nil {
							continue
						}
						capacityRestrictionMinimum, _ := strconv.Atoi(*price.CapacityRestrictionMinimum)
						capacityRestrictionMaximum, _ := strconv.Atoi(*price.CapacityRestrictionMaximum)
						if capacityRestrictionMinimum > 0 &&
							capacityRestriction >= capacityRestrictionMinimum &&
							capacityRestriction <= capacityRestrictionMaximum {
							return price, nil
						}
					}

					if capacityRestrictionType == "STORAGE_TIER_LEVEL" {
						if price.CapacityRestrictionMinimum == nil ||
							price.CapacityRestrictionMaximum == nil {
							continue
						}
						capacityRestrictionMinimum, _ := strconv.Atoi(*price.CapacityRestrictionMinimum)
						capacityRestrictionMaximum, _ := strconv.Atoi(*price.CapacityRestrictionMaximum)
						if capacityRestrictionMinimum > 0 &&
							capacityRestriction == capacityRestrictionMinimum &&
							capacityRestriction == capacityRestrictionMaximum {
							return price, nil
						}
					}

					if capacityRestrictionType == "" && capacityRestriction == 0 {
						return price, nil
					}
				}
			}
		}
	}
	return datatypes.Product_Item_Price{},
		fmt.Errorf("No product items matching with keyName %s and categoryCode %s could be found", keyName, categoryCode)
}

func getIops(storage datatypes.Network_Storage, storageType string) (float64, error) {
	switch storageType {
	case enduranceType:
		for _, property := range storage.Properties {
			if *property.Type.Keyname == "PROVISIONED_IOPS" {
				provisionedIops, err := strconv.Atoi(*property.Value)
				if err != nil {
					return 0, err
				}
				enduranceIops := float64(provisionedIops / *storage.CapacityGb)
				if enduranceIops < 1 {
					enduranceIops = 0.25
				}
				return enduranceIops, nil
			}
		}
	case performanceType:
		if storage.Iops == nil {
			return 0, fmt.Errorf("Failed to retrieve iops information.")
		}
		iops, err := strconv.Atoi(*storage.Iops)
		if err != nil {
			return 0, err
		}
		return float64(iops), nil
	}
	return 0, fmt.Errorf("Invalid storage type %s", storageType)
}

func updateAllowedIpAddresses(d *schema.ResourceData, sess *session.Session, storage datatypes.Network_Storage) error {
	id := *storage.Id
	newIps := d.Get("allowed_ip_addresses").(*schema.Set).List()

	// Add new allowed_ip_addresses
	for _, newIp := range newIps {
		isNewIp := true
		for _, oldAllowedIpAddresses := range storage.AllowedIpAddresses {
			if newIp.(string) == *oldAllowedIpAddresses.IpAddress {
				isNewIp = false
				break
			}
		}
		if isNewIp {
			ipObject, err := services.GetAccountService(sess).
				Filter(filter.Build(
					filter.Path("ipAddresses.ipAddress").
						Eq(newIp.(string)))).GetIpAddresses()
			if err != nil {
				return err
			}
			if len(ipObject) != 1 {
				return fmt.Errorf("Number of IP address is %d", len(ipObject))
			}

			for {
				_, err = services.GetNetworkStorageService(sess).
					Id(id).
					AllowAccessFromHostList([]datatypes.Container_Network_Storage_Host{
						{
							Id:         ipObject[0].Id,
							ObjectType: sl.String("SoftLayer_Network_Subnet_IpAddress"),
						},
					})
				if err != nil {
					if strings.Contains(err.Error(), "SoftLayer_Exception_Network_Storage_Group_MassAccessControlModification") {
						time.Sleep(retryTime * time.Second)
						continue
					}
					return err
				}
				break
			}
		}
	}

	// Remove deleted allowed_hardware_ids
	for _, oldAllowedIpAddresses := range storage.AllowedIpAddresses {
		isDeletedId := true
		for _, newIp := range newIps {
			if newIp.(string) == *oldAllowedIpAddresses.IpAddress {
				isDeletedId = false
				break
			}
		}
		if isDeletedId {
			for {
				_, err := services.GetNetworkStorageService(sess).
					Id(id).
					RemoveAccessFromHostList([]datatypes.Container_Network_Storage_Host{
						{
							Id:         oldAllowedIpAddresses.Id,
							ObjectType: sl.String("SoftLayer_Network_Subnet_IpAddress"),
						},
					})
				if err != nil {
					if strings.Contains(err.Error(), "SoftLayer_Exception_Network_Storage_Group_MassAccessControlModification") {
						time.Sleep(retryTime * time.Second)
						continue
					}
					return err
				}
				break
			}
		}
	}
	return nil
}

func updateAllowedSubnets(d *schema.ResourceData, sess *session.Session, storage datatypes.Network_Storage) error {
	id := *storage.Id
	newSubnets := d.Get("allowed_subnets").(*schema.Set).List()

	// Add new allowed_subnets
	for _, newSubnet := range newSubnets {
		isNewSubnet := true
		newSubnetArr := strings.Split(newSubnet.(string), "/")
		newNetworkIdentifier := newSubnetArr[0]
		newCidr, err := strconv.Atoi(newSubnetArr[1])
		if err != nil {
			return err
		}
		for _, oldAllowedSubnets := range storage.AllowedSubnets {
			if newNetworkIdentifier == *oldAllowedSubnets.NetworkIdentifier && newCidr == *oldAllowedSubnets.Cidr {
				isNewSubnet = false
				break
			}
		}
		if isNewSubnet {
			filterStr := fmt.Sprintf("{\"subnets\":{\"networkIdentifier\":{\"operation\":\"%s\"},\"cidr\":{\"operation\":\"%d\"}}}", newNetworkIdentifier, newCidr)
			subnetObject, err := services.GetAccountService(sess).
				Filter(filterStr).GetSubnets()
			if err != nil {
				return err
			}
			if len(subnetObject) != 1 {
				return fmt.Errorf("Number of subnet is %d", len(subnetObject))
			}
			_, err = services.GetNetworkStorageService(sess).
				Id(id).
				AllowAccessFromHostList([]datatypes.Container_Network_Storage_Host{
					{
						Id:         subnetObject[0].Id,
						ObjectType: sl.String("SoftLayer_Network_Subnet"),
					},
				})
			if err != nil {
				return err
			}
		}
	}

	// Remove deleted allowed_subnets
	for _, oldAllowedSubnets := range storage.AllowedSubnets {
		isDeletedSubnet := true
		for _, newSubnet := range newSubnets {
			newSubnetArr := strings.Split(newSubnet.(string), "/")
			newNetworkIdentifier := newSubnetArr[0]
			newCidr, err := strconv.Atoi(newSubnetArr[1])
			if err != nil {
				return err
			}

			if newNetworkIdentifier == *oldAllowedSubnets.NetworkIdentifier && newCidr == *oldAllowedSubnets.Cidr {
				isDeletedSubnet = false
				break
			}
		}
		if isDeletedSubnet {
			_, err := services.GetNetworkStorageService(sess).
				Id(id).
				RemoveAccessFromHostList([]datatypes.Container_Network_Storage_Host{
					{
						Id:         sl.Int(*oldAllowedSubnets.Id),
						ObjectType: sl.String("SoftLayer_Network_Subnet"),
					},
				})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateAllowedVirtualGuestIds(d *schema.ResourceData, sess *session.Session, storage datatypes.Network_Storage) error {
	id := *storage.Id
	newIds := d.Get("allowed_virtual_guest_ids").(*schema.Set).List()

	// Add new allowed_virtual_guest_ids
	for _, newId := range newIds {
		isNewId := true
		for _, oldAllowedVirtualGuest := range storage.AllowedVirtualGuests {
			if newId.(int) == *oldAllowedVirtualGuest.Id {
				isNewId = false
				break
			}
		}
		if isNewId {
			for {
				_, err := services.GetNetworkStorageService(sess).
					Id(id).
					AllowAccessFromHostList([]datatypes.Container_Network_Storage_Host{
						{
							Id:         sl.Int(newId.(int)),
							ObjectType: sl.String("SoftLayer_Virtual_Guest"),
						},
					})
				if err != nil {
					if strings.Contains(err.Error(), "SoftLayer_Exception_Network_Storage_Group_MassAccessControlModification") {
						time.Sleep(retryTime * time.Second)
						continue
					}
					return err
				}
				break
			}
		}
	}

	// Remove deleted allowed_virtual_guest_ids
	for _, oldAllowedVirtualGuest := range storage.AllowedVirtualGuests {
		isDeletedId := true
		for _, newId := range newIds {
			if newId.(int) == *oldAllowedVirtualGuest.Id {
				isDeletedId = false
				break
			}
		}
		if isDeletedId {
			for {
				_, err := services.GetNetworkStorageService(sess).
					Id(id).
					RemoveAccessFromHostList([]datatypes.Container_Network_Storage_Host{
						{
							Id:         sl.Int(*oldAllowedVirtualGuest.Id),
							ObjectType: sl.String("SoftLayer_Virtual_Guest"),
						},
					})
				if err != nil {
					if strings.Contains(err.Error(), "SoftLayer_Exception_Network_Storage_Group_MassAccessControlModification") {
						time.Sleep(retryTime * time.Second)
						continue
					}
					return err
				}
				break
			}
		}
	}
	return nil
}

func updateAllowedHardwareIds(d *schema.ResourceData, sess *session.Session, storage datatypes.Network_Storage) error {
	id := *storage.Id
	newIds := d.Get("allowed_hardware_ids").(*schema.Set).List()

	// Add new allowed_hardware_ids
	for _, newId := range newIds {
		isNewId := true
		for _, oldAllowedHardware := range storage.AllowedHardware {
			if newId.(int) == *oldAllowedHardware.Id {
				isNewId = false
				break
			}
		}
		if isNewId {
			_, err := services.GetNetworkStorageService(sess).
				Id(id).
				AllowAccessFromHostList([]datatypes.Container_Network_Storage_Host{
					{
						Id:         sl.Int(newId.(int)),
						ObjectType: sl.String("SoftLayer_Hardware"),
					},
				})
			if err != nil {
				return err
			}
		}
	}

	// Remove deleted allowed_hardware_ids
	for _, oldAllowedHardware := range storage.AllowedHardware {
		isDeletedId := true
		for _, newId := range newIds {
			if newId.(int) == *oldAllowedHardware.Id {
				isDeletedId = false
				break
			}
		}
		if isDeletedId {
			_, err := services.GetNetworkStorageService(sess).
				Id(id).
				RemoveAccessFromHostList([]datatypes.Container_Network_Storage_Host{
					{
						Id:         sl.Int(*oldAllowedHardware.Id),
						ObjectType: sl.String("SoftLayer_Hardware"),
					},
				})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
