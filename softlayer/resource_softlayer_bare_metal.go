package softlayer

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerBareMetal() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerBareMetalCreate,
		Read:     resourceSoftLayerBareMetalRead,
		Update:   resourceSoftLayerBareMetalUpdate,
		Delete:   resourceSoftLayerBareMetalDelete,
		Exists:   resourceSoftLayerBareMetalExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DefaultFunc: genId,
				DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
					// FIXME: Work around another bug in terraform.
					// When a default function is used with an optional property,
					// terraform will always execute it on apply, even when the property
					// already has a value in the state for it. This causes a false diff.
					// Making the property Computed:true does not make a difference.
					if strings.HasPrefix(o, "terraformed-") && strings.HasPrefix(n, "terraformed-") {
						return true
					}

					return o == n
				},
			},

			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"os_reference_code": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image_template_id"},
			},

			"hourly_billing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},

			"private_network_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"public_vlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"public_subnet": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"private_vlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"private_subnet": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"network_speed": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
				ForceNew: true,
			},

			"public_ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ssh_key_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				ForceNew: true,
			},

			"user_metadata": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"notes": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"post_install_script_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: true,
			},

			"fixed_config_preset": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"quote_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"image_template_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"os_reference_code"},
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"model": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"disk_controller": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"disks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"redundant_power_supply": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getBareMetalOrderFromResourceData(d *schema.ResourceData, meta interface{}) (datatypes.Hardware, error) {
	dc := datatypes.Location{
		Name: sl.String(d.Get("datacenter").(string)),
	}

	networkComponent := datatypes.Network_Component{
		MaxSpeed: sl.Int(d.Get("network_speed").(int)),
	}

	hardware := datatypes.Hardware{
		Hostname:               sl.String(d.Get("hostname").(string)),
		Domain:                 sl.String(d.Get("domain").(string)),
		HourlyBillingFlag:      sl.Bool(d.Get("hourly_billing").(bool)),
		PrivateNetworkOnlyFlag: sl.Bool(d.Get("private_network_only").(bool)),
		Datacenter:             &dc,
		NetworkComponents:      []datatypes.Network_Component{networkComponent},
		PostInstallScriptUri:   sl.String(d.Get("post_install_script_uri").(string)),
		BareMetalInstanceFlag:  sl.Int(1),
	}

	if fixed_config_preset, ok := d.GetOk("fixed_config_preset"); ok {
		hardware.FixedConfigurationPreset = &datatypes.Product_Package_Preset{
			KeyName: sl.String(fixed_config_preset.(string)),
		}
	}

	if operatingSystemReferenceCode, ok := d.GetOk("os_reference_code"); ok {
		hardware.OperatingSystemReferenceCode = sl.String(operatingSystemReferenceCode.(string))
	}

	public_vlan_id := d.Get("public_vlan_id").(int)
	if public_vlan_id > 0 {
		hardware.PrimaryNetworkComponent = &datatypes.Network_Component{
			NetworkVlan: &datatypes.Network_Vlan{Id: sl.Int(public_vlan_id)},
		}
	}

	private_vlan_id := d.Get("private_vlan_id").(int)
	if private_vlan_id > 0 {
		hardware.PrimaryBackendNetworkComponent = &datatypes.Network_Component{
			NetworkVlan: &datatypes.Network_Vlan{Id: sl.Int(private_vlan_id)},
		}
	}

	if public_subnet, ok := d.GetOk("public_subnet"); ok {
		subnet := public_subnet.(string)
		subnetId, err := getSubnetId(subnet, meta)
		if err != nil {
			return hardware, fmt.Errorf("Error determining id for subnet %s: %s", subnet, err)
		}

		hardware.PrimaryNetworkComponent.NetworkVlan.PrimarySubnetId = sl.Int(subnetId)
	}

	if private_subnet, ok := d.GetOk("private_subnet"); ok {
		subnet := private_subnet.(string)
		subnetId, err := getSubnetId(subnet, meta)
		if err != nil {
			return hardware, fmt.Errorf("Error determining id for subnet %s: %s", subnet, err)
		}

		hardware.PrimaryBackendNetworkComponent.NetworkVlan.PrimarySubnetId = sl.Int(subnetId)
	}

	if userMetadata, ok := d.GetOk("user_metadata"); ok {
		hardware.UserData = []datatypes.Hardware_Attribute{
			{Value: sl.String(userMetadata.(string))},
		}
	}

	// Get configured ssh_keys
	ssh_key_ids := d.Get("ssh_key_ids").([]interface{})
	if len(ssh_key_ids) > 0 {
		hardware.SshKeys = make([]datatypes.Security_Ssh_Key, 0, len(ssh_key_ids))
		for _, ssh_key_id := range ssh_key_ids {
			hardware.SshKeys = append(hardware.SshKeys, datatypes.Security_Ssh_Key{
				Id: sl.Int(ssh_key_id.(int)),
			})
		}
	}

	return hardware, nil
}

func resourceSoftLayerBareMetalCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	var order datatypes.Container_Product_Order

	hardware, err := getBareMetalOrderFromResourceData(d, meta)
	if err != nil {
		return err
	}

	quote_id := d.Get("quote_id").(int)
	if quote_id > 0 {
		// Build a bare metal template from the quote.
		order, err = services.GetBillingOrderQuoteService(sess).
			Id(quote_id).GetRecalculatedOrderContainer(nil, sl.Bool(false))
		if err != nil {
			return fmt.Errorf(
				"Encountered problem trying to get the bare metal order template from quote: %s", err)
		}
		order.Quantity = sl.Int(1)
		order.Hardware = make([]datatypes.Hardware, 0, 1)
		order.Hardware = append(
			order.Hardware,
			hardware,
		)
	} else if _, ok := d.GetOk("fixed_config_preset"); ok {
		// Build a pre-configured bare metal server
		order, err = services.GetHardwareService(sess).GenerateOrderTemplate(&hardware)
		if err != nil {
			return fmt.Errorf(
				"Encountered problem trying to get the bare metal order template: %s", err)
		}
	} else {
		// Build a custom bare metal server
		dc, err := location.GetDatacenterByName(sess, d.Get("datacenter").(string), "id")
		if err != nil {
			return err
		}

		model, ok := d.GetOk("model")
		if !ok {
			return fmt.Errorf("The attribute 'model' is not defined.")
		}

		// 1. Get a package by keyName
		pkg, err := product.GetPackageByKeyName(sess, model.(string))
		if err != nil {
			return err
		}

		// 2. Get all prices for the package
		items, err := product.GetPackageProducts(sess, *pkg.Id, "id,categories,capacity,description,units,keyName,prices[id,categories[id,name,categoryCode]]")
		if err != nil {
			return err
		}

		log.Printf("**************** Length of items %d", len(items))

		// 3. Build price items
		disks := d.Get("disks").([]interface{})
		server, err := getItemPriceId(items, "server", d.Get("cpu").(string))
		if err != nil {
			return err
		}
		os, err := getItemPriceId(items, "os", "OS_UBUNTU_14_04_LTS_TRUSTY_TAHR_64_BIT")
		if err != nil {
			return err
		}
		ram, err := getItemPriceId(items, "ram", d.Get("memory").(string))
		if err != nil {
			return err
		}
		diskController, err := getItemPriceId(items, "disk_controller", d.Get("disk_controller").(string))
		if err != nil {
			return err
		}
		disk0, err := getItemPriceId(items, "disk0", disks[0].(string))
		if err != nil {
			return err
		}
		portSpeed, err := getItemPriceId(items, "port_speed", "1_GBPS_PUBLIC_PRIVATE_NETWORK_UPLINKS")
		if err != nil {
			return err
		}
		/*
			powerSupply, err := getItemPriceId(items, "power_supply", "REDUNDANT_POWER_SUPPLY")
			if err != nil {
				return err
			}
		*/
		bandwidth, err := getItemPriceId(items, "bandwidth", "BANDWIDTH_20000_GB")
		if err != nil {
			return err
		}
		priIpAddress, err := getItemPriceId(items, "pri_ip_addresses", "1_IP_ADDRESS")
		if err != nil {
			return err
		}
		remoteManagement, err := getItemPriceId(items, "remote_management", "REBOOT_KVM_OVER_IP")
		if err != nil {
			return err
		}
		vpnManagement, err := getItemPriceId(items, "vpn_management", "UNLIMITED_SSL_VPN_USERS_1_PPTP_VPN_USER_PER_ACCOUNT")
		if err != nil {
			return err
		}
		monitoring, err := getItemPriceId(items, "monitoring", "MONITORING_HOST_PING")
		if err != nil {
			return err
		}
		notification, err := getItemPriceId(items, "notification", "NOTIFICATION_EMAIL_AND_TICKET")
		if err != nil {
			return err
		}
		response, err := getItemPriceId(items, "response", "AUTOMATED_NOTIFICATION")
		if err != nil {
			return err
		}
		vulnerabilityScanner, err := getItemPriceId(items, "vulnerability_scanner", "NESSUS_VULNERABILITY_ASSESSMENT_REPORTING")
		if err != nil {
			return err
		}
		order = datatypes.Container_Product_Order{
			Quantity: sl.Int(1),
			Hardware: []datatypes.Hardware{{
				Hostname: sl.String(d.Get("hostname").(string)),
				Domain:   sl.String(d.Get("domain").(string)),
			}},
			Location:  sl.String(strconv.Itoa(*dc.Id)),
			PackageId: pkg.Id,
			Prices: []datatypes.Product_Item_Price{
				server,
				os,
				ram,
				diskController,
				disk0,
				portSpeed,
				//	powerSupply,
				bandwidth,
				priIpAddress,
				remoteManagement,
				vpnManagement,
				monitoring,
				notification,
				response,
				vulnerabilityScanner,
			},
		}
	}

	// Set image template id if it exists
	if rawImageTemplateId, ok := d.GetOk("image_template_id"); ok {
		imageTemplateId := rawImageTemplateId.(int)
		order.ImageTemplateId = sl.Int(imageTemplateId)
	}

	log.Println("[INFO] Ordering bare metal server")
	_, err = services.GetProductOrderService(sess).PlaceOrder(&order, sl.Bool(true))
	if err != nil {
		return fmt.Errorf("Error ordering bare metal server: %s\n%+v\n", err, order)
	}

	log.Printf("[INFO] Bare Metal Server ID: %s", d.Id())

	// wait for machine availability
	bm, err := waitForBareMetalProvision(&hardware, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for bare metal server (%s) to become ready: %s", d.Id(), err)
	}

	id := *bm.(datatypes.Hardware).Id
	d.SetId(fmt.Sprintf("%d", id))

	// Set tags
	err = setHardwareTags(id, d, meta)
	if err != nil {
		return err
	}

	// Set notes
	if d.Get("notes").(string) != "" {
		err = setHardwareNotes(id, d, meta)
		if err != nil {
			return err
		}
	}

	return resourceSoftLayerBareMetalRead(d, meta)
}

func resourceSoftLayerBareMetalRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetHardwareService(meta.(ProviderConfig).SoftLayerSession())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).Mask(
		"hostname,domain," +
			"primaryIpAddress,primaryBackendIpAddress,privateNetworkOnlyFlag," +
			"notes,userData[value],tagReferences[id,tag[name]]," +
			"hourlyBillingFlag," +
			"datacenter[id,name,longName]," +
			"primaryNetworkComponent[networkVlan[id,primaryRouter,vlanNumber],maxSpeed]," +
			"primaryBackendNetworkComponent[networkVlan[id,primaryRouter,vlanNumber],maxSpeed]",
	).GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving bare metal server: %s", err)
	}

	d.Set("hostname", *result.Hostname)
	d.Set("domain", *result.Domain)

	if result.Datacenter != nil {
		d.Set("datacenter", *result.Datacenter.Name)
	}

	d.Set("network_speed", *result.PrimaryNetworkComponent.MaxSpeed)
	if result.PrimaryIpAddress != nil {
		d.Set("public_ipv4_address", *result.PrimaryIpAddress)
	}
	d.Set("private_ipv4_address", *result.PrimaryBackendIpAddress)

	d.Set("private_network_only", *result.PrivateNetworkOnlyFlag)
	d.Set("hourly_billing", *result.HourlyBillingFlag)

	if result.PrimaryNetworkComponent.NetworkVlan != nil {
		d.Set("public_vlan_id", *result.PrimaryNetworkComponent.NetworkVlan.Id)
	}

	if result.PrimaryBackendNetworkComponent.NetworkVlan != nil {
		d.Set("private_vlan_id", *result.PrimaryBackendNetworkComponent.NetworkVlan.Id)
	}

	userData := result.UserData
	if len(userData) > 0 && userData[0].Value != nil {
		d.Set("user_metadata", *userData[0].Value)
	}

	d.Set("notes", sl.Get(result.Notes, nil))

	tagReferences := result.TagReferences
	tagReferencesLen := len(tagReferences)
	if tagReferencesLen > 0 {
		tags := make([]string, 0, tagReferencesLen)
		for _, tagRef := range tagReferences {
			tags = append(tags, *tagRef.Tag.Name)
		}
		d.Set("tags", tags)
	}

	connInfo := map[string]string{"type": "ssh"}
	if !*result.PrivateNetworkOnlyFlag && result.PrimaryIpAddress != nil {
		connInfo["host"] = *result.PrimaryIpAddress
	} else {
		connInfo["host"] = *result.PrimaryBackendIpAddress
	}
	d.SetConnInfo(connInfo)

	return nil
}

func resourceSoftLayerBareMetalUpdate(d *schema.ResourceData, meta interface{}) error {
	id, _ := strconv.Atoi(d.Id())

	if d.HasChange("tags") {
		err := setHardwareTags(id, d, meta)
		if err != nil {
			return err
		}
	}

	if d.HasChange("notes") {
		err := setHardwareNotes(id, d, meta)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceSoftLayerBareMetalDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetHardwareService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = waitForNoBareMetalActiveTransactions(id, meta)
	if err != nil {
		return fmt.Errorf("Error deleting bare metal server while waiting for zero active transactions: %s", err)
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
		return fmt.Errorf("Error canceling the bare metal server (%d): %s", id, err)
	}

	return nil
}

func resourceSoftLayerBareMetalExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	service := services.GetHardwareService(meta.(ProviderConfig).SoftLayerSession())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); !ok || apiErr.StatusCode != 404 {
			return false, fmt.Errorf("Error trying to retrieve the Bare Metal server: %s", err)
		}
	}

	return err == nil && result.Id != nil && *result.Id == id, nil
}

// Bare metal creation does not return a bare metal object with an Id.
// Have to wait on provision date to become available on server that matches
// hostname and domain.
// http://sldn.softlayer.com/blog/bpotter/ordering-bare-metal-servers-using-softlayer-api
func waitForBareMetalProvision(d *datatypes.Hardware, meta interface{}) (interface{}, error) {
	hostname := *d.Hostname
	domain := *d.Domain
	log.Printf("Waiting for server (%s.%s) to have to be provisioned", hostname, domain)

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", "pending"},
		Target:  []string{"provisioned"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetAccountService(meta.(ProviderConfig).SoftLayerSession())
			bms, err := service.Filter(
				filter.Build(
					filter.Path("hardware.hostname").Eq(hostname),
					filter.Path("hardware.domain").Eq(domain),
				),
			).Mask("id,provisionDate").GetHardware()
			if err != nil {
				return false, "retry", nil
			}

			if len(bms) == 0 || bms[0].ProvisionDate == nil {
				return datatypes.Hardware{}, "pending", nil
			} else {
				return bms[0], "provisioned", nil
			}
		},
		Timeout:    4 * time.Hour,
		Delay:      30 * time.Second,
		MinTimeout: 2 * time.Minute,
	}

	return stateConf.WaitForState()
}

func waitForNoBareMetalActiveTransactions(id int, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for server (%d) to have zero active transactions", id)
	service := services.GetHardwareServerService(meta.(ProviderConfig).SoftLayerSession())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", "active"},
		Target:  []string{"idle"},
		Refresh: func() (interface{}, string, error) {
			bm, err := service.Id(id).Mask("id,activeTransactionCount").GetObject()
			if err != nil {
				return false, "retry", nil
			}

			if bm.ActiveTransactionCount != nil && *bm.ActiveTransactionCount == 0 {
				return bm, "idle", nil
			} else {
				return bm, "active", nil
			}
		},
		Timeout:    4 * time.Hour,
		Delay:      5 * time.Second,
		MinTimeout: 1 * time.Minute,
	}

	return stateConf.WaitForState()
}

func setHardwareTags(id int, d *schema.ResourceData, meta interface{}) error {
	service := services.GetHardwareService(meta.(ProviderConfig).SoftLayerSession())

	tags := getTags(d)
	_, err := service.Id(id).SetTags(sl.String(tags))
	if err != nil {
		return fmt.Errorf("Could not set tags on bare metal server %d", id)
	}

	return nil
}

func setHardwareNotes(id int, d *schema.ResourceData, meta interface{}) error {
	service := services.GetHardwareServerService(meta.(ProviderConfig).SoftLayerSession())

	result, err := service.Id(id).GetObject()
	if err != nil {
		return err
	}

	result.Notes = sl.String(d.Get("notes").(string))

	_, err = service.Id(id).EditObject(&result)
	if err != nil {
		return err
	}

	return nil
}

// Example : getItemPriceId(items, 'server', 'INTEL_XEON_2690_2_60')
func getItemPriceId(items []datatypes.Product_Item, categoryCode string, keyName string) (datatypes.Product_Item_Price, error) {
	for _, item := range items {
		for _, itemCategory := range item.Categories {
			if *itemCategory.CategoryCode == categoryCode && *item.KeyName == keyName {
				for _, price := range item.Prices {
					if price.LocationGroupId == nil {
						return datatypes.Product_Item_Price{Id: price.Id}, nil
					}
				}
			}
		}
	}
	return datatypes.Product_Item_Price{},
		fmt.Errorf("Could not find the matching item with categorycode %s and keyName %s", categoryCode, keyName)
}
