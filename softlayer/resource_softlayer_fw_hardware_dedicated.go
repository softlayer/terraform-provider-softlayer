package softlayer

import (
	"fmt"
	//	"log"
	"strconv"

	//	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	//	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"log"
	"time"
)

const (
	FwHardwareDedicatedPackageType = "ADDITIONAL_SERVICES_FIREWALL"

	fwMask = "firewallNetworkComponents,networkVlanFirewall.billingItem.orderItemId,dedicatedFirewallFlag" +
		",firewallGuestNetworkComponents,firewallInterfaces,firewallRules,highAvailabilityFirewallFlag"
)

func resourceSoftLayerFwHardwareDedicated() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerFwHardwareDedicatedCreate,
		Read:   resourceSoftLayerFwHardwareDedicatedRead,
		//		Update:   resourceSoftLayerFwHardwareDedicatedUpdate,
		Delete:   resourceSoftLayerFwHardwareDedicatedDelete,
		Exists:   resourceSoftLayerFwHardwareDedicatedExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"ha_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"public_vlan_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSoftLayerFwHardwareDedicatedCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	haEnabled := d.Get("ha_enabled").(bool)
	publicVlanId := d.Get("public_vlan_id").(int)

	var keyName string
	if haEnabled {
		keyName = "HARDWARE_FIREWALL_HIGH_AVAILABILITY"
	} else {
		keyName = "HARDWARE_FIREWALL_DEDICATED"
	}

	pkg, err := product.GetPackageByType(sess, FwHardwareDedicatedPackageType)
	if err != nil {
		return err
	}

	// Get all prices for ADDITIONAL_SERVICES_FIREWALL with the given capacity
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return err
	}

	// Select only those product items with a matching keyname
	targetItems := []datatypes.Product_Item{}
	for _, item := range productItems {
		if *item.KeyName == keyName {
			targetItems = append(targetItems, item)
		}
	}

	if len(targetItems) == 0 {
		return fmt.Errorf("No product items matching %s could be found", keyName)
	}

	productOrderContainer := datatypes.Container_Product_Order_Network_Protection_Firewall_Dedicated{
		Container_Product_Order: datatypes.Container_Product_Order{
			PackageId: pkg.Id,
			Prices: []datatypes.Product_Item_Price{
				{
					Id: targetItems[0].Prices[0].Id,
				},
			},
			Quantity: sl.Int(1),
		},
		VlanId: sl.Int(publicVlanId),
	}

	log.Println("[INFO] Creating dedicated hardware firewall")

	receipt, err := services.GetProductOrderService(sess).
		PlaceOrder(&productOrderContainer, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error during creation of dedicated hardware firewall: %s", err)
	}
	vlan, err := findDedicatedFirewallByOrderId(sess, *receipt.OrderId)

	d.SetId(fmt.Sprintf("%d", *vlan.NetworkVlanFirewall.Id))
	d.Set("ha_enabled", *vlan.HighAvailabilityFirewallFlag)
	d.Set("public_vlan_id", *vlan.Id)

	log.Printf("[INFO] Load Balancer ID: %s", d.Id())

	return resourceSoftLayerFwHardwareDedicatedRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	vipID, _ := strconv.Atoi(d.Id())

	certID := d.Get("security_certificate_id").(int)

	err := setLocalLBSecurityCert(sess, vipID, certID)

	if err != nil {
		return fmt.Errorf("Update load balancer failed: %s", err)
	}

	return resourceSoftLayerLbLocalRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	vipID, _ := strconv.Atoi(d.Id())

	vip, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		Mask(lbMask).
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.Set("connections", getConnectionLimit(*vip.ConnectionLimit))
	d.Set("datacenter", *vip.LoadBalancerHardware[0].Datacenter.Name)
	d.Set("ip_address", *vip.IpAddress.IpAddress)
	d.Set("subnet_id", *vip.IpAddress.SubnetId)
	d.Set("ha_enabled", *vip.HighAvailabilityFlag)
	d.Set("dedicated", *vip.DedicatedFlag)
	d.Set("ssl_enabled", *vip.SslEnabledFlag)

	// Optional fields.  Guard against nil pointer dereferences
	d.Set("security_certificate_id", sl.Get(vip.SecurityCertificateId, nil))

	return nil
}

func resourceSoftLayerFwHardwareDedicatedDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	vipService := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess)

	vipID, _ := strconv.Atoi(d.Id())

	var billingItem datatypes.Billing_Item_Network_LoadBalancer
	var err error

	// Get billing item associated with the load balancer
	if d.Get("dedicated").(bool) {
		billingItem, err = vipService.
			Id(vipID).
			GetDedicatedBillingItem()
	} else {
		billingItem.Billing_Item, err = vipService.
			Id(vipID).
			GetBillingItem()
	}

	if err != nil {
		return fmt.Errorf("Error while looking up billing item associated with the load balancer: %s", err)
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

func resourceSoftLayerFwHardwareDedicatedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

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

func findDedicatedFirewallByOrderId(sess *session.Session, orderId int) (datatypes.Network_Vlan, error) {
	filterPath := "networkVlans.networkVlanFirewall.billingItem.orderItemId"

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			vlans, err := services.GetAccountService(sess).
				Filter(filter.Build(
					filter.Path(filterPath).
						Eq(strconv.Itoa(orderId)))).
				Mask(fwMask).
				GetNetworkVlans()
			if err != nil {
				return datatypes.Network_Vlan{}, "", err
			}

			if len(vlans) == 1 {
				return vlans[0], "complete", nil
			} else if len(vlans) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one dedicated firewall: %s", err)
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Vlan{}, err
	}

	var result, ok = pendingResult.(datatypes.Network_Vlan)

	if ok {
		return result, nil
	}

	return datatypes.Network_Vlan{},
		fmt.Errorf("Cannot find Dedicated Firewall with order id '%d'", orderId)
}
