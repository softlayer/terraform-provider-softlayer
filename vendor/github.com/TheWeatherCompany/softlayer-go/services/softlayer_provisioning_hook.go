package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TheWeatherCompany/softlayer-go/common"
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type softLayer_Provisioning_Hook_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Provisioning_Hook_Service(client softlayer.Client) *softLayer_Provisioning_Hook_Service {
	return &softLayer_Provisioning_Hook_Service{
		client: client,
	}
}

func (slphs *softLayer_Provisioning_Hook_Service) GetName() string {
	return "SoftLayer_Provisioning_Hook"
}

func (slphs *softLayer_Provisioning_Hook_Service) CreateObject(template datatypes.SoftLayer_Provisioning_Hook_Template) (datatypes.SoftLayer_Provisioning_Hook, error) {
	template.TypeId = 1

	parameters := datatypes.SoftLayer_Provisioning_Hook_Template_Parameters{
		Parameters: []datatypes.SoftLayer_Provisioning_Hook_Template{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return datatypes.SoftLayer_Provisioning_Hook{}, fmt.Errorf("Unable to create JSON: %s", err)
	}

	response, errorCode, err := slphs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/createObject.json", slphs.GetName()), "POST", bytes.NewBuffer(requestBody))

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not create SoftLayer_Provisioning_Hook\nResponse from SoftLayer: %s\n HTTP error code: '%d", response, errorCode)
		return datatypes.SoftLayer_Provisioning_Hook{}, errors.New(errorMessage)
	}

	err = slphs.client.GetHttpClient().CheckForHttpResponseErrors(response)
	if err != nil {
		return datatypes.SoftLayer_Provisioning_Hook{}, err
	}

	provisioningHook := datatypes.SoftLayer_Provisioning_Hook{}
	err = json.Unmarshal(response, &provisioningHook)
	if err != nil {
		return datatypes.SoftLayer_Provisioning_Hook{}, err
	}

	return provisioningHook, nil
}

func (slphs *softLayer_Provisioning_Hook_Service) GetObject(id int) (datatypes.SoftLayer_Provisioning_Hook, error) {
	response, errorCode, err := slphs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/getObject.json", slphs.GetName(), id), "GET", new(bytes.Buffer))

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not retrieve SoftLayer_Provisioning_Hook, HTTP error code: '%d'", errorCode)
		return datatypes.SoftLayer_Provisioning_Hook{}, errors.New(errorMessage)
	}

	if err != nil {
		return datatypes.SoftLayer_Provisioning_Hook{}, err
	}

	provisioningHook := datatypes.SoftLayer_Provisioning_Hook{}
	err = json.Unmarshal(response, &provisioningHook)
	if err != nil {
		return datatypes.SoftLayer_Provisioning_Hook{}, err
	}

	return provisioningHook, nil
}

func (slphs *softLayer_Provisioning_Hook_Service) EditObject(id int, template datatypes.SoftLayer_Provisioning_Hook_Template) (bool, error) {
	parameters := datatypes.SoftLayer_Provisioning_Hook_Template_Parameters{
		Parameters: []datatypes.SoftLayer_Provisioning_Hook_Template{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, err
	}

	response, errorCode, err := slphs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/editObject.json", slphs.GetName(), id), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to edit Provisioning Hook with id: %d, got '%s' as response from the API.", id, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Provisioning_Hook, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

func (slphs *softLayer_Provisioning_Hook_Service) DeleteObject(id int) (bool, error) {
	response, errorCode, err := slphs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d.json", slphs.GetName(), id), "DELETE", new(bytes.Buffer))

	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to delete Provisioning Hook with id '%d', got '%s' as a response from the SLAPI.", id, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not remove SoftLayer_Provisioning_Hook with Id: %d, HTTP error code: '%d'", id, errorCode)
		return false, errors.New(errorMessage)
	}

	return true, nil
}
