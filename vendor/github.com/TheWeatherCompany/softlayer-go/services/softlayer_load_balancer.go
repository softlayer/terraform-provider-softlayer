package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	softlayer "github.com/TheWeatherCompany/softlayer-go/softlayer"
	"github.com/hashicorp/terraform/helper/resource"
	"strconv"
	"strings"
	"time"
)

const (
	PACKAGE_TYPE_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER = "ADDITIONAL_SERVICES_LOAD_BALANCER"
	ORDER_TYPE_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER   = "SoftLayer_Container_Product_Order_Network_LoadBalancer"
	PACKAGE_ID_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER   = 194
	LOAD_BALANCER_VIRTUAL_SERVER_NAME                          = "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualServer"
	LOAD_BALANCER_SERVICE_NAME                                 = "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service"
	BILLING_ITEM_TYPE_NAME                                     = "SoftLayer_Billing_Item"
	OBJECT_MASK                                                = "?objectMask=mask[id,connectionLimit,ipAddressId,securityCertificateId,highAvailabilityFlag,sslEnabledFlag,loadBalancerHardware[datacenter[name]],ipAddress[ipAddress,subnet[networkVlan]],virtualServers[serviceGroups[services[healthChecks,groupReferences]]]]"
)

type softLayer_Load_Balancer struct {
	client softlayer.Client
}

func NewSoftLayer_Load_Balancer(client softlayer.Client) *softLayer_Load_Balancer {
	return &softLayer_Load_Balancer{
		client: client,
	}
}

func (slnadcs *softLayer_Load_Balancer) GetName() string {
	return "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress"
}

func (slnadclbs *softLayer_Load_Balancer) CreateLoadBalancer(createOptions *softlayer.SoftLayer_Load_Balancer_CreateOptions) (datatypes.SoftLayer_Load_Balancer, error) {
	orderService, err := slnadclbs.client.GetSoftLayer_Product_Order_Service()
	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	items, err := slnadclbs.FindCreatePriceItems(createOptions)
	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	location, err := common.GetDatacenter(slnadclbs.client, createOptions.Location)

	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	order := datatypes.SoftLayer_Container_Product_Order_Load_Balancer{
		PackageId:   PACKAGE_ID_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER,
		ComplexType: ORDER_TYPE_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER,
		Location:    location.(int),
		Prices:      items,
		Quantity:    1,
	}

	receipt, err := orderService.PlaceContainerOrderLoadBalancer(order)
	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	vpx, err := slnadclbs.findLoadBalancerByOrderId(receipt.OrderId)
	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	return vpx, nil
}

func (slnadclbs *softLayer_Load_Balancer) CreateLoadBalancerVirtualServer(lbId int, createOptions *softlayer.SoftLayer_Load_Balancer_Service_Group_CreateOptions) (bool, error) {
	loadBalancer, err := slnadclbs.GetObject(lbId)

	if err != nil {
		return false, err
	}

	if loadBalancer.Id != lbId {
		return false, fmt.Errorf("Load balancer with id '%d' is not found", lbId)
	}

	routingMethodId, err := common.GetRoutingMethod(slnadclbs.client, createOptions.RoutingMethod)

	if err != nil {
		return false, err
	}

	routingTypeId, err := common.GetRoutingType(slnadclbs.client, createOptions.RoutingType)

	if err != nil {
		return false, err
	}

	parameters := datatypes.SoftLayer_Load_Balancer_Virtual_Server_Update_Parameters{
		Parameters: []datatypes.Softlayer_Load_Balancer_Virtual_Server_Parameters{{
			VirtualServers: []*datatypes.Softlayer_Load_Balancer_Virtual_Server{{
				Allocation: createOptions.Allocation,
				Port:       createOptions.Port,
				ServiceGroups: []*datatypes.Softlayer_Service_Group{{
					RoutingMethodId: routingMethodId.(int),
					RoutingTypeId:   routingTypeId.(int),
				}},
			}},
		}},
	}

	requestBody, err := json.Marshal(parameters)

	response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/%s.json", slnadclbs.GetName(), lbId, "editObject"), "POST", bytes.NewBuffer(requestBody))

	if error != nil {
		return false, error
	} else if errorCode != 200 {
		return false, fmt.Errorf(string(response))
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) UpdateLoadBalancerVirtualServer(lbId int, sgId int, updateOptions *softlayer.SoftLayer_Load_Balancer_Service_Group_CreateOptions) (bool, error) {
	loadBalancer, err := slnadclbs.GetObject(lbId)

	if err != nil {
		return false, err
	}

	if loadBalancer.Id != lbId {
		return false, fmt.Errorf("Load balancer with id '%d' is not found", lbId)
	}

	routingMethodId, err := common.GetRoutingMethod(slnadclbs.client, updateOptions.RoutingMethod)

	if err != nil {
		return false, err
	}

	routingTypeId, err := common.GetRoutingType(slnadclbs.client, updateOptions.RoutingType)

	if err != nil {
		return false, err
	}

	virtualServer, err := getVsFromSgId(sgId, loadBalancer.VirtualServers)

	if err != nil {
		return false, err
	}

	parameters := datatypes.SoftLayer_Load_Balancer_Virtual_Server_Update_Parameters{
		Parameters: []datatypes.Softlayer_Load_Balancer_Virtual_Server_Parameters{{
			VirtualServers: []*datatypes.Softlayer_Load_Balancer_Virtual_Server{{
				Id:         virtualServer.Id,
				Allocation: updateOptions.Allocation,
				Port:       updateOptions.Port,
				ServiceGroups: []*datatypes.Softlayer_Service_Group{{
					Id:              sgId,
					RoutingMethodId: routingMethodId.(int),
					RoutingTypeId:   routingTypeId.(int),
					Services:        []*datatypes.Softlayer_Service{},
				}},
			}},
		}},
	}

	requestBody, err := json.Marshal(parameters)

	response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/%s.json", slnadclbs.GetName(), lbId, "editObject"), "POST", bytes.NewBuffer(requestBody))

	if error != nil {
		return false, error
	} else if errorCode != 200 {
		return false, fmt.Errorf(string(response))
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) CreateLoadBalancerService(lbId int, createOptions *softlayer.SoftLayer_Load_Balancer_Service_CreateOptions) (bool, error) {
	loadBalancer, err := slnadclbs.GetObject(lbId)

	if err != nil {
		return false, err
	}

	if loadBalancer.Id != lbId {
		return false, fmt.Errorf("Load balancer with id '%d' is not found", lbId)
	}

	virtualServer := new(datatypes.Softlayer_Load_Balancer_Virtual_Server)

	for _, vs := range loadBalancer.VirtualServers {
		if vs.ServiceGroups[0].Id == createOptions.ServiceGroupId {
			virtualServer = vs
		}
	}

	if len(virtualServer.ServiceGroups) != 1 {
		return false, fmt.Errorf("Service group with id '%d' is not found", createOptions.ServiceGroupId)
	}

	healthCheckType, err := common.GetHealthCheckType(slnadclbs.client, createOptions.HealthCheckType)

	if err != nil {
		return false, err
	}

	parameters := datatypes.SoftLayer_Load_Balancer_Virtual_Server_Update_Parameters{
		Parameters: []datatypes.Softlayer_Load_Balancer_Virtual_Server_Parameters{{
			VirtualServers: []*datatypes.Softlayer_Load_Balancer_Virtual_Server{{
				Id:         virtualServer.Id,
				Allocation: virtualServer.Allocation,
				Port:       virtualServer.Port,
				ServiceGroups: []*datatypes.Softlayer_Service_Group{{
					Id:              createOptions.ServiceGroupId,
					RoutingMethodId: virtualServer.ServiceGroups[0].RoutingMethodId,
					RoutingTypeId:   virtualServer.ServiceGroups[0].RoutingTypeId,
					Services: []*datatypes.Softlayer_Service{{
						Enabled:     createOptions.Enabled,
						Port:        createOptions.Port,
						IpAddressId: createOptions.IpAddressId,
						HealthChecks: []*datatypes.Softlayer_Health_Check{{
							HealthCheckTypeId: healthCheckType.(int),
						}},
						GroupReferences: []*datatypes.Softlayer_Group_Reference{{
							Weight: createOptions.Weight,
						}},
					}},
				}},
			}},
		}},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, fmt.Errorf("Load balancer with id '%d' is not found: %s", lbId, err)
	}

	response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/%s.json", slnadclbs.GetName(), lbId, "editObject"), "POST", bytes.NewBuffer(requestBody))

	if error != nil {
		return false, error
	} else if errorCode != 200 {
		return false, fmt.Errorf(string(response))
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) UpdateLoadBalancerService(lbId int, sgId int, sId int, updateOptions *softlayer.SoftLayer_Load_Balancer_Service_CreateOptions) (bool, error) {
	loadBalancer, err := slnadclbs.GetObject(lbId)

	if err != nil {
		return false, err
	}

	if loadBalancer.Id != lbId {
		return false, fmt.Errorf("Load balancer with id '%d' is not found", lbId)
	}

	healthCheckTypeId, err := common.GetHealthCheckType(slnadclbs.client, updateOptions.HealthCheckType)

	if err != nil {
		return false, err
	}

	virtualServer, err := getVsFromSgId(sgId, loadBalancer.VirtualServers)

	if err != nil {
		return false, err
	}

	parameters := datatypes.SoftLayer_Load_Balancer_Virtual_Server_Update_Parameters{
		Parameters: []datatypes.Softlayer_Load_Balancer_Virtual_Server_Parameters{{
			VirtualServers: []*datatypes.Softlayer_Load_Balancer_Virtual_Server{{
				Id:         virtualServer.Id,
				Port:       virtualServer.Port,
				Allocation: virtualServer.Allocation,
				ServiceGroups: []*datatypes.Softlayer_Service_Group{{
					Id:              sgId,
					RoutingMethodId: virtualServer.ServiceGroups[0].RoutingMethodId,
					RoutingTypeId:   virtualServer.ServiceGroups[0].RoutingTypeId,
					Services: []*datatypes.Softlayer_Service{{
						Id:          sId,
						Enabled:     1,
						Port:        updateOptions.Port,
						IpAddressId: updateOptions.IpAddressId,
						HealthChecks: []*datatypes.Softlayer_Health_Check{{
							HealthCheckTypeId: healthCheckTypeId.(int),
						}},
						GroupReferences: []*datatypes.Softlayer_Group_Reference{{
							Weight: updateOptions.Weight,
						}},
					}},
				}},
			}},
		}},
	}

	requestBody, err := json.Marshal(parameters)

	response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/%s.json", slnadclbs.GetName(), lbId, "editObject"), "POST", bytes.NewBuffer(requestBody))

	if error != nil {
		return false, error
	} else if errorCode != 200 {
		return false, fmt.Errorf(string(response))
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) UpdateLoadBalancer(lbId int, lb *datatypes.SoftLayer_Load_Balancer_Update) (bool, error) {
	object, err := slnadclbs.GetObject(lbId)
	if err != nil {
		return false, err
	}
	if object.Id != lbId {
		return false, fmt.Errorf("Load balancer with id '%d' is not found", lbId)
	}

	parameters := datatypes.SoftLayer_Load_Balancer_Update_Parameters{
		Parameters: []datatypes.SoftLayer_Load_Balancer_Update{{
			SecurityCertificateId: lb.SecurityCertificateId,
		}},
	}

	if *lb.SecurityCertificateId == 0 {
		parameters = datatypes.SoftLayer_Load_Balancer_Update_Parameters{
			Parameters: []datatypes.SoftLayer_Load_Balancer_Update{{
				SecurityCertificateId: nil,
			}},
		}
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, fmt.Errorf("Load balancer with id '%d' is not found: %s", lbId, err)
	}

	response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/%s.json", slnadclbs.GetName(), lbId, "editObject"), "POST", bytes.NewBuffer(requestBody))

	if error != nil {
		return false, error
	} else if errorCode != 200 {
		return false, fmt.Errorf(string(response))
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) GetObject(id int) (datatypes.SoftLayer_Load_Balancer, error) {
	response, errorCode, err := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/getObject.json%s", slnadclbs.GetName(), id, OBJECT_MASK), "GET", new(bytes.Buffer))

	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Load_Balancer#getObject, error message '%s'", err.Error())
		return datatypes.SoftLayer_Load_Balancer{}, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Load_Balancer#getObject, HTTP error code: '%d'", errorCode)
		return datatypes.SoftLayer_Load_Balancer{}, errors.New(errorMessage)
	}

	lb := datatypes.SoftLayer_Load_Balancer{}
	err = json.Unmarshal(response, &lb)
	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	for _, virtualServer := range lb.VirtualServers {
		for _, serviceGroup := range virtualServer.ServiceGroups {
			routingType, err := common.GetRoutingType(slnadclbs.client, serviceGroup.RoutingTypeId)

			if err != nil {
				return datatypes.SoftLayer_Load_Balancer{}, err
			}

			routingMethod, err := common.GetRoutingMethod(slnadclbs.client, serviceGroup.RoutingMethodId)

			if err != nil {
				return datatypes.SoftLayer_Load_Balancer{}, err
			}

			serviceGroup.RoutingType = routingType.(string)
			serviceGroup.RoutingMethod = routingMethod.(string)

			for _, service := range serviceGroup.Services {
				healthCheckType, err := common.GetHealthCheckType(slnadclbs.client, service.HealthChecks[0].HealthCheckTypeId)

				if err != nil {
					return datatypes.SoftLayer_Load_Balancer{}, err
				}

				service.HealthChecks[0].HealthCheckType = healthCheckType.(string)
			}
		}
	}

	return lb, nil
}

func (slnadclbs *softLayer_Load_Balancer) FindCreatePriceItems(createOptions *softlayer.SoftLayer_Load_Balancer_CreateOptions) ([]datatypes.SoftLayer_Product_Item_Price, error) {
	items, err := slnadclbs.getApplicationDeliveryControllerLoadBalancerItems()
	if err != nil {
		return []datatypes.SoftLayer_Product_Item_Price{}, err
	}

	lbKey := slnadclbs.getLoadBalancerPriceItemKeyName(createOptions.Connections, createOptions.HaEnabled)

	var lbItemPrice datatypes.SoftLayer_Product_Item_Price

	for _, item := range items {
		itemKey := item.Key
		if itemKey == lbKey {
			lbItemPrice = item.Prices[0]
		}
	}

	var errorMessages []string

	if lbItemPrice.Id == 0 {
		errorMessages = append(errorMessages, fmt.Sprintf("LB Connections field has an incorrect value"))
	}

	if len(errorMessages) > 0 {
		err = errors.New(strings.Join(errorMessages, "\n"))
		return []datatypes.SoftLayer_Product_Item_Price{}, err
	}

	return []datatypes.SoftLayer_Product_Item_Price{lbItemPrice}, nil
}

func (slnadclbs *softLayer_Load_Balancer) DeleteObject(id int) (bool, error) {
	billingItem, err := slnadclbs.GetBillingItem(id)
	if err != nil {
		return false, err
	}

	if billingItem.Id > 0 {
		deleted, err := slnadclbs.CancelService(billingItem.Id)
		if err != nil {
			return false, err
		}

		if deleted {
			return false, nil
		}
	}

	return true, fmt.Errorf("softlayer-go: could not SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualIpAddress#deleteLoadBalancer with id: '%d'", id)
}

func (slnadclbs *softLayer_Load_Balancer) DeleteLoadBalancerVirtualServer(id int) (bool, error) {
	_, errorCode, err := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/deleteObject.json", LOAD_BALANCER_VIRTUAL_SERVER_NAME, id), "GET", new(bytes.Buffer))

	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualServer#deleteObject, error message '%s'", err.Error())
		return false, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_VirtualServer#deleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) DeleteLoadBalancerService(id int) (bool, error) {
	_, errorCode, err := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/deleteObject.json", LOAD_BALANCER_SERVICE_NAME, id), "GET", new(bytes.Buffer))

	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service#deleteObject, error message '%s'", err.Error())
		return false, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not perform SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service#deleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) CancelService(billingId int) (bool, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fmt.Sprintf("{\"error\":\"This cancellation could not be processed please contact support.This cancellation could not be processed please contact support. Failed to cancel billing items. Failed to cancel billing item #%d. Error: There is currently an active transaction.\",\"code\":\"SoftLayer_Exception_Public\"}", billingId)},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			response, errorCode, error := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/cancelService.json", BILLING_ITEM_TYPE_NAME, billingId), "GET", new(bytes.Buffer))

			if error != nil {
				return false, "", error
			} else if errorCode == 500 {
				return nil, string(response), nil
			} else {
				return true, "complete", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return false, err
	}

	if !bool(pendingResult.(bool)) {
		return false, nil
	}

	return true, nil
}

func (slnadclbs *softLayer_Load_Balancer) findLoadBalancerByOrderId(orderId int) (datatypes.SoftLayer_Load_Balancer, error) {
	ObjectFilter := string(`{"adcLoadBalancers":{"dedicatedBillingItem":{"orderItem":{"order":{"id":{"operation":` + strconv.Itoa(orderId) + `}}}}}}`)

	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			accountService, err := slnadclbs.client.GetSoftLayer_Account_Service()
			if err != nil {
				return datatypes.SoftLayer_Load_Balancer{}, "", err
			}
			lbs, err := accountService.GetApplicationDeliveryControllerLoadBalancersWithFilterAndMask(ObjectFilter, OBJECT_MASK)
			if err != nil {
				return datatypes.SoftLayer_Load_Balancer{}, "", err
			}

			if len(lbs) == 1 {
				return lbs[0], "complete", nil
			} else if len(lbs) == 0 {
				return nil, "pending", nil
			} else {
				return nil, "", fmt.Errorf("Expected one load balancer: %s", err)
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	pendingResult, err := stateConf.WaitForState()

	if err != nil {
		return datatypes.SoftLayer_Load_Balancer{}, err
	}

	var result, ok = pendingResult.(datatypes.SoftLayer_Load_Balancer)

	if ok {
		return result, nil
	}

	return datatypes.SoftLayer_Load_Balancer{},
		fmt.Errorf("Cannot find Application Delivery Controller Load Balancer with order id '%d'", orderId)
}

func (slnadclbs *softLayer_Load_Balancer) getApplicationDeliveryControllerLoadBalancerItems() ([]datatypes.SoftLayer_Product_Item, error) {
	productPackageService, err := slnadclbs.client.GetSoftLayer_Product_Package_Service()
	if err != nil {
		return []datatypes.SoftLayer_Product_Item{}, err
	}

	return productPackageService.GetItemsByType(PACKAGE_TYPE_APPLICATION_DELIVERY_CONTROLLER_LOAD_BALANCER)
}

func (slnadclbs *softLayer_Load_Balancer) getLoadBalancerPriceItemKeyName(connections int, haEnabled bool) string {
	name := "DEDICATED_LOAD_BALANCER_WITH_HIGH_AVAILABILITY_AND_SSL"

	if !haEnabled {
		name = "LOAD_BALANCER_DEDICATED_WITH_SSL_OFFLOAD"
	}

	return strings.Join([]string{name, strconv.Itoa(connections), "CONNECTIONS"}, DELIMITER)
}

func (slnadclbs *softLayer_Load_Balancer) GetBillingItem(id int) (datatypes.SoftLayer_Billing_Item, error) {

	response, errorCode, err := slnadclbs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/getDedicatedBillingItem.json", slnadclbs.GetName(), id), "GET", new(bytes.Buffer))
	if err != nil {
		return datatypes.SoftLayer_Billing_Item{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not retrieve SoftLayer LoadBalancer Service#getBillingItem, HTTP error code: '%d'", errorCode)
		return datatypes.SoftLayer_Billing_Item{}, errors.New(errorMessage)
	}

	billingItem := datatypes.SoftLayer_Billing_Item{}
	err = json.Unmarshal(response, &billingItem)
	if err != nil {
		return datatypes.SoftLayer_Billing_Item{}, err
	}

	return billingItem, nil
}

func getVsFromSgId(sgId int, virtualServers []*datatypes.Softlayer_Load_Balancer_Virtual_Server) (datatypes.Softlayer_Load_Balancer_Virtual_Server, error) {
	virtualServer := new(datatypes.Softlayer_Load_Balancer_Virtual_Server)

	for _, vs := range virtualServers {
		if vs.ServiceGroups[0].Id == sgId {
			virtualServer = vs
		}
	}

	if virtualServer == nil {
		return datatypes.Softlayer_Load_Balancer_Virtual_Server{}, fmt.Errorf("Virtual server with id '%d' is not found", sgId)
	}

	return *virtualServer, nil
}
