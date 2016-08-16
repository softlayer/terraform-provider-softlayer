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

type softlayer_Scale_Policy_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Scale_Policy_Service(client softlayer.Client) *softlayer_Scale_Policy_Service {
	return &softlayer_Scale_Policy_Service{
		client: client,
	}
}

func (slsps *softlayer_Scale_Policy_Service) GetName() string {
	return "SoftLayer_Scale_Policy"
}

func (slsps *softlayer_Scale_Policy_Service) CreateObject(template datatypes.SoftLayer_Scale_Policy) (datatypes.SoftLayer_Scale_Policy, error) {

	err := slsps.checkCreateObjectRequiredValues(template)
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	parameters := datatypes.SoftLayer_Scale_Policy_Parameters{
		Parameters: []datatypes.SoftLayer_Scale_Policy{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	response, errorCode, err := slsps.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s.json", slsps.GetName()), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Policy#createObject1, HTTP error code: 'Request : %d'"+string(requestBody)+"Resonse : "+string(response), errorCode)
		return datatypes.SoftLayer_Scale_Policy{}, errors.New(errorMessage)
	}

	err = slsps.client.GetHttpClient().CheckForHttpResponseErrors(response)
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	softLayer_Scale_Policy := datatypes.SoftLayer_Scale_Policy{}
	err = json.Unmarshal(response, &softLayer_Scale_Policy)
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}
	return softLayer_Scale_Policy, nil
}

func (slsps *softlayer_Scale_Policy_Service) GetObject(policyId int) (datatypes.SoftLayer_Scale_Policy, error) {
	objectMask := []string{
		"cooldown",
		"id",
		"name",
		"scaleActions",
		"scaleGroupId",
		"oneTimeTriggers",
		"repeatingTriggers",
		"resourceUseTriggers.watches",
		"triggers",
	}

	response, errorCode, err := slsps.client.GetHttpClient().DoRawHttpRequestWithObjectMask(fmt.Sprintf("%s/%d/getObject.json", slsps.GetName(), policyId), objectMask, "GET", new(bytes.Buffer))
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_User_Customer#getObject, HTTP error code: '%d'", errorCode)
		return datatypes.SoftLayer_Scale_Policy{}, errors.New(errorMessage)
	}

	policy := datatypes.SoftLayer_Scale_Policy{}
	err = json.Unmarshal(response, &policy)
	if err != nil {
		return datatypes.SoftLayer_Scale_Policy{}, err
	}

	return policy, nil
}

func (slsps *softlayer_Scale_Policy_Service) EditObject(policyId int, template datatypes.SoftLayer_Scale_Policy) (bool, error) {
	parameters := datatypes.SoftLayer_Scale_Policy_Parameters{
		Parameters: []datatypes.SoftLayer_Scale_Policy{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, err
	}

	response, errorCode, err := slsps.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/editObject.json", slsps.GetName(), policyId), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to edit scale policy with id: %d, got '%s' as response from the API.", policyId, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Policy#editObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

func (slsgs *softlayer_Scale_Policy_Service) DeleteObject(policy int) (bool, error) {
	response, errorCode, err := slsgs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/deleteObject", slsgs.GetName(), policy), "GET", new(bytes.Buffer))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, fmt.Errorf("Failed to force delete scale policy with id '%d', got '%s' as response from the API.", policy, res)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Scale_Policy#deleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}
	return true, err
}

func (slvgs *softlayer_Scale_Policy_Service) checkCreateObjectRequiredValues(template datatypes.SoftLayer_Scale_Policy) error {
	var err error
	errorMessage, errorTemplate := "", "* %s is required and cannot be empty\n"

	if template.Name == "" {
		errorMessage += fmt.Sprintf(errorTemplate, "Name for the scale policy")
	}

	if template.ScaleGroupId == 0 {
		errorMessage += fmt.Sprintf(errorTemplate, "ScaleGroupId for the scale policy")
	}

	if len(template.ScaleActions) == 0 {
		errorMessage += fmt.Sprintf(errorTemplate, "ScaleActions: scaleActions for the scale policy")
	}

	for _, scaleAction := range template.ScaleActions {
		if scaleAction.TypeId == 0 {
			errorMessage += fmt.Sprintf(errorTemplate, "ScaleAction.TypeId: TypeId for the scale action")
		}
		if scaleAction.Amount == 0 {
			errorMessage += fmt.Sprintf(errorTemplate, "ScaleAction.Amount: Amount for the scale action")
		}
		if scaleAction.ScaleType == "" {
			errorMessage += fmt.Sprintf(errorTemplate, "ScaleAction.ScaleType: ScaleType for the scale action")
		}
	}

	if errorMessage != "" {
		err = errors.New(errorMessage)
	}

	return err
}
