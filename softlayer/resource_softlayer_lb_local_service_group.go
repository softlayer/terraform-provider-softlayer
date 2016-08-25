package softlayer

import (
	"errors"
	"fmt"
	"log"

	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/filter"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
	"github.ibm.com/riethm/gopherlayer.git/sl"
)

func resourceSoftLayerLbLocalServiceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerLbLocalServiceGroupCreate,
		Read:   resourceSoftLayerLbLocalServiceGroupRead,
		Update: resourceSoftLayerLbLocalServiceGroupUpdate,
		Delete: resourceSoftLayerLbLocalServiceGroupDelete,
		Exists: resourceSoftLayerLbLocalServiceGroupExists,

		Schema: map[string]*schema.Schema{
			"virtual_server_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_group_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"allocation": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"routing_method": {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerLbLocalServiceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipID := d.Get("load_balancer_id").(int)

	routingMethodId, err := getRoutingMethodId(sess, d.Get("routing_method").(string))
	if err != nil {
		return err
	}

	routingTypeId, err := getRoutingTypeId(sess, d.Get("routing_type").(string))
	if err != nil {
		return err
	}

	vip := datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{

		VirtualServers: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualServer{{
			Allocation: sl.Int(d.Get("allocation").(int)),
			Port:       sl.Int(d.Get("port").(int)),

			ServiceGroups: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group{{
				RoutingMethodId: &routingMethodId,
				RoutingTypeId:   &routingTypeId,
			}},
		}},
	}

	log.Println("[INFO] Creating load balancer service group")

	success, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		EditObject(&vip)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service group: %s", err)
	}

	if !success {
		return errors.New("Error creating load balancer service group")
	}

	// Retrieve the newly created object, to obtain its ID
	vs, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		Filter(filter.New(filter.Path("serviceGroups.port").Eq(d.Get("port"))).Build()).
		Mask("id,serviceGroups[id]").
		GetVirtualServers()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.SetId(strconv.Itoa(*vs[0].Id))
	d.Set("service_group_id", *vs[0].ServiceGroups[0].Id)

	log.Printf("[INFO] Load Balancer Service Group ID: %s", d.Id())

	return resourceSoftLayerLbLocalServiceGroupRead(d, meta)
}

func resourceSoftLayerLbLocalServiceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vipID := d.Get("load_balancer_id").(int)
	vsID, _ := strconv.Atoi(d.Id())
	sgID := d.Get("service_group_id").(int)

	routingMethodId, err := getRoutingMethodId(sess, d.Get("routing_method").(string))
	if err != nil {
		return err
	}

	routingTypeId, err := getRoutingTypeId(sess, d.Get("routing_type").(string))
	if err != nil {
		return err
	}

	vip := datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{

		VirtualServers: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualServer{{
			Id:         &vsID,
			Allocation: sl.Int(d.Get("allocation").(int)),
			Port:       sl.Int(d.Get("port").(int)),

			ServiceGroups: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group{{
				Id:              &sgID,
				RoutingMethodId: &routingMethodId,
				RoutingTypeId:   &routingTypeId,
			}},
		}},
	}

	log.Println("[INFO] Updating load balancer service group")

	success, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		EditObject(&vip)

	if err != nil {
		return fmt.Errorf("Error updating load balancer service group: %s", err)
	}

	if !success {
		return errors.New("Error updating load balancer service group")
	}

	return resourceSoftLayerLbLocalServiceGroupRead(d, meta)
}

func resourceSoftLayerLbLocalServiceGroupRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vsID, _ := strconv.Atoi(d.Id())

	vs, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualServerService(sess).
		Id(vsID).
		Mask("allocation,port,serviceGroups[routingMethod[keyname],routingType[keyname]]").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.Set("allocation", *vs.Allocation)
	d.Set("port", *vs.Port)

	d.Set("routing_method", *vs.ServiceGroups[0].RoutingMethod.Keyname)
	d.Set("routing_type", *vs.ServiceGroups[0].RoutingType.Keyname)

	return nil
}

func resourceSoftLayerLbLocalServiceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	vsID, _ := strconv.Atoi(d.Id())

	// There is a bug in the SoftLayer API metadata.  For some services
	// DeleteObject actually returns null on a successful delete, which
	// causes a parse error (since the metadata says that a boolean is
	// returned). Work around this by calling the API method more
	// directly, and avoid return value parsing.
	var pResult *datatypes.Void
	err := sess.DoRequest(
		"SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualServer",
		"deleteObject",
		nil,
		&sl.Options{Id: &vsID},
		pResult)

	if err != nil {
		return fmt.Errorf("Error deleting service: %s", err)
	}

	return nil
}

func resourceSoftLayerLbLocalServiceGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)

	vsID, _ := strconv.Atoi(d.Id())

	_, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualServerService(sess).
		Id(vsID).
		Mask("id").
		GetObject()

	if err != nil {
		return false, err
	}

	return true, nil
}

func getRoutingTypeId(sess *session.Session, routingTypeName string) (int, error) {
	routingTypes, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerRoutingTypeService(sess).
		Mask("id").
		Filter(filter.Build(
			filter.Path("keyname").Eq(routingTypeName))).
		Limit(1).
		GetAllObjects()

	if err != nil {
		return -1, err
	}

	if len(routingTypes) < 1 {
		return -1, fmt.Errorf("Invalid routing type: %s", routingTypeName)
	}

	return *routingTypes[0].Id, nil
}

func getRoutingMethodId(sess *session.Session, routingMethodName string) (int, error) {
	routingMethods, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerRoutingMethodService(sess).
		Mask("id").
		Filter(filter.Build(
			filter.Path("keyname").Eq(routingMethodName))).
		Limit(1).
		GetAllObjects()

	if err != nil {
		return -1, err
	}

	if len(routingMethods) < 1 {
		return -1, fmt.Errorf("Invalid routing method: %s", routingMethodName)
	}

	return *routingMethods[0].Id, nil
}
