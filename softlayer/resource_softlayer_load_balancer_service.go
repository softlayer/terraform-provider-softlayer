package softlayer

import (
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
)

func resourceSoftLayerLoadBalancerService() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerLoadBalancerServiceCreate,
		Read:   resourceSoftLayerLoadBalancerServiceRead,
		Update: resourceSoftLayerLoadBalancerServiceUpdate,
		Delete: resourceSoftLayerLoadBalancerServiceDelete,
		Exists: resourceSoftLayerLoadBalancerServiceExists,

		Schema: map[string]*schema.Schema{
			"service_group_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ip_address_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"health_check_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"weight": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerLoadBalancerServiceCreate(d *schema.ResourceData, meta interface{}) error {
	serviceGroup, err := getServiceGroup(d.Get("service_group_id").(int), meta)

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	opts := softlayer.SoftLayer_Load_Balancer_Service_CreateOptions{
		ServiceGroupId:  serviceGroup.Id,
		Enabled:         1,
		Port:            d.Get("port").(int),
		IpAddressId:     d.Get("ip_address_id").(int),
		HealthCheckType: d.Get("health_check_type").(string),
		Weight:          d.Get("weight").(int),
	}

	log.Printf("[INFO] Creating load balancer service")

	client := meta.(*Client).loadBalancerService

	success, err := client.CreateLoadBalancerService(serviceGroup.VirtualServer.VirtualIpAddress.Id, &opts)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service: %s", err)
	}

	if !success {
		return fmt.Errorf("Error creating load balancer service")
	}

	lbObj, err := client.GetObject(serviceGroup.VirtualServer.VirtualIpAddress.Id)

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	for _, virtualServer := range lbObj.VirtualServers {
		if virtualServer.Id == virtualServer.Id {
			for _, service := range virtualServer.ServiceGroups[0].Services {
				if service.IpAddressId == d.Get("ip_address_id").(int) &&
					service.Port == d.Get("port").(int) {
					d.SetId(strconv.Itoa(service.Id))
				}
			}
		}
	}

	log.Printf("[INFO] Load Balancer Service ID: %s", d.Id())

	return resourceSoftLayerLoadBalancerServiceRead(d, meta)
}

func resourceSoftLayerLoadBalancerServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	serviceGroup, err := getServiceGroup(d.Get("service_group_id").(int), meta)

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	opts := softlayer.SoftLayer_Load_Balancer_Service_CreateOptions{
		ServiceGroupId:  serviceGroup.Id,
		Enabled:         1,
		Port:            d.Get("port").(int),
		IpAddressId:     d.Get("ip_address_id").(int),
		HealthCheckType: d.Get("health_check_type").(string),
		Weight:          d.Get("weight").(int),
	}

	log.Printf("[INFO] Updating load balancer service")

	client := meta.(*Client).loadBalancerService

	serviceId, _ := strconv.Atoi(d.Id())
	success, err := client.UpdateLoadBalancerService(serviceGroup.VirtualServer.VirtualIpAddress.Id, serviceGroup.Id, serviceId, &opts)

	if err != nil {
		return fmt.Errorf("Error updating load balancer service: %s", err)
	}

	if !success {
		return fmt.Errorf("Error updating load balancer service")
	}

	return resourceSoftLayerLoadBalancerServiceRead(d, meta)
}

func resourceSoftLayerLoadBalancerServiceRead(d *schema.ResourceData, meta interface{}) error {
	serviceGroup, err := getServiceGroup(d.Get("service_group_id").(int), meta)

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	client := meta.(*Client).loadBalancerService

	loadBalancer, err := client.GetObject(serviceGroup.VirtualServer.VirtualIpAddress.Id)
	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	id, _ := strconv.Atoi(d.Id())
	for _, virtualServer := range loadBalancer.VirtualServers {
		serviceGroup := virtualServer.ServiceGroups[0]
		if serviceGroup.Id == serviceGroup.Id {
			for _, service := range serviceGroup.Services {
				if service.Id == id {
					d.Set("ip_address_id", service.IpAddressId)
					d.Set("port", service.Port)
					d.Set("health_check_type", service.HealthChecks[0].HealthCheckType)
					d.Set("weight", service.GroupReferences[0].Weight)
				}
			}
		}
	}

	return nil
}

func resourceSoftLayerLoadBalancerServiceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService

	id, _ := strconv.Atoi(d.Id())
	success, err := client.DeleteLoadBalancerService(id)

	if err != nil {
		return fmt.Errorf("Error deleting service group: %s", err)
	}

	if !success {
		return fmt.Errorf("Error deleting service group")
	}

	return nil
}

func resourceSoftLayerLoadBalancerServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	serviceGroup, err := getServiceGroup(d.Get("service_group_id").(int), meta)

	if err != nil {
		return false, fmt.Errorf("Error retrieving load balancer service group from SoftLayer, %s", err)
	}

	client := meta.(*Client).loadBalancerService

	lb, err := client.GetObject(serviceGroup.VirtualServer.VirtualIpAddress.Id)
	if err != nil {
		return false, err
	}

	id, _ := strconv.Atoi(d.Id())
	for _, virtualServer := range lb.VirtualServers {
		if virtualServer.ServiceGroups[0].Id == serviceGroup.Id {
			for _, service := range virtualServer.ServiceGroups[0].Services {
				if service.Id == id {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func getServiceGroup(sgid int, meta interface{}) (data_types.Softlayer_Service_Group, error) {
	client := meta.(*Client).loadBalancerServiceGroupService

	mask := []string{
		"id",
		"virtualServer.virtualIpAddress.id",
	}

	return client.GetObject(sgid, mask)
}
