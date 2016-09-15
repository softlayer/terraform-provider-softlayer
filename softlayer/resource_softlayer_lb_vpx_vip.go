package softlayer

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/helpers/network"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"strconv"
	"strings"
	"time"
)

func resourceSoftLayerLbVpxVip() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbVpxVipCreate,
		Read:     resourceSoftLayerLbVpxVipRead,
		Update:   resourceSoftLayerLbVpxVipUpdate,
		Delete:   resourceSoftLayerLbVpxVipDelete,
		Exists:   resourceSoftLayerLbVpxVipExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"nad_controller_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"load_balancing_method": {
				Type:     schema.TypeString,
				Required: true,
			},

			// name field is actually used as an ID in SoftLayer
			// http://sldn.softlayer.com/reference/services/SoftLayer_Network_Application_Delivery_Controller/updateLiveLoadBalancer
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"source_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"virtual_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func parseId(id string) (int, string, error) {
	if len(id) < 1 {
		return 0, "", fmt.Errorf("Failed to parse id : Unable to get a VIP ID")
	}

	idList := strings.Split(id, ":")
	if len(idList) != 2 || len(idList[0]) < 1 || len(idList[1]) < 1 {
		return 0, "", fmt.Errorf("Failed to parse id : Invalid VIP ID")
	}

	nadcId, err := strconv.Atoi(idList[0])
	if err != nil {
		return 0, "", fmt.Errorf("Failed to parse id : Unable to get a VIP ID %s", err)
	}

	vipName := idList[1]
	return nadcId, vipName, nil
}

func resourceSoftLayerLbVpxVipCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	vipName := d.Get("name").(string)

	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		LoadBalancingMethod: sl.String(d.Get("load_balancing_method").(string)),
		Name:                sl.String(vipName),
		SourcePort:          sl.Int(d.Get("source_port").(int)),
		Type:                sl.String(d.Get("type").(string)),
		VirtualIpAddress:    sl.String(d.Get("virtual_ip_address").(string)),
	}

	log.Printf("[INFO] Creating Virtual Ip Address %s", *template.VirtualIpAddress)

	var err error
	var successFlag bool

	for count := 0; count < 10; count++ {
		successFlag, err = service.Id(nadcId).CreateLiveLoadBalancer(&template)
		log.Printf("[INFO] Creating Virtual Ip Address %s successFlag : %t", *template.VirtualIpAddress, successFlag)

		if err != nil && strings.Contains(err.Error(), "already exists") {
			log.Printf("[INFO] Creating Virtual Ip Address %s error : %s. Ingore the error.", *template.VirtualIpAddress, err.Error())
			successFlag = true
			err = nil
			break
		}

		if err != nil && strings.Contains(err.Error(), "Operation already in progress") {
			log.Printf("[INFO] Creating Virtual Ip Address %s error : %s. Retry in 10 secs", *template.VirtualIpAddress, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error creating Virtual Ip Address: %s", err)
	}

	if !successFlag {
		return errors.New("Error creating Virtual Ip Address")
	}

	d.SetId(fmt.Sprintf("%d:%s", nadcId, vipName))

	log.Printf("[INFO] Netscaler VPX VIP ID: %s", d.Id())

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	vip, err := network.GetNadcLbVipByName(sess, nadcId, vipName)
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : while looking up a virtual ip address : %s", err)
	}

	d.Set("nad_controller_id", nadcId)
	if vip.LoadBalancingMethod != nil {
		d.Set("load_balancing_method", *vip.LoadBalancingMethod)
	}

	if vip.Name != nil {
		d.Set("name", *vip.Name)
	}

	if vip.SourcePort != nil {
		d.Set("source_port", *vip.SourcePort)
	}

	if vip.Type != nil {
		d.Set("type", *vip.Type)
	}

	if vip.VirtualIpAddress != nil {
		d.Set("virtual_ip_address", *vip.VirtualIpAddress)
	}

	return nil
}

func resourceSoftLayerLbVpxVipUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		Name: sl.String(d.Get("name").(string)),
	}

	if d.HasChange("load_balancing_method") {
		template.LoadBalancingMethod = sl.String(d.Get("load_balancing_method").(string))
	}

	if d.HasChange("virtual_ip_address") {
		template.VirtualIpAddress = sl.String(d.Get("virtual_ip_address").(string))
	}

	var err error

	for count := 0; count < 10; count++ {
		var successFlag bool
		successFlag, err = service.Id(nadcId).UpdateLiveLoadBalancer(&template)
		log.Printf("[INFO]  Updating Virtual Ip Address %s successFlag : %t", *template.VirtualIpAddress, successFlag)

		if err != nil && strings.Contains(err.Error(), "Operation already in progress") {
			log.Printf("[INFO] Updating Virtual Ip Address %s error : %s. Retry in 10 secs", *template.VirtualIpAddress, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error updating Virtual Ip Address: %s", err)
	}

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	for count := 0; count < 10; count++ {
		var successFlag bool
		successFlag, err = service.Id(nadcId).DeleteLiveLoadBalancer(
			&datatypes.Network_LoadBalancer_VirtualIpAddress{Name: sl.String(vipName)},
		)
		log.Printf("[INFO] Deleting Virtual Ip Address %s successFlag : %t", vipName, successFlag)

		if err != nil &&
			(strings.Contains(err.Error(), "Operation already in progress") ||
				strings.Contains(err.Error(), "No Service")) {
			log.Printf("[INFO] Deleting Virtual Ip Address %s Error : %s  Retry in 10 secs", vipName, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		// Check if the resource is already deleted.
		if err != nil && strings.Contains(err.Error(), "Unable to find object with unknown identifier of") {
			log.Printf("[INFO] Deleting Virtual Ip Address %s Error : %s . Ignore the error.", vipName, err.Error())
			err = nil
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address %s: %s", vipName, err)
	}

	return nil
}

func resourceSoftLayerLbVpxVipExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return false, fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	vip, err := network.GetNadcLbVipByName(sess, nadcId, vipName)

	return vip != nil && err == nil && *vip.Name == vipName, nil
}
