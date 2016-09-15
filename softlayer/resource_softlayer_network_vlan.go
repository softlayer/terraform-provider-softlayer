package softlayer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/hardware"
	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"log"
	"time"
)

const (
	NetworkVlanPackageType = "ADDITIONAL_SERVICES_NETWORK_VLAN"
)

func resourceSoftLayerNetworkVlan() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerNetworkVlanCreate,
		Read:     resourceSoftLayerNetworkVlanRead,
		Update:   resourceSoftLayerNetworkVlanUpdate,
		Delete:   resourceSoftLayerNetworkVlanDelete,
		Exists:   resourceSoftLayerNetworkVlanExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"primary_subnet_size": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"primary_router_hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"vlan_number": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"softlayer_managed": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"child_resource_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnets": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"subnet_type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceSoftLayerNetworkVlanCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	var err error
	var dc datatypes.Location_Datacenter
	var rt datatypes.Hardware

	name := d.Get("name").(string)
	router := d.Get("primary_router_hostname").(string)
	datacenter := d.Get("datacenter").(string)

	if len(datacenter) > 0 {
		dc, err = location.GetDatacenterByName(sess, datacenter, "id")
		if err != nil {
			return fmt.Errorf("Error creating network vlan: %s", err)
		}
	} else {
		return fmt.Errorf("Error creating network vlan: datacenter name is empty.")
	}

	// 1. Get a package for ADDITIONAL_SERVICES_NETWORK_VLAN
	pkg, err := product.GetPackageByType(sess, NetworkVlanPackageType)
	if err != nil {
		return err
	}

	// 2. Get all prices for the package
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return err
	}

	// 3. Validate vlanType field
	vlanType := d.Get("type").(string)
	if vlanType != "PRIVATE" && vlanType != "PUBLIC" {
		return fmt.Errorf("Error creating network vlan: vlanType should be either 'PRIVATE' or 'PUBLIC'")
	}
	if (vlanType == "PRIVATE" && len(router) > 0 && strings.Contains(router, "fcr")) ||
		(vlanType == "PUBLIC" && len(router) > 0 && strings.Contains(router, "bcr")) {
		return fmt.Errorf("Error creating network vlan: mismatch between vlan_type '%s' and primary_router_hostname '%s'", vlanType, router)
	}

	// 4. Find vlan and subnet prices
	vlanKeyname := vlanType + "_NETWORK_VLAN"
	subnetKeyname := strconv.Itoa(d.Get("primary_subnet_size").(int)) + "_STATIC_PUBLIC_IP_ADDRESSES"

	// 5. Select items with a matching keyname
	vlanItems := []datatypes.Product_Item{}
	subnetItems := []datatypes.Product_Item{}
	for _, item := range productItems {
		if *item.KeyName == vlanKeyname {
			vlanItems = append(vlanItems, item)
		}
		if strings.Contains(*item.KeyName, subnetKeyname) {
			subnetItems = append(subnetItems, item)
		}
	}

	if len(vlanItems) == 0 {
		return fmt.Errorf("No product items matching %s could be found", vlanKeyname)
	}

	if len(subnetItems) == 0 {
		return fmt.Errorf("No product items matching %s could be found", subnetKeyname)
	}

	productOrderContainer := datatypes.Container_Product_Order_Network_Vlan{
		Container_Product_Order: datatypes.Container_Product_Order{
			PackageId: pkg.Id,
			Location:  sl.String(strconv.Itoa(*dc.Id)),
			Prices: []datatypes.Product_Item_Price{
				{
					Id: vlanItems[0].Prices[0].Id,
				},
				{
					Id: subnetItems[0].Prices[0].Id,
				},
			},
			Quantity: sl.Int(1),
		},
	}

	if len(router) > 0 {
		rt, err = hardware.GetRouterByName(sess, router, "id")
		productOrderContainer.RouterId = rt.Id
		if err != nil {
			return fmt.Errorf("Error creating network vlan: %s", err)
		}
	}

	log.Println("[INFO] Creating network vlan")

	receipt, err := services.GetProductOrderService(sess).
		PlaceOrder(&productOrderContainer, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error during creation of network vlan: %s", err)
	}

	vlan, err := findNetworkVlanByOrderId(sess, *receipt.OrderId)

	if len(name) > 0 {
		_, err = services.GetNetworkVlanService(sess).
			Id(*vlan.Id).EditObject(&datatypes.Network_Vlan{Name: sl.String(name)})
		if err != nil {
			return fmt.Errorf("Error updating Network Vlan: %s", err)
		}
	}

	d.SetId(fmt.Sprintf("%d", *vlan.Id))
	return resourceSoftLayerNetworkVlanRead(d, meta)
}

func resourceSoftLayerNetworkVlanRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	vlan, err := service.Id(vlanId).Mask(
		"id," +
			"name," +
			"primaryRouter[datacenter[name]]," +
			"primaryRouter[hostname]," +
			"vlanNumber," +
			"billingItem[recurringFee]," +
			"guestNetworkComponentCount," +
			"subnets[networkIdentifier,cidr,subnetType]",
	).GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving Network Vlan: %s", err)
	}

	d.Set("id", *vlan.Id)
	d.Set("vlan_number", *vlan.VlanNumber)
	d.Set("child_resource_count", *vlan.GuestNetworkComponentCount)
	if vlan.Name != nil {
		d.Set("name", *vlan.Name)
	} else {
		d.Set("name", "")
	}

	if vlan.PrimaryRouter != nil {
		d.Set("primary_router_hostname", *vlan.PrimaryRouter.Hostname)
		if strings.HasPrefix(*vlan.PrimaryRouter.Hostname, "fcr") {
			d.Set("type", "PUBLIC")
		} else {
			d.Set("type", "PRIVATE")
		}
		if vlan.PrimaryRouter.Datacenter != nil {
			d.Set("datacenter", *vlan.PrimaryRouter.Datacenter.Name)
		}
	}

	if vlan.BillingItem != nil {
		d.Set("softlayer_managed", true)
	} else {
		d.Set("softlayer_managed", false)
	}

	// Subnets
	subnets := make([]map[string]interface{}, 0)

	for _, elem := range vlan.Subnets {
		subnet := make(map[string]interface{})
		subnet["subnet"] = *elem.NetworkIdentifier + "/" + strconv.Itoa(*elem.Cidr)
		subnet["subnet_type"] = *elem.SubnetType
		subnets = append(subnets, subnet)
	}
	d.Set("subnets", subnets)

	if vlan.Subnets != nil && len(vlan.Subnets) > 0 {
		d.Set("primary_subnet_size", 1<<(uint)(32-*vlan.Subnets[0].Cidr))
	} else {
		d.Set("primary_subnet_size", 0)
	}

	return nil
}

func resourceSoftLayerNetworkVlanUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	opts := datatypes.Network_Vlan{}

	if d.HasChange("name") {
		opts.Name = sl.String(d.Get("name").(string))
	}

	_, err = service.Id(vlanId).EditObject(&opts)

	if err != nil {
		return fmt.Errorf("Error updating Network Vlan: %s", err)
	}
	return resourceSoftLayerNetworkVlanRead(d, meta)
}

func resourceSoftLayerNetworkVlanDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	// Check if the VLAN is existed.
	_, err = service.Id(vlanId).Mask("id").GetObject()
	if err != nil {
		return fmt.Errorf("Error deleting Network Vlan: %s", err)
	}

	billingItem, err := service.Id(vlanId).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error deleting Network Vlan: %s", err)
	}
	if billingItem.Id == nil {
		return nil
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

func resourceSoftLayerNetworkVlanExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkVlanService(sess)

	vlanId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid vlan ID, must be an integer: %s", err)
	}

	_, err = service.Id(vlanId).Mask("id").GetObject()

	if err != nil {
		return false, err
	}

	return true, nil
}

func findNetworkVlanByOrderId(sess *session.Session, orderId int) (datatypes.Network_Vlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			vlans, err := services.GetAccountService(sess).
				Filter(filter.Build(
				filter.Path("networkVlans.billingItem.orderItem.order.id").
					Eq(strconv.Itoa(orderId)))).
				Mask("id").
				GetNetworkVlans()
			if err != nil {
				return datatypes.Network_Vlan{}, "", err
			}

			if len(vlans) == 1 {
				return vlans[0], "complete", nil
			} else if len(vlans) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one network vlan: %s", err)
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
		fmt.Errorf("Cannot find Network Vlan with order id '%d'", orderId)
}
