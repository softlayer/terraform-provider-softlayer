package softlayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	//"log"
	"strconv"
	"strings"
	"time"
)

var (
	PACKAGE_IDS = map[string]int{
		"endurance": 240,
	}
	CONTAINERS = map[string]string{
		"endurance": "Softlayer_Container_Product_Order_Network_Storage_Enterprise",
	}
	CATEGORY_CODES = map[string]string{
		"endurance": "storage_service_enterprise",
		"block":     "storage_block",
	}
	ENDURANCE_TIERS = map[float64]int{
		0.25: 100,
		2:    200,
		4:    300,
		10:   1000,
	}
)

func resourceSoftLayerNetworkStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerNetworkStorageCreate,
		Read:   resourceSoftLayerNetworkStorageRead,
		//		Update: resourceSoftLayerEnduranceStorageUpdate,
		Delete: resourceSoftLayerNetworkStorageDelete,
		Exists: resourceSoftLayerNetworkStorageExists,
		//		Importer: &schema.ResourceImporter{},
		//FORCENEW is here until an Update method is created.
		Schema: map[string]*schema.Schema{
			//endurance or performance
			"tier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"capacity_gb": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"datacenter": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"iops": {
				Type:     schema.TypeFloat,
				Required: true,
				ForceNew: true,
			},
			"nas_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"os_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func waitForDriveProvision(sess *session.Session, order *datatypes.Container_Product_Order_Receipt) (interface{}, error) {

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", "pending"},
		Target:  []string{"provisioned"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetAccountService(sess)
			path := strings.Join([]string{
				"iscsiNetworkStorage",
				"billingItem",
				"orderItem",
				"order",
				"id",
			}, ".")

			stores, err := service.
				Filter(filter.Path(path).Eq(order.OrderId).Build()).
				GetIscsiNetworkStorage()

			if err != nil {
				return false, "retry", err
			}

			if len(stores) == 0 || stores[0].CreateDate == nil {
				return nil, "pending", nil
			} else {
				return stores[0], "provisioned", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 5 * time.Minute,
	}

	return stateConf.WaitForState()
}

func resourceSoftLayerNetworkStorageRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetNetworkStorageService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).GetObject()

	if err != nil {
		return err
	}
	resJSON, err := json.Marshal(result)

	if err != nil {
		return err
	}
	d.Set("capacity_gb", *result.CapacityGb)
	d.Set("create_date", *result.CreateDate)

	return nil
}

func findBlockOSId(sess *session.Session, osName string) (*datatypes.Network_Storage_Iscsi_OS_Type, error) {
	osService := services.GetNetworkStorageIscsiOSTypeService(sess)

	oses, err := osService.GetAllObjects()

	if err != nil {
		return nil, err
	}

	for _, os := range oses {
		if *os.KeyName == osName {
			return &os, nil
		}
	}

	return nil, fmt.Errorf("No OS found matching %s.", osName)
}

func getLocationID(sess *session.Session, location string) (*string, error) {
	locationServ := services.GetLocationService(sess)
	datacenters, err := locationServ.Mask("longName,id,name").GetDatacenters()

	if err != nil {
		return nil, err
	}

	for _, datacenter := range datacenters {
		if *datacenter.Name == location {
			id := strconv.Itoa(*datacenter.Id)

			return &id, nil
		}
	}

	return nil, fmt.Errorf("No datacenter found with name %s", location)
}

func hasCategory(categoryCode string, categories []datatypes.Product_Item_Category) bool {
	result := false

	for _, category := range categories {
		if *category.CategoryCode == categoryCode {
			result = true
			break
		}
	}
	return result
}

func findPrice(items []datatypes.Product_Item, category string) (*datatypes.Product_Item_Price, error) {
	for _, item := range items { //Product_Item
		for _, price := range item.Prices { //Product_Item_Price
			if price.LocationGroupId != nil { //could maybe also match datacenter id.
				continue
			}

			if !hasCategory(category, price.Categories) {
				continue
			}

			return &price, nil
		}
	}
	return nil, fmt.Errorf("No Product_Item found with matching category: %s", category)
}

func findIOPSPrice(tier int, items []datatypes.Product_Item) (*datatypes.Product_Item_Price, error) {
	category := "storage_tier_level"
	for _, item := range items { //Product_Item
		if item.Attributes != nil {
			for _, attribute := range item.Attributes {
				value64, err := strconv.ParseInt(*attribute.Value, 10, 0)
				if err != nil {
					return nil, err
				}

				if err != nil {
					return nil, err
				}

				if err != nil {
					return nil, err
				}

				value := int(value64)
				if value == tier {
					for _, price := range item.Prices {
						if price.LocationGroupId != nil {
							continue
						}

						if !hasCategory(category, price.Categories) {
							continue
						}
						return &price, nil
					}
					break
				} else {
					continue
				}
			}
		}
	}

	return nil, fmt.Errorf("No Product_item found with matching IOPS level.")
}

func findSpacePrice(iopsCat int, desiredSize int, items []datatypes.Product_Item) (*datatypes.Product_Item_Price, error) {
	category := "performance_storage_space"
	for _, item := range items {

		size := int(*item.Capacity)

		if size != desiredSize {
			continue
		}

		for _, price := range item.Prices {
			if price.LocationGroupId != nil {
				continue
			}

			if !hasCategory(category, price.Categories) {
				continue
			}

			capMin, err := strconv.ParseInt(*price.CapacityRestrictionMinimum, 10, 0)
			if err != nil {
				return nil, err
			}

			if iopsCat < int(capMin) {
				continue
			}

			capMax, err := strconv.ParseInt(*price.CapacityRestrictionMaximum, 10, 0)
			if err != nil {
				return nil, err
			}
			if iopsCat > int(capMax) {
				continue
			}

			return &price, nil
		}
	}
	return nil, fmt.Errorf("No Product_Item with matching space size for tier.")
}

func resourceSoftLayerNetworkStorageCreate(schem *schema.ResourceData, meta interface{}) error {
	var prices []datatypes.Product_Item_Price
	sess := meta.(*session.Session)
	tier := schem.Get("tier").(string)
	nasType := schem.Get("nas_type").(string)
	capacity := schem.Get("capacity_gb").(int)

	//Get Location ID
	loc, err := getLocationID(sess, schem.Get("datacenter").(string))

	if err != nil {
		return err
	}

	schem.Set("datacenter", loc)

	//Get the Product_Package for storage in question.
	pack, err := getPackagePrices(sess, tier)

	if err != nil {
		return err
	}
	//TIER prices
	tierPrice, err := findPrice(pack.Items, CATEGORY_CODES[tier])

	if err != nil {
		return err
	}

	prices = append(prices, *tierPrice)

	nasTypePrice, err := findPrice(pack.Items, CATEGORY_CODES[nasType])

	if err != nil {
		return err
	}

	prices = append(prices, *nasTypePrice)

	//IOPS prices.
	iopsCategory := ENDURANCE_TIERS[schem.Get("iops").(float64)]

	iopsPrice, err := findIOPSPrice(iopsCategory, pack.Items)

	if err != nil {
		return err
	}

	prices = append(prices, *iopsPrice)

	//SPACE prices
	spacePrice, err := findSpacePrice(iopsCategory, capacity, pack.Items)

	if err != nil {
		return err
	}

	prices = append(prices, *spacePrice)

	//SNAPSHOT SPACE prices.

	//Build Order
	order, err := buildOrder(sess, prices, schem)

	if err != nil {
		return err
	}
	//Verify Order
	orderService := services.GetProductOrderService(sess)
	_, err = orderService.VerifyOrder(order)

	if err != nil {
		return err
	}

	orderService = services.GetProductOrderService(sess)
	saveAsQuote := false
	bill, err := orderService.PlaceOrder(order, &saveAsQuote)

	if err != nil {
		return err
	}

	store, err := waitForDriveProvision(sess, &bill)

	id := strconv.Itoa(*store.(datatypes.Network_Storage).
		Id)

	schem.SetId(id)

	return resourceSoftLayerNetworkStorageRead(schem, sess)
}

func buildOrder(sess *session.Session, prices []datatypes.Product_Item_Price, schem *schema.ResourceData) (order *datatypes.Container_Product_Order_Network_Storage_Enterprise, err error) {
	tier := "endurance"
	osName := schem.Get("os_type").(string)
	os, err := findBlockOSId(sess, osName)

	if err != nil {
		return nil, err
	}

	//This limits one to only ordering Iscsi, need to figure out how to do File type as well.
	performStorContainer := datatypes.Container_Product_Order_Network_Storage_Enterprise{
		Container_Product_Order: datatypes.Container_Product_Order{
			ComplexType: sl.String(CONTAINERS[tier]),
			Location:    sl.String(schem.Get("datacenter").(string)),
			PackageId:   sl.Int(PACKAGE_IDS[tier]),
			Prices:      prices,
			Quantity:    sl.Int(1),
		},
		OsFormatType: os,
	}
	return &performStorContainer, nil
}

/*
To extend this to support other NetworkStorages:
1. Find the categoryCodes for other NetworkStorages.

storeType is one of:
1. endurance
*/
func getPackagePrices(sess *session.Session, storeType string) (pack *datatypes.Product_Package, err error) {

	nasFilters := filter.New(
		filter.Path("categories.categoryCode").Eq(CATEGORY_CODES[storeType]),
	).Build()

	mask := "id,name,items[prices[categories],attributes]"

	if err != nil {
		return nil, err
	}

	packServ := services.GetProductPackageService(sess)
	prices, err := packServ.Id(PACKAGE_IDS[storeType]).
		Mask(mask).
		Filter(nasFilters).
		GetAllObjects()

	if err != nil {
		return nil, err
	}

	if len(prices) > 0 {
		return &prices[0], nil
	} else {
		return nil, errors.New("No Package Prices were returned from SoftLayer.")
	}
}

func resourceSoftLayerNetworkStorageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	service := services.GetNetworkStorageService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); !ok || apiErr.StatusCode != 404 {
			return false, fmt.Errorf("Error trying to retrieve Network Storage with ID %d : %s", id, err)
		}
	}

	return err == nil && result.Id != nil && *result.Id == id, nil
}

func resourceSoftLayerNetworkStorageDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkStorageService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	billingItem, err := service.Id(id).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error getting billing item for bare metal server: %s", err)
	}

	billingItemService := services.GetBillingItemService(sess)
	_, err = billingItemService.Id(*billingItem.Id).CancelItem(
		sl.Bool(true), sl.Bool(true), sl.String("No longer required"), sl.String("Please cancel this server"),
	)
	if err != nil {
		return fmt.Errorf("Error canceling the Network Storage (%d): %s", id, err)
	}

	return nil
}
