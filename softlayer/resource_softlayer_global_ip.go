package softlayer

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const (
	AdditionalServicesGlobalIpAddressesPackageType = "ADDITIONAL_SERVICES_GLOBAL_IP_ADDRESSES"

	GlobalIpMask = "id,ipAddress[ipAddress],destinationIpAddress[ipAddress]"
)

func resourceSoftLayerGlobalIp() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerGlobalIpCreate,
		Read:     resourceSoftLayerGlobalIpRead,
		Update:   resourceSoftLayerGlobalIpUpdate,
		Delete:   resourceSoftLayerGlobalIpDelete,
		Exists:   resourceSoftLayerGlobalIpExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"routes_to": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					address := v.(string)
					if net.ParseIP(address) == nil {
						errors = append(errors, fmt.Errorf("Invalid IP format: %s", address))
					}
					return
				},
				DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
					newRoutesTo := net.ParseIP(n)
					// Return true when n has the appropriate IPv6 format and
					// the compressed value of n equals the compressed value of o.
					return newRoutesTo != nil && (newRoutesTo.String() == net.ParseIP(o).String())
				},
			},
		},
	}
}

func resourceSoftLayerGlobalIpCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	// Find price items with AdditionalServicesGlobalIpAddresses
	productOrderContainer, err := buildGlobalIpProductOrderContainer(d, sess, AdditionalServicesGlobalIpAddressesPackageType)
	if err != nil {
		// Find price items with AdditionalServices
		productOrderContainer, err = buildGlobalIpProductOrderContainer(d, sess, AdditionalServicesPackageType)
		if err != nil {
			return fmt.Errorf("Error creating global ip: %s", err)
		}
	}

	log.Println("[INFO] Creating global ip")

	receipt, err := services.GetProductOrderService(sess).
		PlaceOrder(productOrderContainer, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error during creation of global ip: %s", err)
	}

	globalIp, err := findGlobalIpByOrderId(sess, *receipt.OrderId)
	if err != nil {
		return fmt.Errorf("Error during creation of global ip: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *globalIp.Id))
	d.Set("ip_address", *globalIp.IpAddress.IpAddress)

	return resourceSoftLayerGlobalIpUpdate(d, meta)
}

func resourceSoftLayerGlobalIpRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetIpAddressGlobalService(sess)

	globalIpId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid global ip ID, must be an integer: %s", err)
	}

	globalIp, err := service.Id(globalIpId).Mask(GlobalIpMask).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving Global Ip: %s", err)
	}

	d.Set("id", *globalIp.Id)
	d.Set("ip_address", *globalIp.IpAddress.IpAddress)
	if globalIp.DestinationIpAddress != nil {
		d.Set("routes_to", *globalIp.DestinationIpAddress.IpAddress)
	}
	return nil
}

func resourceSoftLayerGlobalIpUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetIpAddressGlobalService(sess)

	globalIpId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid global ip ID, must be an integer: %s", err)
	}

	routes_to := d.Get("routes_to").(string)
	if strings.Contains(routes_to, ":") && len(routes_to) != 39 {
		parts := strings.Split(routes_to, ":")
		for x, s := range parts {
			if s == "" {
				zeroes := 9 - len(parts)
				parts[x] = strings.Repeat("0000:", zeroes)[:(zeroes*4)+(zeroes-1)]
			} else {
				parts[x] = fmt.Sprintf("%04s", s)
			}
		}

		routes_to = strings.Join(parts, ":")
		d.Set("routes_to", routes_to)
	}

	_, err = service.Id(globalIpId).Route(sl.String(routes_to))
	if err != nil {
		return fmt.Errorf("Error editing Global Ip: %s", err)
	}
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			transaction, err := service.Id(globalIpId).GetActiveTransaction()
			if err != nil {
				return datatypes.Network_Subnet_IpAddress_Global{}, "pending", err
			}
			if transaction.Id == nil {
				return datatypes.Network_Subnet_IpAddress_Global{}, "complete", nil
			}
			return datatypes.Network_Subnet_IpAddress_Global{}, "pending", nil
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("Error waiting for global ip destination ip address to become active: %s", err)
	}

	if _, ok := pendingResult.(datatypes.Network_Subnet_IpAddress_Global); ok {
		return nil
	}

	return nil
}

func resourceSoftLayerGlobalIpDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetIpAddressGlobalService(sess)

	globalIpId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid global ip ID, must be an integer: %s", err)
	}

	billingItem, err := service.Id(globalIpId).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error deleting global ip: %s", err)
	}

	if billingItem.Id == nil {
		return nil
	}

	_, err = services.GetBillingItemService(sess).Id(*billingItem.Id).CancelService()

	return err
}

func resourceSoftLayerGlobalIpExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetIpAddressGlobalService(sess)

	globalIpId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(globalIpId).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving global ip: %s", err)
	}
	return result.Id != nil && *result.Id == globalIpId, nil
}

func findGlobalIpByOrderId(sess *session.Session, orderId int) (datatypes.Network_Subnet_IpAddress_Global, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			globalIps, err := services.GetAccountService(sess).
				Filter(filter.Path("globalIpRecords.billingItem.orderItem.order.id").
					Eq(strconv.Itoa(orderId)).Build()).
				Mask("id,ipAddress[ipAddress]").
				GetGlobalIpRecords()
			if err != nil {
				return datatypes.Network_Subnet_IpAddress_Global{}, "", err
			}

			if len(globalIps) == 1 && globalIps[0].IpAddress != nil {
				return globalIps[0], "complete", nil
			} else if len(globalIps) == 0 || len(globalIps) == 1 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one global ip: %s", err)
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Subnet_IpAddress_Global{}, err
	}

	if result, ok := pendingResult.(datatypes.Network_Subnet_IpAddress_Global); ok {
		return result, nil
	}

	return datatypes.Network_Subnet_IpAddress_Global{},
		fmt.Errorf("Cannot find global ip with order id '%d'", orderId)
}

func buildGlobalIpProductOrderContainer(d *schema.ResourceData, sess *session.Session, packageType string) (
	*datatypes.Container_Product_Order_Network_Subnet, error) {

	// 1. Get a package
	pkg, err := product.GetPackageByType(sess, packageType)
	if err != nil {
		return &datatypes.Container_Product_Order_Network_Subnet{}, err
	}

	// 2. Get all prices for the package
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return &datatypes.Container_Product_Order_Network_Subnet{}, err
	}

	// 3. Find global ip prices
	globalIpKeyname := "GLOBAL_IPV4"
	if strings.Contains(d.Get("routes_to").(string), ":") {
		globalIpKeyname = "GLOBAL_IPV6"
	}

	// 4. Select items with a matching keyname
	globalIpItems := []datatypes.Product_Item{}
	for _, item := range productItems {
		if *item.KeyName == globalIpKeyname {
			globalIpItems = append(globalIpItems, item)
		}
	}

	if len(globalIpItems) == 0 {
		return &datatypes.Container_Product_Order_Network_Subnet{},
			fmt.Errorf("No product items matching %s could be found", globalIpKeyname)
	}

	productOrderContainer := datatypes.Container_Product_Order_Network_Subnet{
		Container_Product_Order: datatypes.Container_Product_Order{
			PackageId: pkg.Id,
			Prices: []datatypes.Product_Item_Price{
				{
					Id: globalIpItems[0].Prices[0].Id,
				},
			},
			Quantity: sl.Int(1),
		},
	}

	return &productOrderContainer, nil
}
