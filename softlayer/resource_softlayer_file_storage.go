package softlayer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"strings"
	"time"
)

const (
	StoragePerformancePackageType = "ADDITIONAL_SERVICES_PERFORMANCE_STORAGE"
	StorageEndurancePackageType   = "ADDITIONAL_SERVICES_ENTERPRISE_STORAGE"
	storageMask                   = "id,networkStorage.billingItem.orderItem.order.id"
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

			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"iops": {
				Type:     schema.TypeFloat,
				Required: true,
				ForceNew: true,
			},

			"snapshot": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func buildStorageProductOrderContainer(sess *session.Session,
	storageType string,
	iops float64,
	size int,
	storageProtocol string,
	datacenter string) (datatypes.Container_Product_Order, error) {

	// Build product item filters for performance storage
	fileStoragePackageType := StoragePerformancePackageType
	iopsDescription := fmt.Sprintf("%.f IOPS", iops)
	sizeKeyName := fmt.Sprintf("%d_GB_PERFORMANCE_STORAGE_SPACE", size)
	iopsDescriptionCount := 0
	sizeKeyNameCount := 0
	storageProtocolCount := 0
	storageTypeCount := 0

	// Build product item filters for endurance storage
	if storageType == "Endurance" {
		fileStoragePackageType = StorageEndurancePackageType
		iopsDescription = fmt.Sprintf("%.f IOPS per GB", iops)
		if iops != float64(int(iops)) {
			iopsDescription = fmt.Sprintf("%.2f IOPS per GB", iops)
		}
	}

	pkg, err := product.GetPackageByType(sess, fileStoragePackageType)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Get all prices
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Select only those product items with a matching keyname
	targetItemPrices := []datatypes.Product_Item_Price{}
	for _, item := range productItems {
		if *item.Description == iopsDescription {
			for _, price := range item.Prices {
				if *price.Categories[0].CategoryCode == "storage_tier_level" ||
					*price.Categories[0].CategoryCode == "performance_storage_iops" {
					targetItemPrices = append(targetItemPrices, price)
					iopsDescriptionCount++
					break
				}
			}
		}
		if iopsDescriptionCount > 0 {
			break
		}
	}
	if iopsDescriptionCount == 0 {
		return datatypes.Container_Product_Order{},
			fmt.Errorf("No product items matching %s could be found", iopsDescription)
	}

	for _, item := range productItems {
		if *item.KeyName == sizeKeyName {
			for _, price := range item.Prices {
				if *price.Categories[0].CategoryCode == "performance_storage_space" {
					targetItemPrices = append(targetItemPrices, price)
					sizeKeyNameCount++
					break
				}
			}
		}
		if sizeKeyNameCount > 0 {
			break
		}
	}
	if sizeKeyNameCount == 0 {
		return datatypes.Container_Product_Order{},
			fmt.Errorf("No product items matching %s could be found", sizeKeyName)
	}

	if storageType == "Endurance" {
		for _, item := range productItems {
			if *item.Description == (storageType + " Storage") {
				for _, price := range item.Prices {
					if *price.Categories[0].CategoryCode == "storage_service_enterprise" {
						targetItemPrices = append(targetItemPrices, price)
						storageTypeCount++
						break
					}
				}
			}
			if storageTypeCount > 0 {
				break
			}
		}
		if storageTypeCount == 0 {
			return datatypes.Container_Product_Order{},
				fmt.Errorf("No product items matching %s could be found", storageType)
		}
	}

	for _, item := range productItems {
		if *item.Description == storageProtocol || *item.Description == (storageProtocol+" (Performance)") {
			for _, price := range item.Prices {
				if *price.Categories[0].CategoryCode == "storage_file" ||
					*price.Categories[0].CategoryCode == "performance_storage_nfs" {
					targetItemPrices = append(targetItemPrices, price)
					storageProtocolCount++
					break
				}
			}
		}
		if storageProtocolCount > 0 {
			break
		}
	}
	if storageProtocolCount == 0 {
		return datatypes.Container_Product_Order{},
			fmt.Errorf("No product items matching %s could be found", storageProtocol)
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

func resourceSoftLayerFileStorageCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	storageType := d.Get("type").(string)
	iops := d.Get("iops").(float64)
	storageProtocol := "File Storage"
	datacenter := d.Get("datacenter").(string)
	size := d.Get("size").(int)

	storageOrderContainer, err := buildStorageProductOrderContainer(sess, storageType, iops, size, storageProtocol, datacenter)
	if err != nil {
		return fmt.Errorf("Error while creating file storage:%s", err)
	}

	log.Println("[INFO] Creating file storage")

	if storageType == "Endurance" {
		receipt, err := services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_Storage_Enterprise{
				Container_Product_Order: storageOrderContainer,
			}, sl.Bool(false))
		fileStorage, err := findFileStorageByOrderId(sess, *receipt.OrderId)

		if err != nil {
			return fmt.Errorf("Error during creation of file storage: %s", err)
		}
		d.SetId(fmt.Sprintf("%d", *fileStorage.Id))

	} else {
		receipt, err := services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_PerformanceStorage_Nfs{
				Container_Product_Order_Network_PerformanceStorage: datatypes.Container_Product_Order_Network_PerformanceStorage{
					Container_Product_Order: storageOrderContainer,
				},
			}, sl.Bool(false))
		fileStorage, err := findFileStorageByOrderId(sess, *receipt.OrderId)

		if err != nil {
			return fmt.Errorf("Error during creation of file storage: %s", err)
		}
		d.SetId(fmt.Sprintf("%d", *fileStorage.Id))
	}
	log.Printf("[INFO] Storage ID: %s", d.Id())

	return resourceSoftLayerFwHardwareDedicatedRead(d, meta)
}

func resourceSoftLayerFileStorageRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	storageId, _ := strconv.Atoi(d.Id())

	storage, err := services.GetNetworkStorageService(sess).
		Id(storageId).
		Mask("id,capacityGb,iops,storageType").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving storage information: %s", err)
	}

	d.Set("type", strings.Fields(*storage.StorageType.Description)[0])
	d.Set("datacenter", "")
	d.Set("size", *storage.CapacityGb)
	iops, err := strconv.Atoi(*storage.Iops)
	d.Set("iops", float64(iops))
	d.Set("snapshot", "")
	return nil
}

func resourceSoftLayerFileStorageUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
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

func findFileStorageByOrderId(sess *session.Session, orderId int) (datatypes.Network_Storage, error) {
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
		Timeout:    45 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
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
