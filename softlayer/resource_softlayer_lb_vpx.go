package softlayer

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"errors"
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
	PACKAGE_ID_APPLICATION_DELIVERY_CONTROLLER = 192
	DELIMITER                                  = "_"
)

func resourceSoftLayerLbVpx() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbVpxCreate,
		Read:     resourceSoftLayerLbVpxRead,
		Delete:   resourceSoftLayerLbVpxDelete,
		Exists:   resourceSoftLayerLbVpxExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"speed": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ip_count": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"front_end_vlan": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlan_number": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"primary_router_hostname": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"front_end_subnet": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"back_end_vlan": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlan_number": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"primary_router_hostname": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"back_end_subnet": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"vip_pool": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func getVlanId(vlanNumber int, primaryRouterHostname string, meta interface{}) (int, error) {
	service := services.GetAccountService(meta.(*session.Session))

	networkVlan, err := service.
		Mask("id").
		Filter(
			filter.Build(
				filter.Path("networkVlans.primaryRouter.hostname").Eq(primaryRouterHostname),
				filter.Path("networkVlans.vlanNumber").Eq(vlanNumber),
			),
		).
		GetNetworkVlans()

	if err != nil {
		return 0, fmt.Errorf("Error looking up Vlan: %s", err)
	}

	if len(networkVlan) < 1 {
		return 0, fmt.Errorf(
			"Unable to locate a vlan matching the provided router hostname and vlan number: %s/%d",
			primaryRouterHostname,
			vlanNumber)
	}

	return *networkVlan[0].Id, nil
}

func getSubnetId(subnet string, meta interface{}) (int, error) {
	service := services.GetAccountService(meta.(*session.Session))

	subnetInfo := strings.Split(subnet, "/")
	if len(subnetInfo) != 2 {
		return 0, fmt.Errorf(
			"Unable to parse the provided subnet: %s", subnet)
	}

	networkIdentifier := subnetInfo[0]
	cidr := subnetInfo[1]

	subnets, err := service.
		Mask("id").
		Filter(
			filter.Build(
				filter.Path("subnets.cidr").Eq(cidr),
				filter.Path("subnets.networkIdentifier").Eq(networkIdentifier),
			),
		).
		GetSubnets()

	if err != nil {
		return 0, fmt.Errorf("Error looking up Subnet: %s", err)
	}

	if len(subnets) < 1 {
		return 0, fmt.Errorf(
			"Unable to locate a subnet matching the provided subnet: %s", subnet)
	}

	return *subnets[0].Id, nil
}

func getVPXPriceItemKeyName(version string, speed int, plan string) string {
	name := "CITRIX_NETSCALER_VPX"
	speedMeasurements := "MBPS"
	versionReplaced := strings.Replace(version, ".", DELIMITER, -1)
	speedString := strconv.Itoa(speed) + speedMeasurements

	return strings.Join([]string{name, versionReplaced, speedString, strings.ToUpper(plan)}, DELIMITER)
}

func getPublicIpItemKeyName(ipCount int) string {
	name := "STATIC_PUBLIC_IP_ADDRESSES"
	ipCountString := strconv.Itoa(ipCount)

	return strings.Join([]string{ipCountString, name}, DELIMITER)
}

func findVPXPriceItems(version string, speed int, plan string, ipCount int, meta interface{}) ([]datatypes.Product_Item_Price, error) {
	sess := meta.(*session.Session)

	// Get VPX package type.
	productPackage, err := product.GetPackageByType(sess, "ADDITIONAL_SERVICES_APPLICATION_DELIVERY_APPLIANCE")
	if err != nil {
		return []datatypes.Product_Item_Price{}, err
	}

	// Get VPX product items
	items, err := product.GetPackageProducts(sess, *productPackage.Id)
	if err != nil {
		return []datatypes.Product_Item_Price{}, err
	}

	// Get VPX and static IP items
	nadcKey := getVPXPriceItemKeyName(version, speed, plan)
	ipKey := getPublicIpItemKeyName(ipCount)

	var nadcItemPrice, ipItemPrice datatypes.Product_Item_Price

	for _, item := range items {
		itemKey := item.KeyName
		if *itemKey == nadcKey {
			nadcItemPrice = item.Prices[0]
		}
		if *itemKey == ipKey {
			ipItemPrice = item.Prices[0]
		}
	}

	var errorMessages []string

	if nadcItemPrice.Id == nil {
		errorMessages = append(errorMessages, fmt.Sprintf("VPX version, speed or plan have incorrect values"))
	}

	if ipItemPrice.Id == nil {
		errorMessages = append(errorMessages, fmt.Sprintf("IP quantity value is incorrect"))
	}

	if len(errorMessages) > 0 {
		err = errors.New(strings.Join(errorMessages, "\n"))
		return []datatypes.Product_Item_Price{}, err
	}

	return []datatypes.Product_Item_Price{
		{
			Id: nadcItemPrice.Id,
		},
		{
			Id: ipItemPrice.Id,
		},
	}, nil
}

func findVPXByOrderId(orderId int, meta interface{}) (datatypes.Network_Application_Delivery_Controller, error) {
	service := services.GetAccountService(meta.(*session.Session))

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			vpxs, err := service.
				Filter(
					filter.Build(
						filter.Path("applicationDeliveryControllers.billingItem.orderItem.order.id").Eq(orderId),
					),
				).GetApplicationDeliveryControllers()
			if err != nil {
				return datatypes.Network_Application_Delivery_Controller{}, "", err
			}

			if len(vpxs) == 1 {
				return vpxs[0], "complete", nil
			} else if len(vpxs) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one VPX: %s", err)
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.Network_Application_Delivery_Controller{}, err
	}

	var result, ok = pendingResult.(datatypes.Network_Application_Delivery_Controller)

	if ok {
		return result, nil
	}

	return datatypes.Network_Application_Delivery_Controller{},
		fmt.Errorf("Cannot find Application Delivery Controller with order id '%d'", orderId)
}

func resourceSoftLayerLbVpxCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	productOrderService := services.GetProductOrderService(sess)
	NADCService := services.GetNetworkApplicationDeliveryControllerService(sess)
	var err error

	opts := datatypes.Container_Product_Order{
		PackageId: sl.Int(PACKAGE_ID_APPLICATION_DELIVERY_CONTROLLER),
		Quantity:  sl.Int(1),
	}

	opts.Prices, err = findVPXPriceItems(
		d.Get("version").(string),
		d.Get("speed").(int),
		d.Get("plan").(string),
		d.Get("ip_count").(int),
		meta)

	if err != nil {
		return fmt.Errorf("Error Cannot find Application Delivery Controller prices '%s'.", err)
	}

	if len(d.Get("datacenter").(string)) > 0 {
		datacenter, err := location.GetDatacenterByName(sess, "dal06", "id")
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		opts.Location = sl.String(strconv.Itoa(*datacenter.Id))
	}

	opts.Hardware = make([]datatypes.Hardware, 1)

	if len(d.Get("front_end_vlan.vlan_number").(string)) > 0 || len(d.Get("front_end_subnet").(string)) > 0 {
		opts.Hardware[0].PrimaryNetworkComponent = &datatypes.Network_Component{}
	}

	if len(d.Get("front_end_vlan.vlan_number").(string)) > 0 {
		vlanNumber, err := strconv.Atoi(d.Get("front_end_vlan.vlan_number").(string))
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		networkVlanId, err := getVlanId(vlanNumber, d.Get("front_end_vlan.primary_router_hostname").(string), meta)
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		opts.Hardware[0].PrimaryNetworkComponent.NetworkVlanId = sl.Int(networkVlanId)
	}

	if len(d.Get("front_end_subnet").(string)) > 0 {
		primarySubnetId, err := getSubnetId(d.Get("front_end_subnet").(string), meta)
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		opts.Hardware[0].PrimaryNetworkComponent.NetworkVlan = &datatypes.Network_Vlan{
			PrimarySubnetId: sl.Int(primarySubnetId),
		}
	}

	if len(d.Get("back_end_vlan.vlan_number").(string)) > 0 || len(d.Get("back_end_subnet").(string)) > 0 {
		opts.Hardware[0].PrimaryBackendNetworkComponent = &datatypes.Network_Component{}
	}

	if len(d.Get("back_end_vlan.vlan_number").(string)) > 0 {
		vlanNumber, err := strconv.Atoi(d.Get("back_end_vlan.vlan_number").(string))
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		networkVlanId, err := getVlanId(vlanNumber, d.Get("back_end_vlan.primary_router_hostname").(string), meta)
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		opts.Hardware[0].PrimaryBackendNetworkComponent.NetworkVlanId = sl.Int(networkVlanId)
	}

	if len(d.Get("back_end_subnet").(string)) > 0 {
		primarySubnetId, err := getSubnetId(d.Get("back_end_subnet").(string), meta)
		if err != nil {
			return fmt.Errorf("Error creating network application delivery controller: %s", err)
		}
		opts.Hardware[0].PrimaryBackendNetworkComponent.NetworkVlan = &datatypes.Network_Vlan{
			PrimarySubnetId: sl.Int(primarySubnetId),
		}
	}

	log.Printf("[INFO] Creating network application delivery controller")

	receipt, err := productOrderService.PlaceOrder(&opts, sl.Bool(false))

	if err != nil {
		return fmt.Errorf("Error creating network application delivery controller: %s", err)
	}

	VPX, err := findVPXByOrderId(*receipt.OrderId, meta)

	if err != nil {
		return fmt.Errorf("Error creating network application delivery controller: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *VPX.Id))

	log.Printf("[INFO] Netscaler VPX ID: %s", d.Id())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	// Wait Virtual IP provisioning
	IsVipReady := false

	for vipWaitCount := 0; vipWaitCount < 60; vipWaitCount++ {
		getObjectResult, err := NADCService.Id(id).Mask("subnets[ipAddresses]").GetObject()
		if err != nil {
			return fmt.Errorf("Error retrieving network application delivery controller: %s", err)
		}

		ipCount := 0
		if getObjectResult.Subnets != nil && len(getObjectResult.Subnets) > 0 && getObjectResult.Subnets[0].IpAddresses != nil {
			ipCount = len(getObjectResult.Subnets[0].IpAddresses)
		}
		if ipCount > 0 {
			IsVipReady = true
			break
		}
		log.Printf("[INFO] Wait 10 seconds for Virtual IP provisioning on Netscaler VPX ID: %d", id)
		time.Sleep(time.Second * 10)
	}

	if !IsVipReady {
		return fmt.Errorf("Failed to create VIPs for Netscaler VPX ID: %d", id)
	}
	return resourceSoftLayerLbVpxRead(d, meta)
}

func resourceSoftLayerLbVpxRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	service := services.GetNetworkApplicationDeliveryControllerService(sess)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	getObjectResult, err := service.
		Id(id).
		Mask("id,name,type[name],datacenter,networkVlans[primaryRouter],networkVlans[primarySubnets],subnets[ipAddresses],description").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving network application delivery controller: %s", err)
	}

	d.Set("name", *getObjectResult.Name)
	d.Set("type", *getObjectResult.Type.Name)
	if getObjectResult.Datacenter != nil {
		d.Set("datacenter", *getObjectResult.Datacenter.Name)
	}

	frontEndVlan := d.Get("front_end_vlan").(map[string]interface{})
	backEndVlan := d.Get("back_end_vlan").(map[string]interface{})
	frontEndSubnet := ""
	backEndSubnet := ""

	for _, vlan := range getObjectResult.NetworkVlans {
		if vlan.PrimaryRouter != nil && *vlan.PrimaryRouter.Hostname != "" && *vlan.VlanNumber > 0 {
			isFcr := strings.HasPrefix(*vlan.PrimaryRouter.Hostname, "fcr")
			isBcr := strings.HasPrefix(*vlan.PrimaryRouter.Hostname, "bcr")
			if isFcr {
				frontEndVlan["primary_router_hostname"] = *vlan.PrimaryRouter.Hostname
				vlanNumber := strconv.Itoa(*vlan.VlanNumber)
				frontEndVlan["vlan_number"] = vlanNumber
				if vlan.PrimarySubnets != nil && len(vlan.PrimarySubnets) > 0 {
					ipAddress := *vlan.PrimarySubnets[0].NetworkIdentifier
					cidr := strconv.Itoa(*vlan.PrimarySubnets[0].Cidr)
					frontEndSubnet = ipAddress + "/" + cidr
				}
			}

			if isBcr {
				backEndVlan["primary_router_hostname"] = *vlan.PrimaryRouter.Hostname
				vlanNumber := strconv.Itoa(*vlan.VlanNumber)
				backEndVlan["vlan_number"] = vlanNumber
				if vlan.PrimarySubnets != nil && len(vlan.PrimarySubnets) > 0 {
					ipAddress := *vlan.PrimarySubnets[0].NetworkIdentifier
					cidr := strconv.Itoa(*vlan.PrimarySubnets[0].Cidr)
					backEndSubnet = ipAddress + "/" + cidr
				}
			}
		}
	}

	d.Set("front_end_vlan", frontEndVlan)
	d.Set("back_end_vlan", backEndVlan)
	d.Set("front_end_subnet", frontEndSubnet)
	d.Set("back_end_subnet", backEndSubnet)

	vips := make([]string, 0)
	ipCount := 0
	for i, subnet := range getObjectResult.Subnets {
		for _, ipAddressObj := range subnet.IpAddresses {
			vips = append(vips, *ipAddressObj.IpAddress)
			if i == 0 {
				ipCount++
			}
		}
	}

	d.Set("vip_pool", vips)
	d.Set("ip_count", ipCount)

	description := *getObjectResult.Description
	r, _ := regexp.Compile(" [0-9]+Mbps")
	speedStr := r.FindString(description)
	r, _ = regexp.Compile("[0-9]+")
	speed, err := strconv.Atoi(r.FindString(speedStr))
	if err == nil && speed > 0 {
		d.Set("speed", speed)
	}

	r, _ = regexp.Compile(" VPX [0-9]+\\.[0-9]+ ")
	versionStr := r.FindString(description)
	r, _ = regexp.Compile("[0-9]+\\.[0-9]+")
	version := r.FindString(versionStr)
	if version != "" {
		d.Set("version", version)
	}

	r, _ = regexp.Compile(" [A-Za-z]+$")
	planStr := r.FindString(description)
	r, _ = regexp.Compile("[A-Za-z]+$")
	plan := r.FindString(planStr)
	if plan != "" {
		d.Set("plan", plan)
	}

	return nil
}

func resourceSoftLayerLbVpxDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	billingItem, err := service.Id(id).GetBillingItem()
	if err != nil {
		return fmt.Errorf("Error deleting network application delivery controller: %s", err)
	}

	if *billingItem.Id > 0 {
		billingItemService := services.GetBillingItemService(sess)
		deleted, err := billingItemService.Id(*billingItem.Id).CancelService()
		if err != nil {
			return fmt.Errorf("Error deleting network application delivery controller: %s", err)
		}

		if deleted {
			return nil
		}
	}

	return nil
}

func resourceSoftLayerLbVpxExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	service := services.GetNetworkApplicationDeliveryControllerService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	nadc, err := service.Mask("id").Id(id).GetObject()

	return nadc.Id != nil && *nadc.Id == id && err == nil, nil
}
