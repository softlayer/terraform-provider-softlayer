package softlayer

import (
	"errors"
	"fmt"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const nasType string = "block"

var (
	packageIds = map[string]int{
		"endurance": 240,
	}
	containers = map[string]string{
		"endurance": "Softlayer_Container_Product_Order_Network_Storage_Enterprise",
	}
	categoryCodes = map[string]string{
		"endurance": "storage_service_enterprise",
		"block":     "storage_block",
	}
	enduranceTiers = map[float64]int{
		0.25: 100,
		2:    200,
		4:    300,
		10:   1000,
	}
)

func resourceSoftLayerBlockStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerBlockStorageCreate,
		Read:   resourceSoftLayerBlockStorageRead,
		Delete: resourceSoftLayerBlockStorageDelete,
		Exists: resourceSoftLayerBlockStorageExists,
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
	var (
		retry       = "retry"
		pending     = "pending"
		provisioned = "provisioned"
	)
	stateConf := &resource.StateChangeConf{
		Pending: []string{retry, pending},
		Target:  []string{provisioned},
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
				return false, retry, err
			}

			if len(stores) == 0 || stores[0].CreateDate == nil {
				return nil, pending, nil
			}

			return stores[0], provisioned, nil

		},
		Timeout:    10 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 5 * time.Minute,
	}

	return stateConf.WaitForState()
}

func resourceSoftLayerBlockStorageRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetNetworkStorageService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).GetObject()

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
	var datacenters []datatypes.Location

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

func findPrice(category string, items []datatypes.Product_Item) (*datatypes.Product_Item_Price, error) {
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

func resourceSoftLayerBlockStorageCreate(schem *schema.ResourceData, meta interface{}) error {
	/*Variables are declared here in concordance with their dependencies.
	The root dependency is the Product_Package, followed by the prices
	that are gathered from that Product_Package.
	*/
	var wg sync.WaitGroup
	var pack *datatypes.Product_Package
	var loc *string
	var tierPrice, nasTypePrice, iopsPrice, spacePrice *datatypes.Product_Item_Price
	var prices []datatypes.Product_Item_Price

	errors := make(chan error, 1)  //catches errors.
	finished := make(chan bool, 1) //signifies doneness.
	iopsCategory := enduranceTiers[schem.Get("iops").(float64)]
	sess := meta.(*session.Session)
	tier := schem.Get("tier").(string)
	capacity := schem.Get("capacity_gb").(int)

	//closure function needs to be higher order.
	waitGroup1 := []func(){
		func() {
			var err error
			defer wg.Done()
			loc, err = getLocationID(sess, schem.Get("datacenter").(string))
			if err != nil {
				errors <- err
			}
		},
		func() {
			var err error
			defer wg.Done()
			pack, err = getPackagePrices(sess, tier)
			if err != nil {
				errors <- err
			}
		},
	}

	//https://golang.org/pkg/sync/#WaitGroup
	for _, call := range waitGroup1 {
		wg.Add(1)
		go call() //can do a better job of parameterizing functions.
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	// either continues or returns the error.
	select {
	case <-finished:
	case err := <-errors:
		if err != nil {
			return err
		}
	}

	finished = make(chan bool, 1) //signifies doneness.

	waitGroup2 := []func(){
		//TIER prices
		func() {
			var err error
			defer wg.Done()
			tierPrice, err = findPrice(categoryCodes[tier], pack.Items)
			if err != nil {
				errors <- err
			}
			prices = append(prices, *tierPrice)
		},
		func() {
			var err error
			defer wg.Done()
			nasTypePrice, err = findPrice(categoryCodes[nasType], pack.Items)

			if err != nil {
				errors <- err
			}

			prices = append(prices, *nasTypePrice)
		},
		func() {
			//IOPS prices.
			var err error
			defer wg.Done()

			iopsPrice, err = findIOPSPrice(iopsCategory, pack.Items)

			if err != nil {
				errors <- err
			}
			prices = append(prices, *iopsPrice)
		},
	}

	for _, priceSearch := range waitGroup2 {
		wg.Add(1)
		go priceSearch()
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err := <-errors:
		if err != nil {
			return err
		}
	}

	//SPACE prices
	spacePrice, err := findSpacePrice(iopsCategory, capacity, pack.Items)

	if err != nil {
		return err
	}

	prices = append(prices, *spacePrice)

	//SNAPSHOT SPACE prices.

	//Build Order
	order, err := buildOrder(sess, prices, schem, loc)

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

	if err != nil {
		return err
	}

	id := strconv.Itoa(*store.(datatypes.Network_Storage).
		Id)

	schem.SetId(id)

	return resourceSoftLayerBlockStorageRead(schem, sess)
}

func buildOrder(sess *session.Session, prices []datatypes.Product_Item_Price, schem *schema.ResourceData, loc *string) (order *datatypes.Container_Product_Order_Network_Storage_Enterprise, err error) {
	tier := schem.Get("tier").(string)
	osName := schem.Get("os_type").(string)
	os, err := findBlockOSId(sess, osName)

	if err != nil {
		return nil, err
	}

	//This limits one to only ordering Iscsi, need to figure out how to do File type as well.
	orderContainer := datatypes.Container_Product_Order_Network_Storage_Enterprise{
		Container_Product_Order: datatypes.Container_Product_Order{
			ComplexType: sl.String(containers[tier]),
			Location:    loc,
			PackageId:   sl.Int(packageIds[tier]),
			Prices:      prices,
			Quantity:    sl.Int(1),
		},
		OsFormatType: os,
	}
	return &orderContainer, nil
}

/*
To extend this to support other BlockStorages:
1. Find the categoryCodes for other BlockStorages.

storeType is one of:
1. endurance
*/
func getPackagePrices(sess *session.Session, storeType string) (pack *datatypes.Product_Package, err error) {
	var packages []datatypes.Product_Package

	nasFilters := filter.New(
		filter.Path("categories.categoryCode").Eq(categoryCodes[storeType]),
	).Build()

	mask := "id,name,items[prices[categories],attributes]"

	if err != nil {
		return nil, err
	}

	packServ := services.GetProductPackageService(sess)
	packages, err = packServ.Id(packageIds[storeType]).
		Mask(mask).
		Filter(nasFilters).
		GetAllObjects()

	if err != nil {
		return nil, err
	}

	if len(packages) > 0 {
		return &packages[0], nil
	}
	return nil, errors.New("No Package Prices were returned from SoftLayer.")

}

func resourceSoftLayerBlockStorageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
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

func resourceSoftLayerBlockStorageDelete(d *schema.ResourceData, meta interface{}) error {
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
