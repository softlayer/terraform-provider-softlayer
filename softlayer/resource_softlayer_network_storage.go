package softlayer

import (
	"encoding/json"
	"errors"
	"fmt"
	//"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	//"log"
	"strconv"
	"strings"
	//"time"
)

var (
	PACKAGE_IDS = map[string]int{
		"endurance": 240,
	}
	CONTAINERS = map[string]string{
		"endurance": "Softlayer_Container_Product_Order_Network_EnduranceStorage_Iscsi",
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
		//		Delete: resourceSoftLayerEnduranceStorageDelete,
		//		Exists: resourceSoftLayerEnduranceStorageExists,
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
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},
			"hourly_pricing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
		},
	}
}

func resourceSoftLayerNetworkStorageRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetNetworkStorageService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	mask := strings.Join([]string{
		"allowableHardware",
		"allowableVirtualGuests",
		"allowedHardware",
		"allowedIpAddresses",
		"allowedReplicationHardware",
		"allowedReplicationIpAddresses",
		"allowedReplicationSubnets",
		"allowedReplicationVirtualGuests",
		"allowedSubnets",
		"allowedVirtualGuests",
		"bytesUsed",
		"credentials",
		"dailySchedule",
		"hardware",
		"hasEncryptionAtRest",
		"hourlySchedule",
		"iops",
		"lunId",
		"mountableFlag",
		"osType",
		"osTypeId",
		"virtualGuest",
		"hourlySchedule",
		"weeklySchedule",
	}, ",")

	result, err := service.Id(id).Mask(mask).GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving storage volume: %s", err)
	}

	d.Set("id", *result.Id)
	d.Set("capacity_gb", *result.CapacityGb)
	d.Set("create_date", *result.CreateDate)

	return nil
}

func hasCategory(categoryCode string, categories []datatypes.Product_Item_Category) (result bool) {
	result = false

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
		for _, attribute := range item.Attributes {
			value64, err := strconv.ParseInt(*attribute.Value, 10, 0)

			if err != nil {
				return nil, err
			}

			value := int(value64)
			if value == tier {
				break
			} else {
				continue
			}

		}

		for _, price := range item.Prices {
			if price.LocationGroupId != nil {
				continue
			}

			if !hasCategory(category, price.Categories) {
				continue
			}
			return &price, nil
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

	pricesJSON, err := json.Marshal(prices)

	if err != nil {
		return err
	}

	fmt.Printf("%q\n", pricesJSON)

	//TODO: SNAPSHOT SPACE prices.

	//Build Order
	order, err := buildOrder(prices, schem)
	//TODO: Verify Order
	orderService := services.GetProductOrderService(sess)
	verify, err := orderService.VerifyOrder(&order)

	if err != nil {
		return err
	}

	verifyJSON, err := json.Marshal(verify)

	if err != nil {
		return err
	}

	fmt.Printf("%q\n", verifyJSON)

	//TODO: SNAPSHOT SCHEDULES

	return nil
}

func buildOrder(prices []datatypes.Product_Item_Price, schem *schema.ResourceData) (order datatypes.Container_Product_Order_Network_PerformanceStorage_Iscsi, err error) {
	hourlyPricing := sl.Bool(schem.Get("hourly_pricing").(bool))
	tier := "endurance"

	//This limits one to only ordering Iscsi, need to figure out how to do File type as well.
	performStorContainer := datatypes.Container_Product_Order_Network_PerformanceStorage_Iscsi{
		Container_Product_Order_Network_PerformanceStorage: datatypes.Container_Product_Order_Network_PerformanceStorage{
			Container_Product_Order: datatypes.Container_Product_Order{
				ComplexType:      sl.String(CONTAINERS[tier]),
				Location:         sl.String(schem.Get("datacenter").(string)),
				PackageId:        sl.Int(PACKAGE_IDS[tier]),
				Prices:           prices,
				Quantity:         sl.Int(1),
				UseHourlyPricing: hourlyPricing,
			},
		},
	}

	return performStorContainer, nil
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
