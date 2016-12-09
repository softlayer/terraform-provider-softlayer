package softlayer

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"log"
	"time"
)

const (
	FwHardwareDedicatedPackageType = "ADDITIONAL_SERVICES_FIREWALL"

	vlanMask = "firewallNetworkComponents,networkVlanFirewall.billingItem.orderItem.order.id,dedicatedFirewallFlag" +
		",firewallGuestNetworkComponents,firewallInterfaces,firewallRules,highAvailabilityFirewallFlag"
	fwMask = "id,networkVlan.highAvailabilityFirewallFlag"
)

func resourceSoftLayerFwHardwareDedicated() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerFwHardwareDedicatedCreate,
		Read:     resourceSoftLayerFwHardwareDedicatedRead,
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

	keyName := "HARDWARE_FIREWALL_DEDICATED"
	if haEnabled {
		keyName = "HARDWARE_FIREWALL_HIGH_AVAILABILITY"
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
	if err != nil {
		return fmt.Errorf("Error during creation of dedicated hardware firewall: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *vlan.NetworkVlanFirewall.Id))
	d.Set("ha_enabled", *vlan.HighAvailabilityFirewallFlag)
	d.Set("public_vlan_id", *vlan.Id)

	log.Printf("[INFO] Firewall ID: %s", d.Id())

	return resourceSoftLayerFwHardwareDedicatedRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	fwID, _ := strconv.Atoi(d.Id())

	fw, err := services.GetNetworkVlanFirewallService(sess).
		Id(fwID).
		Mask(fwMask).
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving firewall information: %s", err)
	}

	d.Set("public_vlan_id", *fw.NetworkVlan.Id)
	d.Set("ha_enabled", *fw.NetworkVlan.HighAvailabilityFirewallFlag)

	return nil
}

func resourceSoftLayerFwHardwareDedicatedDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	fwService := services.GetNetworkVlanFirewallService(sess)

	fwID, _ := strconv.Atoi(d.Id())

	// Get billing item associated with the firewall
	billingItem, err := fwService.Id(fwID).GetBillingItem()

	if err != nil {
		return fmt.Errorf("Error while looking up billing item associated with the firewall: %s", err)
	}

	if billingItem.Id == nil {
		return fmt.Errorf("Error while looking up billing item associated with the firewall: No billing item for ID:%d", fwID)
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

	fwID, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = services.GetNetworkVlanFirewallService(sess).
		Id(fwID).
		GetObject()

	if err != nil {
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving firewall information: %s", err)
	}

	return true, nil
}

func findDedicatedFirewallByOrderId(sess *session.Session, orderId int) (datatypes.Network_Vlan, error) {
	filterPath := "networkVlans.networkVlanFirewall.billingItem.orderItem.order.id"

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			vlans, err := services.GetAccountService(sess).
				Filter(filter.Build(
					filter.Path(filterPath).
						Eq(strconv.Itoa(orderId)))).
				Mask(vlanMask).
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
		Timeout:    45 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
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
