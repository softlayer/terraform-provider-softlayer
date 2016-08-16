package services

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type softlayer_Scale_Policy_Trigger_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Scale_Policy_Trigger_Service(client softlayer.Client) *softlayer_Scale_Policy_Trigger_Service {
	return &softlayer_Scale_Policy_Trigger_Service{
		client: client,
	}
}

func (slsps *softlayer_Scale_Policy_Trigger_Service) GetName() string {
	return "SoftLayer_Scale_Policy_Trigger"
}

func (slsgs *softlayer_Scale_Policy_Trigger_Service) DeleteObject(trigger int) (bool, error) {
	response, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/deleteObject", slsgs.GetName(), trigger), "GET", new(bytes.Buffer))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, fmt.Errorf("Failed to force delete scale policy trigger with id '%d', got '%s' as response from the API.", trigger, res)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Policy_Trigger#deleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}
	return true, err
}
