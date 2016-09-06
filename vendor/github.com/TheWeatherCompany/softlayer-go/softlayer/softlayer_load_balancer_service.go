package softlayer

import (
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
)

type SoftLayer_Load_Balancer_CreateOptions struct {
	Connections int
	Location    string
	HaEnabled   bool
}

type SoftLayer_Load_Balancer_Service_Group_CreateOptions struct {
	Allocation    int
	Port          int
	RoutingMethod string
	RoutingType   string
}

type SoftLayer_Load_Balancer_Service_CreateOptions struct {
	ServiceGroupId  int
	Enabled         int
	Port            int
	IpAddressId     int
	HealthCheckType string
	Weight          int
}

type SoftLayer_Load_Balancer_Service interface {
	Service

	CreateLoadBalancer(createOptions *SoftLayer_Load_Balancer_CreateOptions) (datatypes.SoftLayer_Load_Balancer, error)
	UpdateLoadBalancer(lbId int, lb *datatypes.SoftLayer_Load_Balancer_Update) (bool, error)

	CreateLoadBalancerVirtualServer(lbId int, createOptions *SoftLayer_Load_Balancer_Service_Group_CreateOptions) (bool, error)
	UpdateLoadBalancerVirtualServer(lbId int, sgId int, updateOptions *SoftLayer_Load_Balancer_Service_Group_CreateOptions) (bool, error)

	CreateLoadBalancerService(lbId int, createOptions *SoftLayer_Load_Balancer_Service_CreateOptions) (bool, error)
	UpdateLoadBalancerService(lbId int, sgId int, sId int, updateOptions *SoftLayer_Load_Balancer_Service_CreateOptions) (bool, error)

	GetObject(id int) (datatypes.SoftLayer_Load_Balancer, error)

	DeleteObject(id int) (bool, error)
	DeleteLoadBalancerVirtualServer(id int) (bool, error)
	DeleteLoadBalancerService(id int) (bool, error)

	FindCreatePriceItems(createOptions *SoftLayer_Load_Balancer_CreateOptions) ([]datatypes.SoftLayer_Product_Item_Price, error)
}
