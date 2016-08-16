package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
	"strconv"
)

const (
	DATACENTER_TYPE_NAME         = "SoftLayer_Location_Datacenter"
	LOCATION_GROUP_REGIONAL_NAME = "SoftLayer_Location_Group_Regional"
	ROUTING_TYPE_NAME            = "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Routing_Type"
	ROUTING_METHOD_NAME          = "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Routing_Method"
	HEALTH_CHECK_TYPE_NAME       = "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Health_Check_Type"

	DATACENTER_VALUE_NAME              = "name"
	LOCATION_GROUP_REGIONAL_VALUE_NAME = "name"
	ROUTING_TYPE_VALUE_NAME            = "keyname"
	ROUTING_METHOD_VALUE_NAME          = "keyname"
	HEALTH_CHECK_TYPE_VALUE_NAME       = "keyname"

	DATACENTER_GET_JSON_NAME              = "getDatacenters.json"
	LOCATION_GROUP_REGIONAL_GET_JSON_NAME = "getAllObjects.json"
	ROUTING_TYPE_GET_JSON_NAME            = "getAllObjects.json"
	ROUTING_METHOD_GET_JSON_NAME          = "getAllObjects.json"
	HEALTH_CHECK_TYPE_GET_JSON_NAME       = "getAllObjects.json"
)

type lookup func([]byte) (interface{}, error)

func isInt(key interface{}) (bool, error) {
	switch v := key.(type) {
	case int:
		return true, nil
	case string:
		return false, nil
	default:
		return false, fmt.Errorf("Expected type string or int. Recieved type %s", v)
	}
}

func getValueFromKey(client softlayer.Client, nameMask string, nameType string, nameTypeGet string, key interface{}, getById bool, lookupFunc lookup) (interface{}, error) {
	var ObjectFilter, Url string
	var response []byte
	var errorCode int
	var err error

	ObjectMasks := []string{"id", nameMask}

	if getById {
		Url = fmt.Sprintf("%s/%d/getObject.json", nameType, key.(int))
		ObjectFilter = string(`{"id":{"operation":"` + strconv.Itoa(key.(int)) + `"}}`)
		response, errorCode, err = client.GetHttpClient().DoRawHttpRequestWithObjectMask(Url, ObjectMasks, "GET", new(bytes.Buffer))
	} else {
		Url = fmt.Sprintf("%s/%s", nameType, nameTypeGet)
		ObjectFilter = string(`{"` + nameMask + `":{"operation":"` + key.(string) + `"}}`)
		response, errorCode, err = client.GetHttpClient().DoRawHttpRequestWithObjectFilterAndObjectMask(Url, ObjectMasks, ObjectFilter, "GET", new(bytes.Buffer))
	}

	if err != nil {
		return -1, err
	}

	if IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not retrieve %s, HTTP error code: '%d'", nameType, errorCode)
		return -1, errors.New(errorMessage)
	}

	return lookupFunc(response)
}

func GetDatacenter(client softlayer.Client, key interface{}) (interface{}, error) {
	getById, err := isInt(key)

	if err != nil {
		return -1, err
	}

	return getValueFromKey(client, DATACENTER_VALUE_NAME, DATACENTER_TYPE_NAME, DATACENTER_GET_JSON_NAME, key, getById,
		func(response []byte) (interface{}, error) {
			if getById {
				location := datatypes.SoftLayer_Location{}

				err := json.Unmarshal(response, &location)

				if err != nil {
					return -1, err
				}

				return location.Name, nil
			} else {
				locations := []datatypes.SoftLayer_Location{}

				err := json.Unmarshal(response, &locations)

				if err != nil {
					return -1, err
				}

				for _, location := range locations {
					if location.Name == key.(string) {
						return location.Id, nil
					}
				}

				return -1, fmt.Errorf("Datacenter %s not found", key)
			}
		})
}

func GetRoutingType(client softlayer.Client, key interface{}) (interface{}, error) {
	getById, err := isInt(key)

	if err != nil {
		return -1, err
	}

	return getValueFromKey(client, ROUTING_TYPE_VALUE_NAME, ROUTING_TYPE_NAME, ROUTING_TYPE_GET_JSON_NAME, key, getById,
		func(response []byte) (interface{}, error) {
			if getById {
				routingType := datatypes.SoftLayer_Routing_Type{}

				err := json.Unmarshal(response, &routingType)

				if err != nil {
					return -1, err
				}

				return routingType.KeyName, nil
			} else {
				routingTypes := []datatypes.SoftLayer_Routing_Type{}

				err := json.Unmarshal(response, &routingTypes)

				if err != nil {
					return -1, err
				}

				for _, routingType := range routingTypes {
					if routingType.KeyName == key.(string) {
						return routingType.Id, nil
					}
				}

				return -1, fmt.Errorf("Routing type %s not found", key)
			}
		})
}

func GetRoutingMethod(client softlayer.Client, key interface{}) (interface{}, error) {
	getById, err := isInt(key)

	if err != nil {
		return -1, err
	}

	return getValueFromKey(client, ROUTING_METHOD_VALUE_NAME, ROUTING_METHOD_NAME, ROUTING_METHOD_GET_JSON_NAME, key, getById,
		func(response []byte) (interface{}, error) {
			if getById {
				routingMethod := datatypes.SoftLayer_Routing_Method{}

				err := json.Unmarshal(response, &routingMethod)

				if err != nil {
					return -1, err
				}

				return routingMethod.KeyName, nil
			} else {
				routingMethods := []datatypes.SoftLayer_Routing_Method{}

				err := json.Unmarshal(response, &routingMethods)

				if err != nil {
					return -1, err
				}

				for _, routingMethod := range routingMethods {
					if routingMethod.KeyName == key.(string) {
						return routingMethod.Id, nil
					}
				}

				return -1, fmt.Errorf("Routing method %s not found", key)
			}
		})
}

func GetHealthCheckType(client softlayer.Client, key interface{}) (interface{}, error) {
	getById, err := isInt(key)

	if err != nil {
		return -1, err
	}

	return getValueFromKey(client, HEALTH_CHECK_TYPE_VALUE_NAME, HEALTH_CHECK_TYPE_NAME, HEALTH_CHECK_TYPE_GET_JSON_NAME, key, getById,
		func(response []byte) (interface{}, error) {
			if getById {
				healthCheckType := datatypes.SoftLayer_Health_Check_Type{}

				err := json.Unmarshal(response, &healthCheckType)

				if err != nil {
					return -1, err
				}

				return healthCheckType.KeyName, nil
			} else {
				healthCheckTypes := []datatypes.SoftLayer_Health_Check_Type{}

				err := json.Unmarshal(response, &healthCheckTypes)

				if err != nil {
					return -1, err
				}

				for _, healthCheckType := range healthCheckTypes {
					if healthCheckType.KeyName == key.(string) {
						return healthCheckType.Id, nil
					}
				}

				return -1, fmt.Errorf("Health check type %s not found", key)
			}
		})
}

func GetLocationGroupRegional(client softlayer.Client, key interface{}) (interface{}, error) {
	getById, err := isInt(key)

	if err != nil {
		return -1, err
	}

	return getValueFromKey(client, LOCATION_GROUP_REGIONAL_VALUE_NAME, LOCATION_GROUP_REGIONAL_NAME, LOCATION_GROUP_REGIONAL_GET_JSON_NAME, key, getById,
		func(response []byte) (interface{}, error) {
			if getById {
				locationGroupRegional := datatypes.SoftLayer_Location_Group_Regional{}

				err := json.Unmarshal(response, &locationGroupRegional)

				if err != nil {
					return -1, err
				}

				return locationGroupRegional.Name, nil
			} else {
				locationGroupRegionals := []datatypes.SoftLayer_Location_Group_Regional{}

				err := json.Unmarshal(response, &locationGroupRegionals)

				if err != nil {
					return -1, err
				}

				for _, locationGroupRegional := range locationGroupRegionals {
					if locationGroupRegional.Name == key.(string) {
						return locationGroupRegional.Id, nil
					}
				}

				return -1, fmt.Errorf("Location group regional %s not found", key)
			}
		})
}
