package softlayer

import (
	"fmt"
	"log"

	"github.com/TheWeatherCompany/softlayer-go/softlayer"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

func resourceSoftLayerLoadBalancerServiceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerLoadBalancerServiceGroupCreate,
		Read:   resourceSoftLayerLoadBalancerServiceGroupRead,
		Update: resourceSoftLayerLoadBalancerServiceGroupUpdate,
		Delete: resourceSoftLayerLoadBalancerServiceGroupDelete,
		Exists: resourceSoftLayerLoadBalancerServiceGroupExists,

		Schema: map[string]*schema.Schema{
			"virtual_server_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_group_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"load_balancer_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"allocation": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"routing_method": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerLoadBalancerServiceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService

	loadBalancer, err := client.GetObject(d.Get("load_balancer_id").(int))

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	opts := softlayer.SoftLayer_Load_Balancer_Service_Group_CreateOptions{
		Allocation:    d.Get("allocation").(int),
		Port:          d.Get("port").(int),
		RoutingMethod: d.Get("routing_method").(string),
		RoutingType:   d.Get("routing_type").(string),
	}

	log.Printf("[INFO] Creating load balancer service group")

	success, err := client.CreateLoadBalancerVirtualServer(loadBalancer.Id, &opts)

	if err != nil {
		return fmt.Errorf("Error creating load balancer service group: %s", err)
	}

	if !success {
		return fmt.Errorf("Error creating load balancer service group")
	}

	loadBalancer, err = client.GetObject(d.Get("load_balancer_id").(int))

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	for _, virtualServer := range loadBalancer.VirtualServers {
		if virtualServer.Port == d.Get("port").(int) {
			d.SetId(strconv.Itoa(virtualServer.Id))
			break
		}
	}

	log.Printf("[INFO] Load Balancer Service Group ID: %s", d.Id())

	return resourceSoftLayerLoadBalancerServiceGroupRead(d, meta)
}

func resourceSoftLayerLoadBalancerServiceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService

	loadBalancer, err := client.GetObject(d.Get("load_balancer_id").(int))

	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	opts := softlayer.SoftLayer_Load_Balancer_Service_Group_CreateOptions{
		Allocation:    d.Get("allocation").(int),
		Port:          d.Get("port").(int),
		RoutingMethod: d.Get("routing_method").(string),
		RoutingType:   d.Get("routing_type").(string),
	}

	log.Printf("[INFO] Updating load balancer service group")

	success, err := client.UpdateLoadBalancerVirtualServer(
		loadBalancer.Id,
		d.Get("service_group_id").(int),
		&opts)

	if err != nil {
		return fmt.Errorf("Error updating load balancer service group: %s", err)
	}

	if !success {
		return fmt.Errorf("Error updating load balancer service group")
	}

	return resourceSoftLayerLoadBalancerServiceGroupRead(d, meta)
}

func resourceSoftLayerLoadBalancerServiceGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService

	id, _ := strconv.Atoi(d.Id())

	loadBalancer, err := client.GetObject(d.Get("load_balancer_id").(int))
	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	for _, virtualServer := range loadBalancer.VirtualServers {
		if virtualServer.Id == id {
			d.Set("virtual_server_id", virtualServer.Id)
			d.Set("allocation", virtualServer.Allocation)
			d.Set("port", virtualServer.Port)

			serviceGroup := virtualServer.ServiceGroups[0]
			d.Set("service_group_id", serviceGroup.Id)
			d.Set("routing_method", serviceGroup.RoutingMethod)
			d.Set("routing_type", serviceGroup.RoutingType)
			break
		}
	}

	return nil
}

func resourceSoftLayerLoadBalancerServiceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService

	success, err := client.DeleteLoadBalancerVirtualServer(d.Get("virtual_server_id").(int))

	if err != nil {
		return fmt.Errorf("Error deleting service group: %s", err)
	}

	if !success {
		return fmt.Errorf("Error deleting service group")
	}

	return nil
}

func resourceSoftLayerLoadBalancerServiceGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client).loadBalancerService

	id, _ := strconv.Atoi(d.Id())

	lb, err := client.GetObject(d.Get("load_balancer_id").(int))
	if err != nil {
		return false, err
	}

	for _, virtualServer := range lb.VirtualServers {
		if virtualServer.Id == id {
			return true, nil
		}
	}

	return false, nil
}
