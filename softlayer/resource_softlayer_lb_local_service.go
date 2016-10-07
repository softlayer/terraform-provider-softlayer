package softlayer

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerLbLocalService() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbLocalServiceCreate,
		Read:     resourceSoftLayerLbLocalServiceRead,
		Update:   resourceSoftLayerLbLocalServiceUpdate,
		Delete:   resourceSoftLayerLbLocalServiceDelete,
		Exists:   resourceSoftLayerLbLocalServiceExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"service_group_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ip_address_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"health_check_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerLbLocalServiceCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	// SoftLayer Local LBs consist of a multi-level hierarchy of types.
	// (virtualIpAddress -> []virtualServer -> []serviceGroup -> []service)

	// Using the service group ID provided in the config, find the IDs of the
	// respective virtualServer and virtualIpAddress
	sgID := d.Get("service_group_id").(int)
	serviceGroup, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceGroupService(sess).
		Id(sgID).
		Mask("id,routingMethodId,routingTypeId,virtualServer[id,allocation,port,virtualIpAddress[id]]").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	// Store the IDs for later use
	vsID := *serviceGroup.VirtualServer.Id
	vipID := *serviceGroup.VirtualServer.VirtualIpAddress.Id

	// Convert the health check type name to an ID
	healthCheckTypeId, err := getHealthCheckTypeId(sess, d.Get("health_check_type").(string))
	if err != nil {
		return err
	}

	// The API only exposes edit capability at the root of the tree (virtualIpAddress),
	// so need to send the full structure from the root down to the node to be added or
	// modified
	vip := datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{

		VirtualServers: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualServer{{
			Id:         &vsID,
			Allocation: serviceGroup.VirtualServer.Allocation,
			Port:       serviceGroup.VirtualServer.Port,

			ServiceGroups: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group{{
				Id:              &sgID,
				RoutingMethodId: serviceGroup.RoutingMethodId,
				RoutingTypeId:   serviceGroup.RoutingTypeId,

				Services: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service{{
					Enabled:     sl.Int(1),
					Port:        sl.Int(d.Get("port").(int)),
					IpAddressId: sl.Int(d.Get("ip_address_id").(int)),

					HealthChecks: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{{
						HealthCheckTypeId: &healthCheckTypeId,
					}},

					GroupReferences: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group_CrossReference{{
						Weight: sl.Int(d.Get("weight").(int)),
					}},
				}},
			}},
		}},
	}

	log.Println("[INFO] Creating load balancer service")

	success, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		EditObject(&vip)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service: %s", err)
	}

	if !success {
		return errors.New("Error creating load balancer service")
	}

	// Retrieve the newly created object, to obtain its ID
	svcs, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceGroupService(sess).
		Id(sgID).
		Mask("mask[id,port,ipAddressId]").
		Filter(filter.New(
			filter.Path("services.port").Eq(d.Get("port")),
			filter.Path("services.ipAddressId").Eq(d.Get("ip_address_id"))).Build()).
		GetServices()

	if err != nil || len(svcs) == 0 {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.SetId(strconv.Itoa(*svcs[0].Id))

	log.Printf("[INFO] Load Balancer Service ID: %s", d.Id())

	return resourceSoftLayerLbLocalServiceRead(d, meta)
}

func resourceSoftLayerLbLocalServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	// Using the ID stored in the config, find the IDs of the respective
	// serviceGroup, virtualServer and virtualIpAddress
	svcID, _ := strconv.Atoi(d.Id())
	svc, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceService(sess).
		Id(svcID).
		Mask("id,serviceGroup[id,routingTypeId,routingMethodId,virtualServer[id,allocation,port,virtualIpAddress[id]]]").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	// Store the IDs for later use
	sgID := *svc.ServiceGroup.Id
	vsID := *svc.ServiceGroup.VirtualServer.Id
	vipID := *svc.ServiceGroup.VirtualServer.VirtualIpAddress.Id

	// Convert the health check type name to an ID
	healthCheckTypeId, err := getHealthCheckTypeId(sess, d.Get("health_check_type").(string))
	if err != nil {
		return err
	}

	// The API only exposes edit capability at the root of the tree (virtualIpAddress),
	// so need to send the full structure from the root down to the node to be added or
	// modified
	vip := datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress{

		VirtualServers: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_VirtualServer{{
			Id:         &vsID,
			Allocation: svc.ServiceGroup.VirtualServer.Allocation,
			Port:       svc.ServiceGroup.VirtualServer.Port,

			ServiceGroups: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group{{
				Id:              &sgID,
				RoutingMethodId: svc.ServiceGroup.RoutingMethodId,
				RoutingTypeId:   svc.ServiceGroup.RoutingTypeId,

				Services: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service{{
					Id:          &svcID,
					Enabled:     sl.Int(1),
					Port:        sl.Int(d.Get("port").(int)),
					IpAddressId: sl.Int(d.Get("ip_address_id").(int)),

					HealthChecks: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Health_Check{{
						HealthCheckTypeId: &healthCheckTypeId,
					}},

					GroupReferences: []datatypes.Network_Application_Delivery_Controller_LoadBalancer_Service_Group_CrossReference{{
						Weight: sl.Int(d.Get("weight").(int)),
					}},
				}},
			}},
		}},
	}

	log.Println("[INFO] Updating load balancer service")

	success, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerVirtualIpAddressService(sess).
		Id(vipID).
		EditObject(&vip)

	if err != nil {
		return fmt.Errorf("Error updating load balancer service: %s", err)
	}

	if !success {
		return errors.New("Error updating load balancer service")
	}

	return resourceSoftLayerLbLocalServiceRead(d, meta)
}

func resourceSoftLayerLbLocalServiceRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	svcID, _ := strconv.Atoi(d.Id())

	svc, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceService(sess).
		Id(svcID).
		Mask("ipAddressId,port,healthChecks[type[keyname]],groupReferences[weight]").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving service: %s", err)
	}

	d.Set("ip_address_id", *svc.IpAddressId)
	d.Set("port", *svc.Port)
	d.Set("health_check_type", *svc.HealthChecks[0].Type.Keyname)
	d.Set("weight", *svc.GroupReferences[0].Weight)

	return nil
}

func resourceSoftLayerLbLocalServiceDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)

	svcID, _ := strconv.Atoi(d.Id())

	err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceService(sess).
		Id(svcID).
		DeleteObject()

	if err != nil {
		return fmt.Errorf("Error deleting service: %s", err)
	}

	return nil
}

func resourceSoftLayerLbLocalServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)

	svcID, _ := strconv.Atoi(d.Id())

	_, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerServiceService(sess).
		Id(svcID).
		Mask("id").
		GetObject()

	if err != nil {
		return false, err
	}

	return true, nil
}

func getHealthCheckTypeId(sess *session.Session, healthCheckTypeName string) (int, error) {
	healthCheckTypes, err := services.GetNetworkApplicationDeliveryControllerLoadBalancerHealthCheckTypeService(sess).
		Mask("id").
		Filter(filter.Build(
			filter.Path("keyname").Eq(healthCheckTypeName))).
		Limit(1).
		GetAllObjects()

	if err != nil {
		return -1, err
	}

	if len(healthCheckTypes) < 1 {
		return -1, fmt.Errorf("Invalid health check type: %s", healthCheckTypeName)
	}

	return *healthCheckTypes[0].Id, nil
}
