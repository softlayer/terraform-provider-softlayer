package softlayer

import (
	"fmt"
	"log"

	"bytes"

	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/filter"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
	"github.ibm.com/riethm/gopherlayer.git/sl"
)

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddress() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerNetworkLoadBalancerVirtualIpAddressCreate,
		Read:     resourceSoftLayerNetworkLoadBalancerVirtualIpAddressRead,
		Update:   resourceSoftLayerNetworkLoadBalancerVirtualIpAddressUpdate,
		Delete:   resourceSoftLayerNetworkLoadBalancerVirtualIpAddressDelete,
		Exists:   resourceSoftLayerNetworkLoadBalancerVirtualIpAddressExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"nad_controller_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"connection_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"load_balancing_method": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"modify_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			// name field is actually used as an ID in SoftLayer
			// http://sldn.softlayer.com/reference/services/SoftLayer_Network_Application_Delivery_Controller/updateLiveLoadBalancer
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"security_certificate_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"source_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"virtual_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddressCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)

	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		ConnectionLimit:       sl.Int(d.Get("connection_limit").(int)),
		LoadBalancingMethod:   sl.String(d.Get("load_balancing_method").(string)),
		Name:                  sl.String(d.Get("name").(string)),
		SourcePort:            sl.Int(d.Get("source_port").(int)),
		Type:                  sl.String(d.Get("type").(string)),
		VirtualIpAddress:      sl.String(d.Get("virtual_ip_address").(string)),
		SecurityCertificateId: sl.Int(d.Get("security_certificate_id").(int)),
	}

	log.Printf("[INFO] Creating Virtual Ip Address %s", template.VirtualIpAddress)

	successFlag, err := service.Id(nadcId).CreateLiveLoadBalancer(&template)

	if err != nil {
		return fmt.Errorf("Error creating Virtual Ip Address: %s", err)
	}

	if !successFlag {
		return fmt.Errorf("Error creating Virtual Ip Address")
	}

	return resourceSoftLayerNetworkLoadBalancerVirtualIpAddressRead(d, meta)
}

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddressRead(d *schema.ResourceData, meta interface{}) error {
	nadcId := d.Get("nad_controller_id").(int)
	vipName := d.Get("name").(string)

	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	vips, err := service.
		Id(nadcId).
		Filter(filter.Path("name").Eq(vipName).Build()).
		GetLoadBalancers()
	if err != nil {
		return fmt.Errorf("Error getting Virtual Ip Address: %s", err)
	}

	if len(vips) == 0 {
		return fmt.Errorf("Could not find any VIPs for NADC %d matching name %s", nadcId, vipName)
	}
	vip := vips[0]

	d.SetId(fmt.Sprintf("%s;%d", *vip.Name, nadcId))
	d.Set("nad_controller_id", nadcId)
	d.Set("load_balancing_method", *vip.LoadBalancingMethod)
	d.Set("load_balancing_method_name", *vip.LoadBalancingMethodFullName)
	d.Set("modify_date", *vip.ModifyDate)
	d.Set("name", *vip.Name)
	d.Set("connection_limit", *vip.ConnectionLimit)
	d.Set("security_certificate_id", *vip.SecurityCertificateId)
	d.Set("source_port", *vip.SourcePort)
	d.Set("type", *vip.Type)
	d.Set("virtual_ip_address", *vip.VirtualIpAddress)

	return nil
}

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddressUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		Name: sl.String(d.Get("name").(string)),
	}

	if d.HasChange("load_balancing_method") {
		template.LoadBalancingMethod = sl.String(d.Get("load_balancing_method").(string))
	}

	if d.HasChange("security_certificate_id") {
		template.SecurityCertificateId = sl.Int(d.Get("security_certificate_id").(int))
	}

	if d.HasChange("source_port") {
		template.SourcePort = sl.Int(d.Get("source_port").(int))
	}

	if d.HasChange("type") {
		template.Type = sl.String(d.Get("type").(string))
	}

	if d.HasChange("virtual_ip_address") {
		template.VirtualIpAddress = sl.String(d.Get("virtual_ip_address").(string))
	}

	_, err := service.Id(nadcId).UpdateLiveLoadBalancer(&template)

	if err != nil {
		return fmt.Errorf("Error updating Virtual Ip Address: %s", err)
	}

	return nil
}

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddressDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	vipName := d.Get("name").(string)

	_, err := service.Id(nadcId).DeleteLiveLoadBalancer(
		&datatypes.Network_LoadBalancer_VirtualIpAddress{Name: sl.String(vipName)},
	)
	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address %s: %s", vipName, err)
	}

	return nil
}

func resourceSoftLayerNetworkLoadBalancerVirtualIpAddressExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	vipName := d.Get("name").(string)
	nadcId := d.Get("nad_controller_id").(int)

	vips, err := service.
		Id(nadcId).
		Filter(filter.Path("name").Eq(vipName).Build()).
		GetLoadBalancers()
	if err != nil {
		return false, fmt.Errorf("Error fetching Virtual Ip Address: %s", err)
	}

	if len(vips) == 0 {
		return false, fmt.Errorf("Could not find any VIPs for NADC %d matching name %s", nadcId, vipName)
	}
	vip := vips[0]

	return *vip.Name == vipName && err == nil, nil
}
