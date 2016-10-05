package softlayer

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerBareMetal() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerBareMetalCreate,
		Read:     resourceSoftLayerBareMetalRead,
		Delete:   resourceSoftLayerBareMetalDelete,
		Exists:   resourceSoftLayerBareMetalExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Required: true,
				ForceNew: true,
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

			"post_install_script_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: true,
			},

			"fixed_config_preset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_template_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"os_reference_code"},
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

		FixedConfigurationPreset: &datatypes.Product_Package_Preset{
			KeyName: sl.String(d.Get("fixed_config_preset").(string)),
		},
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
	sess := meta.(*session.Session)
	hwService := services.GetHardwareService(sess)
	orderService := services.GetProductOrderService(sess)

	hardware, err := getBareMetalOrderFromResourceData(d, meta)
	if err != nil {
		return err
	}

	order, err := hwService.GenerateOrderTemplate(&hardware)
	if err != nil {
		return fmt.Errorf(
			"Encountered problem trying to get the bare metal order template: %s", err)
	}

	// Set image template id if it exists
	if rawImageTemplateId, ok := d.GetOk("image_template_id"); ok {
		imageTemplateId := rawImageTemplateId.(int)
		order.ImageTemplateId = sl.Int(imageTemplateId)
	}

	log.Println("[INFO] Ordering bare metal server")

	_, err = orderService.PlaceOrder(&order, sl.Bool(false))
	if err != nil {
		return fmt.Errorf("Error ordering bare metal server: %s", err)
	}

	log.Printf("[INFO] Bare Metal Server ID: %s", d.Id())

	// wait for machine availability
	bm, err := waitForBareMetalProvision(&hardware, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for bare metal server (%s) to become ready: %s", d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%d", *bm.(datatypes.Hardware).Id))

	return resourceSoftLayerBareMetalRead(d, meta)
}

func resourceSoftLayerBareMetalRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetHardwareService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).Mask(
		"hostname,domain," +
			"primaryIpAddress,primaryBackendIpAddress,privateNetworkOnlyFlag," +
			"userData[value]," +
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
		d.Set("user_data", *userData[0].Value)
	}

	return nil
}

func resourceSoftLayerBareMetalDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
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
	service := services.GetHardwareService(meta.(*session.Session))

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
		Pending: []string{"", "pending"},
		Target:  []string{"provisioned"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetAccountService(meta.(*session.Session))
			bms, err := service.Filter(
				filter.Build(
					filter.Path("hardware.hostname").Eq(hostname),
					filter.Path("hardware.domain").Eq(domain),
				),
			).Mask("id,provisionDate").GetHardware()
			if err != nil {
				return nil, "", fmt.Errorf("Problem fetching bare metal servers matching %s.%s: %s", hostname, domain, err)
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
	service := services.GetHardwareServerService(meta.(*session.Session))

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "active"},
		Target:  []string{"idle"},
		Refresh: func() (interface{}, string, error) {
			bm, err := service.Id(id).Mask("id,activeTransactionCount").GetObject()
			if err != nil {
				return nil, "", fmt.Errorf("Couldn't get active transactions: %s", err)
			}

			if bm.ActiveTransactionCount != nil && *bm.ActiveTransactionCount == 0 {
				return bm, "idle", nil
			} else {
				return bm, "active", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}
