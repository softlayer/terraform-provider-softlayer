package softlayer

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/helpers/product"
	"github.com/softlayer/softlayer-go/helpers/virtual"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerVirtualGuest() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerVirtualGuestCreate,
		Read:     resourceSoftLayerVirtualGuestRead,
		Update:   resourceSoftLayerVirtualGuestUpdate,
		Delete:   resourceSoftLayerVirtualGuestDelete,
		Exists:   resourceSoftLayerVirtualGuestExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},

			"os_reference_code": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image_id"},
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

			"cpu": {
				Type:     schema.TypeInt,
				Required: true,
				// TODO: This fields for now requires recreation, because currently for some reason SoftLayer resets "dedicated_acct_host_only"
				// TODO: flag to false, while upgrading CPUs. That problem is reported to SoftLayer team. "ForceNew" can be set back
				// TODO: to false as soon as it is fixed at their side. Also corresponding test for virtual guest upgrade will be uncommented.
				ForceNew: true,
			},

			"ram": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					memoryInMB := float64(v.(int))

					// Validate memory to match gigs format
					remaining := math.Mod(memoryInMB, 1024)
					if remaining > 0 {
						suggested := math.Ceil(memoryInMB/1024) * 1024
						errors = append(errors, fmt.Errorf(
							"Invalid 'ram' value %d megabytes, must be a multiple of 1024 (e.g. use %d)", int(memoryInMB), int(suggested)))
					}

					return
				},
			},

			"dedicated_acct_host_only": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"front_end_vlan": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlan_number": {
							Type:     schema.TypeString,
							Required: true,
						},

						"primary_router_hostname": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"front_end_subnet": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"back_end_vlan": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlan_number": {
							Type:     schema.TypeString,
							Required: true,
						},

						"primary_router_hostname": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"back_end_subnet": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"disks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"network_speed": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
			},

			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv4_address_private": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ip_address_id_private": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ssh_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"local_disk": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"post_install_script_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: true,
			},

			"image_id": {
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
		},
	}
}

func getNameForBlockDevice(i int) string {
	// skip 1, which is reserved for the swap disk.
	// so we get 0, 2, 3, 4, 5 ...
	if i == 0 {
		return "0"
	} else {
		return strconv.Itoa(i + 1)
	}
}

func getBlockDevices(d *schema.ResourceData) []datatypes.Virtual_Guest_Block_Device {
	numBlocks := d.Get("disks.#").(int)
	if numBlocks == 0 {
		return nil
	} else {
		blocks := make([]datatypes.Virtual_Guest_Block_Device, 0, numBlocks)
		for i := 0; i < numBlocks; i++ {
			blockRef := fmt.Sprintf("disks.%d", i)
			name := getNameForBlockDevice(i)
			capacity := d.Get(blockRef).(int)
			block := datatypes.Virtual_Guest_Block_Device{
				Device: &name,
				DiskImage: &datatypes.Virtual_Disk_Image{
					Capacity: &capacity,
				},
			}
			blocks = append(blocks, block)
		}
		return blocks
	}
}
func getVirtualGuestTemplateFromResourceData(d *schema.ResourceData, meta interface{}) (datatypes.Virtual_Guest, error) {

	dc := datatypes.Location{
		Name: sl.String(d.Get("datacenter").(string)),
	}

	networkComponent := datatypes.Virtual_Guest_Network_Component{
		MaxSpeed: sl.Int(d.Get("network_speed").(int)),
	}

	opts := datatypes.Virtual_Guest{
		Hostname:               sl.String(d.Get("name").(string)),
		Domain:                 sl.String(d.Get("domain").(string)),
		HourlyBillingFlag:      sl.Bool(d.Get("hourly_billing").(bool)),
		PrivateNetworkOnlyFlag: sl.Bool(d.Get("private_network_only").(bool)),
		Datacenter:             &dc,
		StartCpus:              sl.Int(d.Get("cpu").(int)),
		MaxMemory:              sl.Int(d.Get("ram").(int)),
		NetworkComponents:      []datatypes.Virtual_Guest_Network_Component{networkComponent},
		BlockDevices:           getBlockDevices(d),
		LocalDiskFlag:          sl.Bool(d.Get("local_disk").(bool)),
		PostInstallScriptUri:   sl.String(d.Get("post_install_script_uri").(string)),
	}

	if dedicatedAcctHostOnly, ok := d.GetOk("dedicated_acct_host_only"); ok {
		opts.DedicatedAccountHostOnlyFlag = sl.Bool(dedicatedAcctHostOnly.(bool))
	}

	if imgId, ok := d.GetOk("image_id"); ok {
		imageId := imgId.(int)
		service := services.
			GetVirtualGuestBlockDeviceTemplateGroupService(meta.(*session.Session))

		image, err := service.
			Mask("id,globalIdentifier").Id(imageId).
			GetObject()
		if err != nil {
			return opts, fmt.Errorf("Error looking up image %d: %s", imageId, err)
		} else if image.GlobalIdentifier == nil {
			return opts, fmt.Errorf(
				"Image template %d does not have a global identifier", imageId)
		}

		opts.BlockDeviceTemplateGroup = &datatypes.Virtual_Guest_Block_Device_Template_Group{
			GlobalIdentifier: image.GlobalIdentifier,
		}
	}

	if operatingSystemReferenceCode, ok := d.GetOk("os_reference_code"); ok {
		opts.OperatingSystemReferenceCode = sl.String(operatingSystemReferenceCode.(string))
	}

	frontEndVlanNumber := d.Get("front_end_vlan.vlan_number").(string)
	frontEndSubnet := d.Get("front_end_subnet").(string)
	backEndVlanNumber := d.Get("back_end_vlan.vlan_number").(string)
	backEndVlanSubnet := d.Get("back_end_subnet").(string)

	if len(frontEndVlanNumber) > 0 || len(frontEndSubnet) > 0 {
		opts.PrimaryNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
			NetworkVlan: &datatypes.Network_Vlan{},
		}
	}

	// Apply frontend VLAN if provided
	if len(frontEndVlanNumber) > 0 {
		vlanNumber, err := strconv.Atoi(frontEndVlanNumber)
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		frontendVlanId, err := getVlanId(vlanNumber, d.Get("front_end_vlan.primary_router_hostname").(string), meta)
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		opts.PrimaryNetworkComponent.NetworkVlan.Id = sl.Int(frontendVlanId)
	}

	// Apply frontend subnet if provided
	if len(frontEndSubnet) > 0 {
		primarySubnetId, err := getSubnetId(frontEndSubnet, meta)
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		opts.PrimaryNetworkComponent.NetworkVlan.PrimarySubnetId = sl.Int(primarySubnetId)
	}

	if len(backEndVlanNumber) > 0 || len(backEndVlanSubnet) > 0 {
		opts.PrimaryBackendNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
			NetworkVlan: &datatypes.Network_Vlan{},
		}
	}

	// Apply backend VLAN if provided
	if len(d.Get("back_end_vlan.vlan_number").(string)) > 0 {
		vlanNumber, err := strconv.Atoi(d.Get("back_end_vlan.vlan_number").(string))
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		backendVlanId, err := getVlanId(vlanNumber, d.Get("back_end_vlan.primary_router_hostname").(string), meta)
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		opts.PrimaryBackendNetworkComponent.NetworkVlan.Id = sl.Int(backendVlanId)
	}

	// Apply backend subnet if provided
	if len(backEndVlanSubnet) > 0 {
		primarySubnetId, err := getSubnetId(backEndVlanSubnet, meta)
		if err != nil {
			return opts, fmt.Errorf("Error creating virtual guest: %s", err)
		}
		opts.PrimaryBackendNetworkComponent.NetworkVlan.PrimarySubnetId = sl.Int(primarySubnetId)
	}

	if userData, ok := d.GetOk("user_data"); ok {
		opts.UserData = []datatypes.Virtual_Guest_Attribute{
			{
				Value: sl.String(userData.(string)),
			},
		}
	}

	// Get configured ssh_keys
	ssh_keys := d.Get("ssh_keys").([]interface{})
	if len(ssh_keys) > 0 {
		opts.SshKeys = make([]datatypes.Security_Ssh_Key, 0, len(ssh_keys))
		for _, ssh_key := range ssh_keys {
			opts.SshKeys = append(opts.SshKeys, datatypes.Security_Ssh_Key{
				Id: sl.Int(ssh_key.(int)),
			})
		}
	}

	return opts, nil
}

func resourceSoftLayerVirtualGuestCreate(d *schema.ResourceData, meta interface{}) error {
	service := services.GetVirtualGuestService(meta.(*session.Session))

	opts, err := getVirtualGuestTemplateFromResourceData(d, meta)
	if err != nil {
		return err
	}

	log.Println("[INFO] Creating virtual machine")

	guest, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating virtual guest: %s", err)
	}

	id := *guest.Id
	d.SetId(fmt.Sprintf("%d", id))

	log.Printf("[INFO] Virtual Machine ID: %s", d.Id())

	// Set tags
	err = setGuestTags(id, d, meta)
	if err != nil {
		return err
	}

	// wait for machine availability
	_, err = WaitForNoActiveTransactions(d, meta)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to become ready: %s", d.Id(), err)
	}

	privateNetworkOnly := d.Get("private_network_only").(bool)
	if !privateNetworkOnly {
		_, err = WaitForPublicIpAvailable(d, meta)
		if err != nil {
			return fmt.Errorf(
				"Error waiting for virtual machine (%s) public ip to become ready: %s", d.Id(), err)
		}
	}

	return resourceSoftLayerVirtualGuestRead(d, meta)
}

func resourceSoftLayerVirtualGuestRead(d *schema.ResourceData, meta interface{}) error {
	service := services.GetVirtualGuestService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).Mask(
		"hostname,domain,startCpus,maxMemory,dedicatedAccountHostOnlyFlag," +
			"primaryIpAddress,primaryBackendIpAddress,privateNetworkOnlyFlag," +
			"hourlyBillingFlag,localDiskFlag," +
			"userData[value],tagReferences[id,tag[name]]," +
			"datacenter[id,name,longName]," +
			"primaryNetworkComponent[networkVlan[id,primaryRouter,vlanNumber],primaryIpAddressRecord[subnet,guestNetworkComponentBinding[ipAddressId]]]," +
			"primaryBackendNetworkComponent[networkVlan[id,primaryRouter,vlanNumber],primaryIpAddressRecord[subnet,guestNetworkComponentBinding[ipAddressId]]]",
	).GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving virtual guest: %s", err)
	}

	d.Set("name", *result.Hostname)
	d.Set("domain", *result.Domain)

	if result.Datacenter != nil {
		d.Set("datacenter", *result.Datacenter.Name)
	}

	d.Set("network_speed", *result.PrimaryNetworkComponent.MaxSpeed)
	d.Set("cpu", *result.StartCpus)
	d.Set("ram", *result.MaxMemory)
	d.Set("dedicated_acct_host_only", *result.DedicatedAccountHostOnlyFlag)
	if result.PrimaryIpAddress != nil {
		d.Set("has_public_ip", *result.PrimaryIpAddress != "")
		d.Set("ipv4_address", *result.PrimaryIpAddress)
	}
	d.Set("ipv4_address_private", *result.PrimaryBackendIpAddress)
	if result.PrimaryNetworkComponent.PrimaryIpAddressRecord != nil {
		d.Set("ip_address_id", *result.PrimaryNetworkComponent.PrimaryIpAddressRecord.GuestNetworkComponentBinding.IpAddressId)
	}
	d.Set("ip_address_id_private",
		*result.PrimaryBackendNetworkComponent.PrimaryIpAddressRecord.GuestNetworkComponentBinding.IpAddressId)
	d.Set("private_network_only", *result.PrivateNetworkOnlyFlag)
	d.Set("hourly_billing", *result.HourlyBillingFlag)
	d.Set("local_disk", *result.LocalDiskFlag)

	if result.PrimaryNetworkComponent.NetworkVlan != nil {
		frontEndVlan := d.Get("front_end_vlan").(map[string]interface{})
		resultFrontEndVlan := result.PrimaryNetworkComponent.NetworkVlan
		frontEndVlan["primary_router_hostname"] = *resultFrontEndVlan.PrimaryRouter.Hostname
		frontEndVlan["vlan_number"] = strconv.Itoa(*resultFrontEndVlan.VlanNumber)
		d.Set("front_end_vlan", frontEndVlan)
	}

	backEndVlan := d.Get("back_end_vlan").(map[string]interface{})
	resultBackEndVlan := result.PrimaryBackendNetworkComponent.NetworkVlan
	backEndVlan["primary_router_hostname"] = *resultBackEndVlan.PrimaryRouter.Hostname
	backEndVlan["vlan_number"] = strconv.Itoa(*resultBackEndVlan.VlanNumber)
	d.Set("back_end_vlan", backEndVlan)

	if result.PrimaryNetworkComponent.PrimaryIpAddressRecord != nil {
		resultFrontendSubnet := result.PrimaryNetworkComponent.PrimaryIpAddressRecord.Subnet
		d.Set("front_end_subnet", *resultFrontendSubnet.NetworkIdentifier+"/"+strconv.Itoa(*resultFrontendSubnet.Cidr))
	}

	resultBackendSubnet := result.PrimaryBackendNetworkComponent.PrimaryIpAddressRecord.Subnet
	d.Set("back_end_subnet", *resultBackendSubnet.NetworkIdentifier+"/"+strconv.Itoa(*resultBackendSubnet.Cidr))

	userData := result.UserData
	if userData != nil && len(userData) > 0 {
		data, err := base64.StdEncoding.DecodeString(*userData[0].Value)
		if err != nil {
			log.Printf("Can't base64 decode user data %s. error: %s", *userData[0].Value, err)
			d.Set("user_data", userData)
		} else {
			d.Set("user_data", string(data))
		}
	}

	tagReferences := result.TagReferences
	if len(tagReferences) > 0 {
		tags := []string{}
		for _, tagRef := range tagReferences {
			tags = append(tags, *tagRef.Tag.Name)
		}
		d.Set("tags", tags)
	}

	return nil
}

func resourceSoftLayerVirtualGuestUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetVirtualGuestService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(id).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving virtual guest: %s", err)
	}

	// Update "name" and "domain" fields if present and changed
	// Those are the only fields, which could be updated
	if d.HasChange("name") || d.HasChange("domain") {
		result.Hostname = sl.String(d.Get("name").(string))
		result.Domain = sl.String(d.Get("domain").(string))

		_, err = service.Id(id).EditObject(&result)

		if err != nil {
			return fmt.Errorf("Couldn't update virtual guest: %s", err)
		}
	}

	// Set user data if provided and not empty
	if d.HasChange("user_data") {
		_, err := service.Id(id).SetUserMetadata([]string{d.Get("user_data").(string)})
		if err != nil {
			return fmt.Errorf("Couldn't update user data for virtual guest: %s", err)
		}
	}

	// Update tags
	if d.HasChange("tags") {
		err := setGuestTags(id, d, meta)
		if err != nil {
			return err
		}
	}

	// Upgrade "cpu", "ram" and "nic_speed" if provided and changed
	upgradeOptions := map[string]float64{}
	if d.HasChange("cpu") {
		upgradeOptions[product.CPUCategoryCode] = float64(d.Get("cpu").(int))
	}

	if d.HasChange("ram") {
		memoryInMB := float64(d.Get("ram").(int))

		// Convert memory to GB, as softlayer only allows to upgrade RAM in Gigs
		// Must be already validated at this step
		upgradeOptions[product.MemoryCategoryCode] = float64(int(memoryInMB / 1024))
	}

	if d.HasChange("network_speed") {
		upgradeOptions[product.NICSpeedCategoryCode] = float64(d.Get("network_speed").(int))
	}

	if len(upgradeOptions) > 0 {
		_, err = virtual.UpgradeVirtualGuest(sess, &result, upgradeOptions)
		if err != nil {
			return fmt.Errorf("Couldn't upgrade virtual guest: %s", err)
		}

		// Wait for softlayer to start upgrading...
		_, err = WaitForUpgradeTransactionsToAppear(d, meta)

		// Wait for upgrade transactions to finish
		_, err = WaitForNoActiveTransactions(d, meta)

		return err
	}

	return nil
}

func resourceSoftLayerVirtualGuestDelete(d *schema.ResourceData, meta interface{}) error {
	service := services.GetVirtualGuestService(meta.(*session.Session))

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = WaitForNoActiveTransactions(d, meta)

	if err != nil {
		return fmt.Errorf("Error deleting virtual guest, couldn't wait for zero active transactions: %s", err)
	}

	_, err = service.Id(id).DeleteObject()

	if err != nil {
		return fmt.Errorf("Error deleting virtual guest: %s", err)
	}

	return nil
}

func WaitForUpgradeTransactionsToAppear(d *schema.ResourceData, meta interface{}) (interface{}, error) {

	log.Printf("Waiting for server (%s) to have upgrade transactions", d.Id())

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The instance ID %s must be numeric", d.Id())
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending_upgrade"},
		Target:  []string{"upgrade_started"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetVirtualGuestService(meta.(*session.Session))
			transactions, err := service.Id(id).GetActiveTransactions()
			if err != nil {
				return nil, "", fmt.Errorf("Couldn't fetch active transactions: %s", err)
			}
			for _, transaction := range transactions {
				if strings.Contains(*transaction.TransactionStatus.Name, "UPGRADE") {
					return transactions, "upgrade_started", nil
				}
			}
			return transactions, "pending_upgrade", nil
		},
		Timeout:    5 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func WaitForPublicIpAvailable(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for server (%s) to get a public IP", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "unavailable"},
		Target:  []string{"available"},
		Refresh: func() (interface{}, string, error) {
			fmt.Println("Refreshing server state...")
			service := services.GetVirtualGuestService(meta.(*session.Session))
			id, err := strconv.Atoi(d.Id())
			if err != nil {
				return nil, "", fmt.Errorf("Not a valid ID, must be an integer: %s", err)
			}
			result, err := service.Id(id).GetObject()
			if err != nil {
				return nil, "", fmt.Errorf("Error retrieving virtual guest: %s", err)
			}
			if result.PrimaryIpAddress == nil || *result.PrimaryIpAddress == "" {
				return result, "unavailable", nil
			} else {
				return result, "available", nil
			}
		},
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func WaitForNoActiveTransactions(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for server (%s) to have zero active transactions", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The instance ID %s must be numeric", d.Id())
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "active"},
		Target:  []string{"idle"},
		Refresh: func() (interface{}, string, error) {
			service := services.GetVirtualGuestService(meta.(*session.Session))
			transactions, err := service.Id(id).GetActiveTransactions()
			if err != nil {
				return nil, "", fmt.Errorf("Couldn't get active transactions: %s", err)
			}
			if len(transactions) == 0 {
				return transactions, "idle", nil
			} else {
				return transactions, "active", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func resourceSoftLayerVirtualGuestExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	service := services.GetVirtualGuestService(meta.(*session.Session))

	guestId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(guestId).GetObject()
	return result.Id != nil && err == nil && *result.Id == guestId, nil
}

func getTags(d *schema.ResourceData) string {
	tagSet := d.Get("tags").(*schema.Set)

	if tagSet.Len() == 0 {
		return ""
	}

	tags := make([]string, 0, tagSet.Len())
	for _, elem := range tagSet.List() {
		tag := elem.(string)
		tags = append(tags, tag)
	}
	return strings.Join(tags, ",")
}

func setGuestTags(id int, d *schema.ResourceData, meta interface{}) error {
	service := services.GetVirtualGuestService(meta.(*session.Session))

	tags := getTags(d)
	if tags != "" {
		_, err := service.Id(id).SetTags(sl.String(tags))
		if err != nil {
			return fmt.Errorf("Could not set tags on virtual guest %d", id)
		}
	}

	return nil
}
