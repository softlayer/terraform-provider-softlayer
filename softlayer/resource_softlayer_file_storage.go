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

var (
	enduranceIOPSMap = map[float64]string{
		0.25: "LOW_INTENSITY_TIER",
		2:    "READHEAVY_TIER",
		4:    "WRITEHEAVY_TIER",
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

func buildStorageProductOrderContainer(
	sess *session.Session,
	storageType string,
	iops float64,
	size int,
	storageProtocol string,
	datacenter string) (datatypes.Container_Product_Order, error) {

	// Build product item filters for performance storage
	storagePackageType := StoragePerformancePackageType
	iopsKeyName, _ := getIOPSKeyName(iops, storageType)
	iopsCategoryCode := "performance_storage_iops"
	storageProtocolCategoryCode := "performance_storage_nfs"
	sizeKeyName := fmt.Sprintf("%d_GB_PERFORMANCE_STORAGE_SPACE", size)

	// Update product item filters for endurance storage
	if storageType == "Endurance" {
		storagePackageType = StorageEndurancePackageType
		iopsCategoryCode = "storage_tier_level"
		storageProtocolCategoryCode = "storage_file"
	}

	// Get a package type
	pkg, err := product.GetPackageByType(sess, storagePackageType)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Get all prices
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}

	// Add IOPS price
	targetItemPrices := []datatypes.Product_Item_Price{}
	iopsPrice, err := getPrice(productItems, iopsKeyName, iopsCategoryCode)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}
	targetItemPrices = append(targetItemPrices, iopsPrice)

	// Add size price
	sizePrice, err := getPrice(productItems, sizeKeyName, "performance_storage_space")
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}
	targetItemPrices = append(targetItemPrices, sizePrice)

	// Add Endurane Storage price
	if storageType == "Endurance" {
		endurancePrice, err := getPrice(productItems, "CODENAME_PRIME_STORAGE_SERVICE", "performance_storage_space")
		if err != nil {
			return datatypes.Container_Product_Order{}, err
		}
		targetItemPrices = append(targetItemPrices, endurancePrice)
	}

	// Add storageProtocol price
	storageProtocolPrice, err := getPrice(productItems, storageProtocol, storageProtocolCategoryCode)
	if err != nil {
		return datatypes.Container_Product_Order{}, err
	}
	targetItemPrices = append(targetItemPrices, storageProtocolPrice)

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
	datacenter := d.Get("datacenter").(string)
	size := d.Get("size").(int)

	storageOrderContainer, err := buildStorageProductOrderContainer(sess, storageType, iops, size, "FILE_STORAGE", datacenter)
	if err != nil {
		return fmt.Errorf("Error while creating file storage:%s", err)
	}

	log.Println("[INFO] Creating file storage")

	if storageType == "Endurance" {
		receipt, err := services.GetProductOrderService(sess).PlaceOrder(
			&datatypes.Container_Product_Order_Network_Storage_Enterprise{
				Container_Product_Order: storageOrderContainer,
			}, sl.Bool(false))
		fileStorage, err := findStorageByOrderId(sess, *receipt.OrderId)

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
		fileStorage, err := findStorageByOrderId(sess, *receipt.OrderId)

		if err != nil {
			return fmt.Errorf("Error during creation of file storage: %s", err)
		}
		d.SetId(fmt.Sprintf("%d", *fileStorage.Id))
	}

	// wait for storage availability
	_, err = WaitForNoActiveStorageTransactions(d, meta)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for storage (%s) to become ready: %s", d.Id(), err)
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

// Wait for no active transactions
func WaitForNoActiveStorageTransactions(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for storage (%s) to have zero active transactions", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The storage ID %s must be numeric", d.Id())
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", "active"},
		Target:  []string{"idle"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetNetworkStorageService(meta.(ProviderConfig).SoftLayerSession())
			transactions, err := service.Id(id).GetActiveTransactions()
			if err != nil {
				if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
					return nil, "", fmt.Errorf("Couldn't get active transactions: %s", err)
				}

				return false, "retry", nil
			}

			if len(transactions) == 0 {
				return transactions, "idle", nil
			}

			return transactions, "active", nil
		},
		Timeout:    45 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func getIOPSKeyName(iops float64, storageType string) (string, error) {
	if storageType == "Endurance" {
		return enduranceIOPSMap[iops], nil
	}
	if storageType == "Performance" {
		return fmt.Sprintf("%.f_IOPS", iops), nil
	}
	return "", fmt.Errorf("Invalid storageType %s. StorageType should be 'Endurnace' or 'Performance'.", storageType)
}

func getPrice(productItems []datatypes.Product_Item, keyName string, categoryCode string) (datatypes.Product_Item_Price, error) {
	for _, item := range productItems {
		if strings.HasPrefix(*item.KeyName, keyName) {
			for _, price := range item.Prices {
				if *price.Categories[0].CategoryCode == categoryCode {
					return price, nil
				}
			}
		}
	}
	return datatypes.Product_Item_Price{},
		fmt.Errorf("No product items matching with keyName %s and categoryCode %s could be found", keyName, categoryCode)
}
