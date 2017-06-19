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
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const (
	SubnetMask = "id,addressSpace,subnetType,version,ipAddressCount," +
		"networkIdentifier,cidr,note,endPointIpAddress[ipAddress],networkVlan[id]"
)

var (
	// Map subnet types to product package keyname in SoftLayer_Product_Item
	subnetPackageTypeMap = map[string]string{
		"Static":   "ADDITIONAL_SERVICES_STATIC_IP_ADDRESSES",
		"Portable": "ADDITIONAL_SERVICES_PORTABLE_IP_ADDRESSES",
	}

	// Map SL internal type code to subnet type.
	subnetTypeMap = map[string]string{
		"SECONDARY_ON_VLAN": "Portable",
		"ROUTED_TO_VLAN":    "Portable",
		"SUBNET_ON_VLAN":    "Portable",
		"STATIC_IP_ROUTED":  "Static",
	}
)

func resourceSoftLayerSubnet() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerSubnetCreate,
		Read:     resourceSoftLayerSubnetRead,
		Update:   resourceSoftLayerSubnetUpdate,
		Delete:   resourceSoftLayerSubnetDelete,
		Exists:   resourceSoftLayerSubnetExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"network": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					network := v.(string)
					if network != "PRIVATE" && network != "PUBLIC" {
						errs = append(errs, errors.New(
							"network should be either 'PRIVATE' or 'PUBLIC'"))
					}
					return
				},
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					typeStr := v.(string)
					if typeStr != "Portable" && typeStr != "Static" {
						errs = append(errs, errors.New(
							"type should be either Portable or Static."))
					}
					return
				},
			},

			// IP version 4 or IP version 6
			"ip_version": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					ipVersion := v.(int)
					if ipVersion != 4 && ipVersion != 6 {
						errs = append(errs, errors.New(
							"ip version should be either 4 or 6."))
					}
					return
				},
			},

			"capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			// vlan_id should be configured when type is "Portable"
			"vlan_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"endpoint_ip"},
			},

			// endpoint_ip should be configured when type is "Static"
			"endpoint_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vlan_id"},
			},

			// Provides IP address/netmask format (ex. 10.10.10.10/28)
			"subnet": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSoftLayerSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	// Find price items with AdditionalServicesSubnetAddresses
	productOrderContainer, err := buildSubnetProductOrderContainer(d, sess)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	log.Println("[INFO] Creating subnet")

	receipt, err := services.GetProductOrderService(sess).
		PlaceOrder(productOrderContainer, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error during creation of subnet: %s", err)
	}

	Subnet, err := findSubnetByOrderId(sess, *receipt.OrderId)
	if err != nil {
		return fmt.Errorf("Error during creation of subnet: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *Subnet.Id))

	return resourceSoftLayerSubnetUpdate(d, meta)
}

func resourceSoftLayerSubnetRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetService(sess)

	subnetId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid subnet ID, must be an integer: %s", err)
	}

	subnet, err := service.Id(subnetId).Mask(SubnetMask).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving a subnet: %s", err)
	}

	d.Set("network", *subnet.AddressSpace)
	d.Set("type", *subnet.SubnetType)
	d.Set("type", subnetTypeMap[*subnet.SubnetType])
	d.Set("ip_version", *subnet.Version)
	d.Set("capacity", *subnet.IpAddressCount)
	d.Set("subnet", *subnet.NetworkIdentifier+"/"+strconv.Itoa(*subnet.Cidr))
	if subnet.Note != nil {
		d.Set("notes", *subnet.Note)
	}
	if subnet.EndPointIpAddress != nil {
		d.Set("endpoint_ip", *subnet.EndPointIpAddress.IpAddress)
	}
	if subnet.NetworkVlan != nil {
		d.Set("vlan_id", *subnet.NetworkVlan.Id)
	}
	return nil
}

func resourceSoftLayerSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetService(sess)

	subnetId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid subnet ID, must be an integer: %s", err)
	}

	if d.HasChange("notes") {
		_, err = service.Id(subnetId).EditNote(sl.String(d.Get("notes").(string)))
		if err != nil {
			return fmt.Errorf("Error updating subnet: %s", err)
		}
	}
	return resourceSoftLayerSubnetRead(d, meta)
}

func resourceSoftLayerSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetService(sess)

	subnetId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid subnet ID, must be an integer: %s", err)
	}

	billingItem, err := service.Id(subnetId).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error deleting subnet: %s", err)
	}

	if billingItem.Id == nil {
		return nil
	}

	_, err = services.GetBillingItemService(sess).Id(*billingItem.Id).CancelService()

	return err
}

func resourceSoftLayerSubnetExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkSubnetService(sess)

	subnetId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(subnetId).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving subnet: %s", err)
	}
	return result.Id != nil && *result.Id == subnetId, nil
}

func findSubnetByOrderId(sess *session.Session, orderId int) (datatypes.Network_Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			subnets, err := services.GetAccountService(sess).
				Filter(filter.Path("subnets.billingItem.orderItem.order.id").
					Eq(strconv.Itoa(orderId)).Build()).
				Mask("id").
				GetSubnets()
			if err != nil {
				return datatypes.Network_Subnet{}, "", err
			}

			if len(subnets) == 1 {
				return subnets[0], "complete", nil
			}
			return nil, "pending", nil
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Subnet{}, err
	}

	if result, ok := pendingResult.(datatypes.Network_Subnet); ok {
		return result, nil
	}

	return datatypes.Network_Subnet{},
		fmt.Errorf("Cannot find a subnet with order id '%d'", orderId)
}

func buildSubnetProductOrderContainer(d *schema.ResourceData, sess *session.Session) (
	*datatypes.Container_Product_Order_Network_Subnet, error) {

	// 1. Get a package
	typeStr := d.Get("type").(string)
	vlanId := d.Get("vlan_id").(int)
	network := d.Get("network").(string)

	pkg, err := product.GetPackageByType(sess, subnetPackageTypeMap[typeStr])
	if err != nil {
		return &datatypes.Container_Product_Order_Network_Subnet{}, err
	}

	// 2. Get all prices for the package
	productItems, err := product.GetPackageProducts(sess, *pkg.Id)
	if err != nil {
		return &datatypes.Container_Product_Order_Network_Subnet{}, err
	}

	// 3. Select items which have a matching capacity, network, and IP version.
	capacity := d.Get("capacity").(int)
	ipVersionStr := "_IP_"
	if d.Get("ip_version").(int) == 6 {
		ipVersionStr = "_IPV6_"
	}
	SubnetItems := []datatypes.Product_Item{}
	for _, item := range productItems {
		if int(*item.Capacity) == d.Get("capacity").(int) &&
			strings.Contains(*item.KeyName, network) &&
			strings.Contains(*item.KeyName, ipVersionStr) {
			SubnetItems = append(SubnetItems, item)
		}
	}

	if len(SubnetItems) == 0 {
		return &datatypes.Container_Product_Order_Network_Subnet{},
			fmt.Errorf("No product items matching with capacity %d could be found", capacity)
	}

	productOrderContainer := datatypes.Container_Product_Order_Network_Subnet{
		Container_Product_Order: datatypes.Container_Product_Order{
			PackageId: pkg.Id,
			Prices: []datatypes.Product_Item_Price{
				{
					Id: SubnetItems[0].Prices[0].Id,
				},
			},
			Quantity: sl.Int(1),
		},
		EndPointVlanId: sl.Int(vlanId),
	}

	if endpointIp, ok := d.GetOk("endpoint_ip"); ok {
		if typeStr != "Static" {
			return &datatypes.Container_Product_Order_Network_Subnet{},
				fmt.Errorf("endpoint_ip is only available when type is Static.")
		}
		endpointIpStr := endpointIp.(string)
		subnet, err := services.GetNetworkSubnetService(sess).Mask("ipAddresses").GetSubnetForIpAddress(sl.String(endpointIpStr))
		if err != nil {
			return &datatypes.Container_Product_Order_Network_Subnet{}, err
		}
		for _, ipSubnet := range subnet.IpAddresses {
			if *ipSubnet.IpAddress == endpointIpStr {
				productOrderContainer.EndPointIpAddressId = ipSubnet.Id
			}
		}
		if productOrderContainer.EndPointIpAddressId == nil {
			return &datatypes.Container_Product_Order_Network_Subnet{},
				fmt.Errorf("Unable to find an ID of ipAddress: %s", endpointIpStr)
		}
	}
	return &productOrderContainer, nil
}

func getVlanType(sess *session.Session, vlanId int) (string, error) {
	vlan, err := services.GetNetworkVlanService(sess).Id(vlanId).Mask(VlanMask).GetObject()

	if err != nil {
		return "", fmt.Errorf("Error retrieving vlan: %s", err)
	}

	if vlan.PrimaryRouter != nil {
		if strings.HasPrefix(*vlan.PrimaryRouter.Hostname, "fcr") {
			return "PUBLIC", nil
		} else {
			return "PRIVATE", nil
		}
	}
	return "", fmt.Errorf("Unable to determine network.")
}
