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
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerLbLocalServiceGroup() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbLocalServiceGroupCreate,
		Read:     resourceSoftLayerLbLocalServiceGroupRead,
		Update:   resourceSoftLayerLbLocalServiceGroupUpdate,
		Delete:   resourceSoftLayerLbLocalServiceGroupDelete,
		Exists:   resourceSoftLayerLbLocalServiceGroupExists,
		Importer: &schema.ResourceImporter{},

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
	sess := meta.(ProviderConfig).SoftLayerSession()

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

	err = updateLoadBalancerService(sess, vipID, &vip)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service group: %s", err)
	}

	// Retrieve the newly created object, to obtain its ID
	vs, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		Filter(filter.New(filter.Path("virtualServers.port").Eq(d.Get("port"))).Build()).
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
	sess := meta.(ProviderConfig).SoftLayerSession()

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

	err = updateLoadBalancerService(sess, vipID, &vip)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service group: %s", err)
	}

	return resourceSoftLayerLbLocalServiceGroupRead(d, meta)
}

func resourceSoftLayerLbLocalServiceGroupRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

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
	sess := meta.(ProviderConfig).SoftLayerSession()

	vsID, _ := strconv.Atoi(d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualServerService(sess).
				Id(vsID).
				DeleteObject()

			if apiErr, ok := err.(sl.Error); ok {
				switch {
				case apiErr.Exception == "SoftLayer_Exception_Network_Timeout" ||
					strings.Contains(apiErr.Message, "There was a problem saving your configuration to the load balancer.") ||
					strings.Contains(apiErr.Message, "The selected group could not be removed from the load balancer.") ||
					strings.Contains(apiErr.Message, "An error has occurred while processing your request.") ||
					strings.Contains(apiErr.Message, "The resource '480' is already in use."):
					// The LB is busy with another transaction. Retry
					return false, "pending", nil
				case apiErr.StatusCode == 404:
					// 404 - service was deleted on the previous attempt
					return true, "complete", nil
				default:
					// Any other error is unexpected. Abort
					return false, "", err
				}
			}

			return true, "complete", nil
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("Error deleting service: %s", err)
	}

	return nil
}

func resourceSoftLayerLbLocalServiceGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

	vsID, _ := strconv.Atoi(d.Id())

	_, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualServerService(sess).
		Id(vsID).
		Mask("id").
		GetObject()

	if err != nil {
		if apiErr, ok := err.(sl.Error); ok {
			if apiErr.StatusCode == 404 {
				return false, nil
			}
		}

		return false, fmt.Errorf("Error retrieving local lb service group: %s", err)
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
