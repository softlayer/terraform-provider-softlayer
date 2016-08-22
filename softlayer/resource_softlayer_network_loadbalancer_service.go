package softlayer

import (
	"errors"
	"fmt"
	"log"

	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
	"github.ibm.com/riethm/gopherlayer.git/sl"
	"github.ibm.com/riethm/gopherlayer.git/helpers/network"
)

func resourceSoftLayerNetworkLoadBalancerService() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerNetworkLoadBalancerServiceCreate,
		Read:     resourceSoftLayerNetworkLoadBalancerServiceRead,
		Update:   resourceSoftLayerNetworkLoadBalancerServiceUpdate,
		Delete:   resourceSoftLayerNetworkLoadBalancerServiceDelete,
		Exists:   resourceSoftLayerNetworkLoadBalancerServiceExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},

			"vip_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"destination_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"destination_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"weight": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},

			"connection_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},

			"health_check": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func parseVipUniqueId(vipUniqueId string) (string, int, error) {
	parts := strings.Split(vipUniqueId, ";")
	vipId := parts[0]
	nacdId, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", -1, fmt.Errorf("Error parsing vip id: %s", err)
	}

	return vipId, nacdId, nil
}

func resourceSoftLayerNetworkLoadBalancerServiceCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	vipUniqueId := d.Get("vip_id").(string)

	vipId, nadcId, err := parseVipUniqueId(vipUniqueId)

	if err != nil {
		return fmt.Errorf("Error parsing vip id: %s", err)
	}

	lb_services := []datatypes.Network_LoadBalancer_Service{
		{
			Name:                 sl.String(d.Get("name").(string)),
			DestinationIpAddress: sl.String(d.Get("destination_ip_address").(string)),
			DestinationPort:      sl.Int(d.Get("destination_port").(int)),
			Weight:               sl.Int(d.Get("weight").(int)),
			HealthCheck:          sl.String(d.Get("health_check").(string)),
			ConnectionLimit:      sl.Int(d.Get("connection_limit").(int)),
		},
	}

	lb_vip := &datatypes.Network_LoadBalancer_VirtualIpAddress{
		Name:     sl.String(vipId),
		Services: lb_services,
	}

	log.Printf("[INFO] Creating LoadBalancer Service %s", lb_services[0].Name)

	// TODO: This API call might return an error until the LB VIP is ready to be updated with a service.
	// In that case, would need to inspect the api return code and use StateChangeConf to retry or a new helper.
	// See CreateLoadBalancerService in softlayer-go as a reference.
	successFlag, err := service.Id(nadcId).UpdateLiveLoadBalancer(lb_vip)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancer Service: %s", err)
	}

	if !successFlag {
		return errors.New("Error creating LoadBalancer Service")
	}

	return resourceSoftLayerNetworkLoadBalancerServiceRead(d, meta)
}

func resourceSoftLayerNetworkLoadBalancerServiceRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipUniqueId := d.Get("vip_id").(string)

	vipName, nadcId, err := parseVipUniqueId(vipUniqueId)
	if err != nil {
		return fmt.Errorf("Error parsing vip id: %s", err)
	}

	lbService, err := network.GetNadcLbVipServiceByName(sess, nadcId, vipName, d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("Unable to get load balancer service: %s", err)
	}

	d.SetId(*lbService.Name)
	d.Set("name", *lbService.Name)
	d.Set("destination_ip_address", *lbService.DestinationIpAddress)
	d.Set("destination_port", *lbService.DestinationPort)
	d.Set("weight", *lbService.Weight)
	d.Set("health_check", *lbService.HealthCheck)
	d.Set("connection_limit", *lbService.ConnectionLimit)

	return nil
}

func resourceSoftLayerNetworkLoadBalancerServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	vipUniqueId := d.Get("vip_id").(string)
	vipName, nadcId, err := parseVipUniqueId(vipUniqueId)
	if err != nil {
		return fmt.Errorf("Error parsing vip id: %s", err)
	}

	lbService, err := network.GetNadcLbVipServiceByName(sess, nadcId, vipName, d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("Unable to get load balancer service: %s", err)
	}

	// copy current service
	template := datatypes.Network_LoadBalancer_Service(*lbService)

	if data, ok := d.GetOk("name"); ok {
		template.Name = sl.String(data.(string))
	}
	if data, ok := d.GetOk("destination_ip_address"); ok {
		template.DestinationIpAddress = sl.String(data.(string))
	}
	if data, ok := d.GetOk("destination_port"); ok {
		template.DestinationPort = sl.Int(data.(int))
	}
	if data, ok := d.GetOk("weight"); ok {
		template.Weight = sl.Int(data.(int))
	}
	if data, ok := d.GetOk("health_check"); ok {
		template.HealthCheck = sl.String(data.(string))
	}
	if data, ok := d.GetOk("connection_limit"); ok {
		template.ConnectionLimit = sl.Int(data.(int))
	}

	// TODO: This API call might return an error until the LB VIP is ready to be updated with a service.
	// In that case, would need to inspect the api return code and use StateChangeConf to retry or a new helper.
	// See CreateLoadBalancerService in softlayer-go as a reference.
	_, err = service.Id(nadcId).UpdateLiveLoadBalancer(&datatypes.Network_LoadBalancer_VirtualIpAddress{
		Name:     sl.String(vipName),
		Services: []datatypes.Network_LoadBalancer_Service{template},
	})
	if err != nil {
		return fmt.Errorf("Error updating LoadBalancer Service: %s", err)
	}

	return nil
}

func resourceSoftLayerNetworkLoadBalancerServiceDelete(d *schema.ResourceData, meta interface{}) error {
	vipName, nadcId, err := parseVipUniqueId(d.Get("vip_id").(string))
	if err != nil {
		return fmt.Errorf("Error parsing vip id: %s", err)
	}

	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)
	serviceName := d.Get("name").(string)

	_, err = service.Id(nadcId).DeleteLiveLoadBalancerService(&datatypes.Network_LoadBalancer_Service{
		Name: sl.String(serviceName),
		Vip: &datatypes.Network_LoadBalancer_VirtualIpAddress{
			Name: sl.String(vipName),
		},
	})

	if err != nil {
		return fmt.Errorf("Error deleting Load Balancer Service %s: %s", serviceName, err)
	}

	return nil
}

func resourceSoftLayerNetworkLoadBalancerServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	vipUniqueId := d.Get("vip_id").(string)
	vipName, nadcId, err := parseVipUniqueId(vipUniqueId)
	if err != nil {
		return false, fmt.Errorf("Error parsing vip id: %s", err)
	}

	serviceName := d.Get("name").(string)
	sess := meta.(*session.Session)
	lbService, err := network.GetNadcLbVipServiceByName(sess, nadcId, vipName, serviceName)
	if err != nil {
		return false, fmt.Errorf("Unable to get load balancer service: %s", err)
	}

	return err == nil && *lbService.Name == serviceName , nil
}
