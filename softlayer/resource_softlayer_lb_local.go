package softlayer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/filter"
	"github.ibm.com/riethm/gopherlayer.git/helpers/location"
	"github.ibm.com/riethm/gopherlayer.git/helpers/product"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
	"github.ibm.com/riethm/gopherlayer.git/sl"
)

const (
	LB_LARGE_150000_CONNECTIONS = 150000
	LB_SMALL_15000_CONNECTIONS  = 15000

	LbLocalPackageType = "ADDITIONAL_SERVICES_LOAD_BALANCER"

	lbMask = "id,connectionLimit,ipAddressId,securityCertificateId,highAvailabilityFlag," +
		"sslEnabledFlag,loadBalancerHardware[datacenter[name]],ipAddress[ipAddress,subnetId]"
)

func resourceSoftLayerLbLocal() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerLbLocalCreate,
		Read:   resourceSoftLayerLbLocalRead,
		Update: resourceSoftLayerLbLocalUpdate,
		Delete: resourceSoftLayerLbLocalDelete,
		Exists: resourceSoftLayerLbLocalExists,

		Schema: map[string]*schema.Schema{
			"connections": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ha_enabled": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"security_certificate_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceSoftLayerLbLocalCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	connections := d.Get("connections").(int)
	haEnabled := d.Get("ha_enabled").(bool)

	// SoftLayer capacities don't match the published capacities as seen in the local lb
	// ordering screen in the customer portal. Terraform exposes the published capacities.
	// Create a translation map for those cases where the published capacity does not
	// equal the actual actual capacity on the product_item.
	capacities := map[int]float64{
		15000:  65000.0,
		150000: 130000.0,
	}

	var capacity float64
	if c, ok := capacities[connections]; !ok {
		capacity = float64(connections)
	} else {
		capacity = c
	}

	var keyname string
	if haEnabled {
		keyname = "DEDICATED_LOAD_BALANCER_WITH_HIGH_AVAILABILITY_AND_SSL"
	} else {
		keyname = "LOAD_BALANCER_DEDICATED_WITH_SSL_OFFLOAD"
	}

	keyname = strings.Join([]string{keyname, strconv.Itoa(connections), "CONNECTIONS"}, "_")

	pkg, err := product.GetPackageByType(sess, LbLocalPackageType)
	if err != nil {
		return err
	}

	// Get all prices for ADDITIONAL_SERVICE_LOAD_BALANCER with the given capacity
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return err
	}

	// Select only those product items with a matching keyname
	targetItems := []datatypes.Product_Item{}
	for _, item := range productItems {
		if *item.KeyName == keyname {
			targetItems = append(targetItems, item)
		}
	}

	if len(targetItems) == 0 {
		return fmt.Errorf("No product items matching %s could be found", keyname)
	}

	//select prices with the required capacity
	prices := product.SelectProductPricesByCategory(
		targetItems,
		map[string]float64{
			product.DedicatedLoadBalancerCategoryCode: capacity,
		},
	)

	// Lookup the datacenter ID
	dc, err := location.GetDatacenterByName(sess, d.Get("location").(string))

	productOrderContainer := datatypes.Container_Product_Order_Network_LoadBalancer{
		Container_Product_Order: datatypes.Container_Product_Order{
			PackageId: pkg.Id,
			Location:  sl.String(strconv.Itoa(*dc.Id)),
			Prices:    prices[:1],
			Quantity:  sl.Int(1),
		},
	}

	log.Println("[INFO] Creating load balancer")

	receipt, err := services.GetProductOrderService(sess).
		PlaceOrder(&productOrderContainer, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error during creation of load balancer: %s", err)
	}

	loadBalancer, err := findLoadBalancerByOrderId(sess, *receipt.OrderId)

	d.SetId(fmt.Sprintf("%d", *loadBalancer.Id))
	d.Set("connections", getConnectionLimit(*loadBalancer.ConnectionLimit))
	d.Set("location", *loadBalancer.LoadBalancerHardware[0].Datacenter.Name)
	d.Set("ip_address", *loadBalancer.IpAddress.IpAddress)
	d.Set("subnet_id", *loadBalancer.IpAddress.SubnetId)
	d.Set("ha_enabled", *loadBalancer.HighAvailabilityFlag)

	log.Printf("[INFO] Load Balancer ID: %s", d.Id())

	return resourceSoftLayerLbLocalUpdate(d, meta)
}

func resourceSoftLayerLbLocalUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipID, _ := strconv.Atoi(d.Id())

	vip := datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{
		SecurityCertificateId: sl.Int(d.Get("security_certificate_id").(int)),
	}

	success, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		EditObject(&vip)

	if err != nil {
		return fmt.Errorf("Update load balancer failed: %s", err)
	}

	if !success {
		return fmt.Errorf("Update load balancer failed: %s", err)
	}

	return resourceSoftLayerLbLocalRead(d, meta)
}

func resourceSoftLayerLbLocalRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipID, _ := strconv.Atoi(d.Id())

	vip, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		Mask(lbMask).
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.Set("connections", getConnectionLimit(*vip.ConnectionLimit))
	d.Set("location", *vip.LoadBalancerHardware[0].Datacenter.Name)
	d.Set("ip_address", *vip.IpAddress.IpAddress)
	d.Set("subnet_id", *vip.IpAddress.SubnetId)
	d.Set("ha_enabled", *vip.HighAvailabilityFlag)

	// Optional fields.  Guard against nil pointer dereferences
	if vip.SecurityCertificateId == nil {
		d.Set("security_certificate_id", nil)
	} else {
		d.Set("security_certificate_id", *vip.SecurityCertificateId)
	}

	return nil
}

func resourceSoftLayerLbLocalDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipID, _ := strconv.Atoi(d.Id())

	// Get billing item associated with the load balancer
	bi, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		GetDedicatedBillingItem()

	if err != nil {
		return fmt.Errorf("Error while looking up billing item associated with the load balancer: %s", err)
	}

	return cancelService(sess, *bi.Id)
}

func resourceSoftLayerLbLocalExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)

	vipID, _ := strconv.Atoi(d.Id())

	_, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		Mask("id").
		GetObject()

	if err != nil {
		return false, err
	}

	return true, nil
}

/* When requesting 15000 SL creates between 15000 and 150000. When requesting 150000 SL creates >= 150000 */
func getConnectionLimit(connectionLimit int) int {
	if connectionLimit >= LB_LARGE_150000_CONNECTIONS {
		return LB_LARGE_150000_CONNECTIONS
	} else if connectionLimit >= LB_SMALL_15000_CONNECTIONS &&
		connectionLimit < LB_LARGE_150000_CONNECTIONS {
		return LB_SMALL_15000_CONNECTIONS
	} else {
		return 0
	}
}

func cancelService(sess *session.Session, billingId int) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"transactions_pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			success, err := services.GetBillingItemService(sess).Id(billingId).CancelService()

			if err != nil {
				if apiErr, ok := err.(sl.Error); ok {
					// TODO this logic depends too heavily on localized strings which could be translated.
					// Need to change this to check for pending transactions in a wait loop, and then
					// cancel the service once, reporting any success/error
					if strings.Index(apiErr.Message, "There is currently an active transaction") != -1 {
						return false, "transactions_pending", nil
					}
				}
				return success, "error", err
			}

			return success, "complete", nil
		},
		Timeout:    10 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return err
	}

	if pendingResult != nil && !(pendingResult.(bool)) {
		return errors.New("SoftLayer reported an unsuccessful cancellation, but did not provide a reason.")
	}

	return nil
}

func findLoadBalancerByOrderId(sess *session.Session, orderId int) (datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			lbs, err := services.GetAccountService(sess).
				Filter(filter.Build(
					filter.Path("adcLoadBalancers.dedicatedBillingItem.orderItem.order.id").
						Eq(strconv.Itoa(orderId)))).
				Mask(lbMask).
				GetAdcLoadBalancers()
			if err != nil {
				return datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{}, "", err
			}

			if len(lbs) == 1 {
				return lbs[0], "complete", nil
			} else if len(lbs) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one load balancer: %s", err)
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{}, err
	}

	var result, ok = pendingResult.(datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress)

	if ok {
		return result, nil
	}

	return datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{},
		fmt.Errorf("Cannot find Application Delivery Controller Load Balancer with order id '%d'", orderId)
}
