package services

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type softlayer_Scale_Network_Vlan_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Scale_Network_Vlan_Service(client softlayer.Client) *softlayer_Scale_Network_Vlan_Service {
	return &softlayer_Scale_Network_Vlan_Service{
		client: client,
	}
}

func (slsnvs *softlayer_Scale_Network_Vlan_Service) GetName() string {
	return "SoftLayer_Scale_Network_Vlan"
}

func (slsnvs *softlayer_Scale_Network_Vlan_Service) DeleteObject(scaleNetworkVlanId int) (bool, error) {
	response, errorCode, err := slsnvs.client.GetHttpClient().DoRawHttpRequest(
		fmt.Sprintf("%s/%d/deleteObject", slsnvs.GetName(), scaleNetworkVlanId), "GET", new(bytes.Buffer))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, fmt.Errorf(
			"Failed to delete scale network vlan with id '%d', got '%s' as response from the API.",
			scaleNetworkVlanId,
			res)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not SoftLayer_Scale_Network_Vlan#deleteObject, HTTP error code: '%d'",
			errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}
