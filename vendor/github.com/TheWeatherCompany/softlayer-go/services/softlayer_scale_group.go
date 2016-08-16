package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	"github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
	"log"
)

type softlayer_Scale_Group_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Scale_Group_Service(client softlayer.Client) *softlayer_Scale_Group_Service {
	return &softlayer_Scale_Group_Service{
		client: client,
	}
}

func (slsgs *softlayer_Scale_Group_Service) GetName() string {
	return "SoftLayer_Scale_Group"
}

func (slsgs *softlayer_Scale_Group_Service) CreateObject(template data_types.SoftLayer_Scale_Group) (data_types.SoftLayer_Scale_Group, error) {

	if template.RegionalGroup != nil && template.RegionalGroup.Name != "" {
		// Replace the regionalGroup sub-structure with the regionalGroupId from a lookup
		// This seems to have a higher success rate for this particular API
		locationGroupRegionalId, err := common.GetLocationGroupRegional(slsgs.client, template.RegionalGroup.Name)
		if err != nil {
			return data_types.SoftLayer_Scale_Group{},
				fmt.Errorf("Error while looking up regionalGroupId from name [%s]: %s", template.RegionalGroup.Name, err)
		}
		template.RegionalGroupId = locationGroupRegionalId.(int)
		template.RegionalGroup = nil
	}

	parameters := data_types.SoftLayer_Scale_Group_Parameters{
		Parameters: []interface{}{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	log.Printf("[INFO]  ***** request body: %s", requestBody)
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	data, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/createObject", slsgs.GetName()), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#createObject\nResponse from SoftLayer: %s\nHTTP error code: '%d'", data, errorCode)
		return data_types.SoftLayer_Scale_Group{}, errors.New(errorMessage)
	}

	err = slsgs.client.GetHttpClient().CheckForHttpResponseErrors(data)
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	softLayer_Scale_Group := data_types.SoftLayer_Scale_Group{}
	err = json.Unmarshal(data, &softLayer_Scale_Group)
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	return softLayer_Scale_Group, nil
}

func (slsgs *softlayer_Scale_Group_Service) GetNetworkVlans(groupId int, objectMask []string, objectFilter string) ([]data_types.SoftLayer_Scale_Network_Vlan, error) {
	path := fmt.Sprintf("%s/%d/%s", slsgs.GetName(), groupId, "getNetworkVlans.json")

	responseBytes, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequestWithObjectFilterAndObjectMask(path, objectMask, objectFilter, "GET", &bytes.Buffer{})
	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#getNetworkVlans, error message '%s'", err.Error())
		return []data_types.SoftLayer_Scale_Network_Vlan{}, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#getNetworkVlans, HTTP error code: '%d'", errorCode)
		return []data_types.SoftLayer_Scale_Network_Vlan{}, errors.New(errorMessage)
	}

	vlans := []data_types.SoftLayer_Scale_Network_Vlan{}
	err = json.Unmarshal(responseBytes, &vlans)
	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: failed to decode JSON response, err message '%s'", err.Error())
		err := errors.New(errorMessage)
		return []data_types.SoftLayer_Scale_Network_Vlan{}, err
	}

	return vlans, nil
}

func (slsgs *softlayer_Scale_Group_Service) GetObject(groupId int, objectMask []string) (data_types.SoftLayer_Scale_Group, error) {

	response, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequestWithObjectMask(
		fmt.Sprintf("%s/%d/getObject.json", slsgs.GetName(), groupId),
		objectMask,
		"GET",
		new(bytes.Buffer))
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#getObject, HTTP error code: '%d'", errorCode)
		return data_types.SoftLayer_Scale_Group{}, errors.New(errorMessage)
	}

	log.Printf("[INFO]  ***** response json: %s", response)

	group := data_types.SoftLayer_Scale_Group{}
	err = json.Unmarshal(response, &group)
	if err != nil {
		return data_types.SoftLayer_Scale_Group{}, err
	}

	return group, nil
}

func (slsgs *softlayer_Scale_Group_Service) EditObject(groupId int, template data_types.SoftLayer_Scale_Group) (bool, error) {
	parameters := data_types.SoftLayer_Scale_Group_Parameters{
		Parameters: []interface{}{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	log.Printf("[INFO]  ***** request body: %s", requestBody)
	if err != nil {
		return false, err
	}

	response, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/editObject.json", slsgs.GetName(), groupId), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to edit SoftLayer Scale Group with id: %d, got '%s' as response from the API.", groupId, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#editObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

func (slsgs *softlayer_Scale_Group_Service) ForceDeleteObject(group int) (bool, error) {
	response, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/forceDeleteObject", slsgs.GetName(), group), "GET", new(bytes.Buffer))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, fmt.Errorf("Failed to force delete scale group with id '%d', got '%s' as response from the API.", group, res)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Group#forceDeleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}
