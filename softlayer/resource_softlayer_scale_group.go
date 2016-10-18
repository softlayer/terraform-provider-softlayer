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
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const HEALTH_CHECK_TYPE_HTTP_CUSTOM = "HTTP-CUSTOM"

var SoftLayerScaleGroupObjectMask = []string{
	"id",
	"name",
	"minimumMemberCount",
	"maximumMemberCount",
	"cooldown",
	"status[keyName]",
	"regionalGroup[id,name]",
	"terminationPolicy[keyName]",
	"virtualGuestMemberTemplate[blockDeviceTemplateGroup,primaryNetworkComponent[networkVlan[id]],primaryBackendNetworkComponent[networkVlan[id]]]",
	"loadBalancers[id,port,virtualServerId,healthCheck[id]]",
	"networkVlans[id,networkVlanId,networkVlan[vlanNumber,primaryRouter[hostname]]]",
	"loadBalancers[healthCheck[healthCheckTypeId,type[keyname],attributes[value,type[id,keyname]]]]",
}

func resourceSoftLayerScaleGroup() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerScaleGroupCreate,
		Read:     resourceSoftLayerScaleGroupRead,
		Update:   resourceSoftLayerScaleGroupUpdate,
		Delete:   resourceSoftLayerScaleGroupDelete,
		Exists:   resourceSoftLayerScaleGroupExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"regional_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"minimum_member_count": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"maximum_member_count": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"cooldown": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"termination_policy": {
				Type:     schema.TypeString,
				Required: true,
			},

			"virtual_server_id": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"health_check": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: false,
						},

						// Conditionally-required fields, based on value of "type"
						"custom_method": {
							Type:     schema.TypeString,
							Optional: true,
							// TODO: Must be GET or HEAD
						},

						"custom_request": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"custom_response": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// This has to be a TypeList, because TypeMap does not handle non-primitive
			// members properly.
			"virtual_guest_member_template": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     getModifiedVirtualGuestResource(),
			},

			"network_vlan_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

// Returns a modified version of the virtual guest resource, with all members set to ForceNew = false.
// Otherwise a modified template parameter unnecessarily forces scale group drop/create
func getModifiedVirtualGuestResource() *schema.Resource {

	r := resourceSoftLayerVirtualGuest()

	for _, elem := range r.Schema {
		elem.ForceNew = false
	}

	return r
}

// Helper method to parse healthcheck data in the resource schema format to the SoftLayer datatypes
func buildHealthCheckFromResourceData(d map[string]interface{}) (datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check, error) {
	healthCheckOpts := datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{
		Type: &datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check_Type{
			Keyname: sl.String(d["type"].(string)),
		},
	}

	if *healthCheckOpts.Type.Keyname == HEALTH_CHECK_TYPE_HTTP_CUSTOM {
		// Validate and apply type-specific fields
		healthCheckMethod, ok := d["custom_method"]
		if !ok {
			return datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{}, errors.New("\"custom_method\" is required when HTTP-CUSTOM healthcheck is specified")
		}

		healthCheckRequest, ok := d["custom_request"]
		if !ok {
			return datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{}, errors.New("\"custom_request\" is required when HTTP-CUSTOM healthcheck is specified")
		}

		healthCheckResponse, ok := d["custom_response"]
		if !ok {
			return datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{}, errors.New("\"custom_response\" is required when HTTP-CUSTOM healthcheck is specified")
		}

		// HTTP-CUSTOM values are represented as an array of SoftLayer_Health_Check_Attributes
		healthCheckOpts.Attributes = []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Attribute{
			{
				Type: &datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Attribute_Type{
					Keyname: sl.String("HTTP_CUSTOM_TYPE"),
				},
				Value: sl.String(healthCheckMethod.(string)),
			},
			{
				Type: &datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Attribute_Type{
					Keyname: sl.String("LOCATION"),
				},
				Value: sl.String(healthCheckRequest.(string)),
			},
			{
				Type: &datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Attribute_Type{
					Keyname: sl.String("EXPECTED_RESPONSE"),
				},
				Value: sl.String(healthCheckResponse.(string)),
			},
		}
	}

	return healthCheckOpts, nil
}

// Helper method to parse network vlan information in the resource schema format to the SoftLayer datatypes
func buildScaleVlansFromResourceData(v interface{}, meta interface{}) ([]datatypes.Scale_Network_Vlan, error) {
	vlanIds := v.([]interface{})
	scaleNetworkVlans := make([]datatypes.Scale_Network_Vlan, 0, len(vlanIds))

	for _, iVlanId := range vlanIds {
		vlanId := iVlanId.(int)
		scaleNetworkVlans = append(
			scaleNetworkVlans,
			datatypes.Scale_Network_Vlan{NetworkVlanId: &vlanId},
		)
	}

	return scaleNetworkVlans, nil
}

func getVirtualGuestTemplate(vGuestTemplateList []interface{}, meta interface{}) (datatypes.Virtual_Guest, error) {
	if len(vGuestTemplateList) != 1 {
		return datatypes.Virtual_Guest{},
			errors.New("Only one virtual_guest_member_template can be provided")
	}

	// Retrieve the map of virtual_guest_member_template attributes
	vGuestMap := vGuestTemplateList[0].(map[string]interface{})

	// Create an empty ResourceData instance for a SoftLayer_Virtual_Guest resource
	vGuestResourceData := resourceSoftLayerVirtualGuest().Data(nil)

	// For each item in the map, call Set on the ResourceData.  This handles
	// validation and yields a completed ResourceData object
	for k, v := range vGuestMap {
		log.Printf("****** %s: %#v", k, v)
		err := vGuestResourceData.Set(k, v)
		if err != nil {
			return datatypes.Virtual_Guest{},
				fmt.Errorf("Error while parsing virtual_guest_member_template values: %s", err)
		}
	}

	// Get the virtual guest creation template from the completed resource data object
	return getVirtualGuestTemplateFromResourceData(vGuestResourceData, meta)
}

func resourceSoftLayerScaleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetScaleGroupService(sess)

	virtualGuestTemplateOpts, err := getVirtualGuestTemplate(d.Get("virtual_guest_member_template").([]interface{}), meta)
	if err != nil {
		return fmt.Errorf("Error while parsing virtual_guest_member_template values: %s", err)
	}

	scaleNetworkVlans, err := buildScaleVlansFromResourceData(d.Get("network_vlan_ids"), meta)
	if err != nil {
		return fmt.Errorf("Error while parsing network vlan values: %s", err)
	}

	locationGroupRegionalId, err := getLocationGroupRegionalId(sess, d.Get("regional_group").(string))
	if err != nil {
		return err
	}

	// Build up our creation options
	opts := datatypes.Scale_Group{
		Name:                       sl.String(d.Get("name").(string)),
		Cooldown:                   sl.Int(d.Get("cooldown").(int)),
		MinimumMemberCount:         sl.Int(d.Get("minimum_member_count").(int)),
		MaximumMemberCount:         sl.Int(d.Get("maximum_member_count").(int)),
		SuspendedFlag:              sl.Bool(false),
		VirtualGuestMemberTemplate: &virtualGuestTemplateOpts,
		NetworkVlans:               scaleNetworkVlans,
		RegionalGroupId:            &locationGroupRegionalId,
	}

	opts.TerminationPolicy = &datatypes.Scale_Termination_Policy{
		KeyName: sl.String(d.Get("termination_policy").(string)),
	}

	healthCheckOpts, err := buildHealthCheckFromResourceData(d.Get("health_check").(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Error while parsing health check options: %s", err)
	}

	opts.LoadBalancers = []datatypes.Scale_LoadBalancer{
		{
			HealthCheck:     &healthCheckOpts,
			Port:            sl.Int(d.Get("port").(int)),
			VirtualServerId: sl.Int(d.Get("virtual_server_id").(int)),
		},
	}

	res, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating Scale Group: %s", err)
	}

	d.SetId(strconv.Itoa(*res.Id))
	log.Printf("[INFO] Scale Group ID: %d", *res.Id)

	// wait for scale group to become active
	_, err = waitForActiveStatus(d, meta)

	if err != nil {
		return fmt.Errorf("Error waiting for scale group (%s) to become active: %s", d.Id(), err)
	}

	return resourceSoftLayerScaleGroupRead(d, meta)
}

func resourceSoftLayerScaleGroupRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetScaleGroupService(sess)

	groupId, _ := strconv.Atoi(d.Id())

	slGroupObj, err := service.Id(groupId).Mask(strings.Join(SoftLayerScaleGroupObjectMask, ",")).GetObject()
	if err != nil {
		// If the scale group is somehow already destroyed, mark as successfully gone
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving SoftLayer Scale Group: %s", err)
	}

	d.Set("id", *slGroupObj.Id)
	d.Set("name", *slGroupObj.Name)
	d.Set("regional_group", *slGroupObj.RegionalGroup.Name)
	d.Set("minimum_member_count", *slGroupObj.MinimumMemberCount)
	d.Set("maximum_member_count", *slGroupObj.MaximumMemberCount)
	d.Set("cooldown", *slGroupObj.Cooldown)
	d.Set("status", *slGroupObj.Status.KeyName)
	d.Set("termination_policy", *slGroupObj.TerminationPolicy.KeyName)
	d.Set("virtual_server_id", *slGroupObj.LoadBalancers[0].VirtualServerId)
	d.Set("port", *slGroupObj.LoadBalancers[0].Port)

	// Health Check
	healthCheckObj := slGroupObj.LoadBalancers[0].HealthCheck
	currentHealthCheck := d.Get("health_check").(map[string]interface{})

	currentHealthCheck["type"] = *healthCheckObj.Type.Keyname

	if *healthCheckObj.Type.Keyname == HEALTH_CHECK_TYPE_HTTP_CUSTOM {
		for _, elem := range healthCheckObj.Attributes {
			switch *elem.Type.Keyname {
			case "HTTP_CUSTOM_TYPE":
				currentHealthCheck["custom_method"] = *elem.Value
			case "LOCATION":
				currentHealthCheck["custom_request"] = *elem.Value
			case "EXPECTED_RESPONSE":
				currentHealthCheck["custom_response"] = *elem.Value
			}
		}
	}

	d.Set("health_check", currentHealthCheck)

	// Network Vlans
	vlanTotal := len(slGroupObj.NetworkVlans)
	// Don't refresh vlan ids, unless this is an import operation
	// to avoid the problem of getting the list out of order from
	// the original order in the config and trigger a false change/update.
	// This should be fine as we don't expect this to change after the group
	// is created. Else, this code needs to be refactored to use a TypeSet.
	if vlanTotal == 0 {
		vlanIds := make([]int, vlanTotal)
		for i, vlan := range slGroupObj.NetworkVlans {
			vlanIds[i] = *vlan.NetworkVlanId
		}
		d.Set("network_vlan_ids", vlanIds)
	}

	virtualGuestTemplate := populateMemberTemplateResourceData(*slGroupObj.VirtualGuestMemberTemplate)
	d.Set("virtual_guest_member_template", virtualGuestTemplate)

	return nil
}

func populateMemberTemplateResourceData(template datatypes.Virtual_Guest) map[string]interface{} {

	d := make(map[string]interface{})

	d["name"] = *template.Hostname
	d["domain"] = *template.Domain
	d["datacenter"] = *template.Datacenter.Name
	d["network_speed"] = *template.NetworkComponents[0].MaxSpeed
	d["cpu"] = *template.StartCpus
	d["ram"] = *template.MaxMemory
	d["private_network_only"] = *template.PrivateNetworkOnlyFlag
	d["hourly_billing"] = *template.HourlyBillingFlag
	d["local_disk"] = *template.LocalDiskFlag

	// Guard against nil values for optional fields in virtual_guest resource
	d["dedicated_acct_host_only"] = sl.Get(template.DedicatedAccountHostOnlyFlag)
	d["os_reference_code"] = sl.Get(template.OperatingSystemReferenceCode)
	d["post_install_script_uri"] = sl.Get(template.PostInstallScriptUri)

	if template.PrimaryNetworkComponent != nil && template.PrimaryNetworkComponent.NetworkVlan != nil {
		d["frontend_vlan_id"] = sl.Get(template.PrimaryNetworkComponent.NetworkVlan.Id)
	} else {
		d["frontend_vlan_id"] = nil
	}

	if template.PrimaryBackendNetworkComponent != nil && template.PrimaryBackendNetworkComponent.NetworkVlan != nil {
		d["backend_vlan_id"] = sl.Get(template.PrimaryBackendNetworkComponent.NetworkVlan.Id)
	} else {
		d["backend_vlan_id"] = nil
	}

	if template.BlockDeviceTemplateGroup != nil {
		d["block_device_template_group_gid"] = sl.Get(template.BlockDeviceTemplateGroup.GlobalIdentifier)
	} else {
		d["block_device_template_group_gid"] = nil
	}

	if len(template.UserData) > 0 {
		d["user_data"] = *template.UserData[0].Value
	} else {
		d["user_data"] = ""
	}

	sshKeys := make([]interface{}, 0, len(template.SshKeys))
	for _, elem := range template.SshKeys {
		sshKeys = append(sshKeys, *elem.Id)
	}
	d["ssh_keys"] = sshKeys

	disks := make([]interface{}, 0, len(template.BlockDevices))
	for _, elem := range template.BlockDevices {
		disks = append(disks, *elem.DiskImage.Capacity)
	}
	d["disks"] = disks

	return d
}

func resourceSoftLayerScaleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	scaleGroupService := services.GetScaleGroupService(sess)
	scaleNetworkVlanService := services.GetScaleNetworkVlanService(sess)

	groupId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID. Must be an integer: %s", err)
	}

	// Fetch the complete object from SoftLayer, update with current values from the configuration, and send the
	// whole thing back to SoftLayer (effectively, a PUT)
	groupObj, err := scaleGroupService.Id(groupId).Mask(strings.Join(SoftLayerScaleGroupObjectMask, ",")).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving softlayer_scale_group resource: %s", err)
	}

	groupObj.Name = sl.String(d.Get("name").(string))
	groupObj.MinimumMemberCount = sl.Int(d.Get("minimum_member_count").(int))
	groupObj.MaximumMemberCount = sl.Int(d.Get("maximum_member_count").(int))
	groupObj.Cooldown = sl.Int(d.Get("cooldown").(int))
	groupObj.TerminationPolicy.KeyName = sl.String(d.Get("termination_policy").(string))
	groupObj.LoadBalancers[0].VirtualServerId = sl.Int(d.Get("virtual_server_id").(int))
	groupObj.LoadBalancers[0].Port = sl.Int(d.Get("port").(int))

	healthCheck, err := buildHealthCheckFromResourceData(d.Get("health_check").(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Unable to parse health check options: %s", err)
	}

	groupObj.LoadBalancers[0].HealthCheck = &healthCheck

	if d.HasChange("network_vlan_ids") {
		// Vlans require special handling:
		//
		// 1. Delete any scale_network_vlans which no longer appear in the updated configuration
		// 2. Pass the updated list of vlans to the Scale_Group.editObject function.  SoftLayer determines
		// which Vlans are new, and which already exist.

		oldIds, newIds := d.GetChange("network_vlan_ids")

		// Delete entries from 'old' set not appearing in new (old - new)
		for _, o := range oldIds.([]int) {
			for _, n := range newIds.([]int) {
				if n == o {
					goto nextOld
				}
			}

			_, err = scaleNetworkVlanService.Id(o).DeleteObject()
			if err != nil {
				return fmt.Errorf("Error deleting scale network vlan: %s", err)
			}
		nextOld:
		}

		// Parse the new list of vlans into the appropriate input structure
		scaleVlans, err := buildScaleVlansFromResourceData(newIds, meta)

		if err != nil {
			return fmt.Errorf("Unable to parse network vlan options: %s", err)
		}

		groupObj.NetworkVlans = scaleVlans
	}

	if d.HasChange("virtual_guest_member_template") {
		virtualGuestTemplateOpts, err := getVirtualGuestTemplate(d.Get("virtual_guest_member_template").([]interface{}), meta)
		if err != nil {
			return fmt.Errorf("Unable to parse virtual guest member template options: %s", err)
		}

		groupObj.VirtualGuestMemberTemplate = &virtualGuestTemplateOpts

	}
	_, err = scaleGroupService.Id(groupId).EditObject(&groupObj)
	if err != nil {
		return fmt.Errorf("Error received while editing softlayer_scale_group: %s", err)
	}

	// wait for scale group to become active
	_, err = waitForActiveStatus(d, meta)

	if err != nil {
		return fmt.Errorf("Error waiting for scale group (%s) to become active: %s", d.Id(), err)
	}

	return nil
}

func resourceSoftLayerScaleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	scaleGroupService := services.GetScaleGroupService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting scale group: %s", err)
	}

	log.Printf("[INFO] Deleting scale group: %d", id)
	_, err = scaleGroupService.Id(id).ForceDeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting scale group: %s", err)
	}

	d.SetId("")

	return nil
}

func waitForActiveStatus(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	sess := meta.(*session.Session)
	scaleGroupService := services.GetScaleGroupService(sess)

	log.Printf("Waiting for scale group (%s) to become active", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The scale group ID %s must be numeric", d.Id())
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"BUSY", "SCALING", "SUSPENDED"},
		Target:  []string{"ACTIVE"},
		Refresh: func() (interface{}, string, error) {
			// get the status of the scale group
			result, err := scaleGroupService.Id(id).Mask("status.keyName").GetObject()

			log.Printf("The status of scale group with id (%s) is (%s)", d.Id(), *result.Status.KeyName)
			if err != nil {
				return nil, "", fmt.Errorf("Couldn't get status of the scale group: %s", err)
			}

			return result, *result.Status.KeyName, nil
		},
		Timeout:    10 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	return stateConf.WaitForState()
}

func resourceSoftLayerScaleGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	scaleGroupService := services.GetScaleGroupService(sess)

	groupId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := scaleGroupService.Id(groupId).Mask("id").GetObject()
	return result.Id != nil && err == nil && *result.Id == groupId, nil
}

func getLocationGroupRegionalId(sess *session.Session, locationGroupRegionalName string) (int, error) {
	locationGroupRegionals, err := services.GetLocationGroupRegionalService(sess).
		Mask("id,name").
		// FIXME: Someday, filters may actually work in SoftLayer
		//Filter(filter.Build(
		//	filter.Path("name").Eq(locationGroupRegionalName))).
		//Limit(1).
		GetAllObjects()

	if err != nil {
		return -1, err
	}

	if len(locationGroupRegionals) < 1 {
		return -1, fmt.Errorf("Invalid location group regional: %s", locationGroupRegionalName)
	}

	for _, locationGroupRegional := range locationGroupRegionals {
		if *locationGroupRegional.Name == locationGroupRegionalName {
			return *locationGroupRegional.Id, nil
		}
	}

	return -1, fmt.Errorf("Invalid regional_group_id: %s", locationGroupRegionalName)
}
