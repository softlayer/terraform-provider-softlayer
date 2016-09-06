package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	"github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type softlayer_user_customer_service struct {
	client softlayer.Client
}

func NewSoftLayer_User_Customer_Service(client softlayer.Client) *softlayer_user_customer_service {
	return &softlayer_user_customer_service{
		client: client,
	}
}

func (slucs *softlayer_user_customer_service) GetName() string {
	return "SoftLayer_User_Customer"
}

func (slucs *softlayer_user_customer_service) CreateObject(template data_types.SoftLayer_User_Customer, password string) (data_types.SoftLayer_User_Customer, error) {
	parameters := data_types.SoftLayer_User_Customer_Parameters{
		Parameters: []interface{}{
			template,
			password,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	data, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/createObject", slucs.GetName()), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_User_Customer#createObject\nResponse from SoftLayer: %s\nHTTP error code: '%d'", data, errorCode)
		return data_types.SoftLayer_User_Customer{}, errors.New(errorMessage)
	}

	err = slucs.client.GetHttpClient().CheckForHttpResponseErrors(data)
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	softLayer_User_Customer := data_types.SoftLayer_User_Customer{}
	err = json.Unmarshal(data, &softLayer_User_Customer)
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	return softLayer_User_Customer, nil
}

/*
 * https://api.softlayer.com/rest/v3/SoftLayer_User_Customer/630887.json?objectMask=id;username;\
       firstname;lastname;email;companyName;address1;address2;city;state;country;timezoneId;\
       userStatusId;displayName;parentId;permissions;apiAuthenticationKeys
*/
func (slucs *softlayer_user_customer_service) GetObject(userid int) (data_types.SoftLayer_User_Customer, error) {
	objectMask := []string{
		"id",
		"username",
		"firstName",
		"lastName",
		"email",
		"companyName",
		"address1",
		"address2",
		"city",
		"state",
		"country",
		"timezoneId",
		"userStatusId",
		"displayName",
		"parentId",
		"permissions",
		"apiAuthenticationKeys",
	}

	response, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequestWithObjectMask(fmt.Sprintf("%s/%d/getObject.json", slucs.GetName(), userid), objectMask, "GET", new(bytes.Buffer))
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_User_Customer#getObject, HTTP error code: '%d'", errorCode)
		return data_types.SoftLayer_User_Customer{}, errors.New(errorMessage)
	}

	user := data_types.SoftLayer_User_Customer{}
	err = json.Unmarshal(response, &user)
	if err != nil {
		return data_types.SoftLayer_User_Customer{}, err
	}

	return user, nil
}

func (slucs *softlayer_user_customer_service) EditObject(userid int, template data_types.SoftLayer_User_Customer) (bool, error) {
	parameters := data_types.SoftLayer_User_Customer_Parameters{
		Parameters: []interface{}{
			template,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, err
	}

	response, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/editObject.json", slucs.GetName(), userid), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to edit SoftLayer user with id: %d, got '%s' as response from the API.", userid, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_User_Customer#editObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

func (slucs *softlayer_user_customer_service) DeleteObject(userid int) (bool, error) {
	deleteForm := data_types.SoftLayer_User_Customer_Delete{
		UserStatus: 1021,
	}
	parameters := data_types.SoftLayer_User_Customer_Parameters{
		Parameters: []interface{}{
			deleteForm,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, err
	}

	response, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/editObject.json", slucs.GetName(), userid), "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, errors.New(fmt.Sprintf("Failed to destroy SoftLayer user with id '%d', got '%s' as response from the API.", userid, res))
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_User_Customer#deleteObject, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

/*
 * https://api.softlayer.com/rest/v3/SoftLayer_User_Customer/630887/getApiAuthenticationKeys.json
 */
func (slucs *softlayer_user_customer_service) GetApiAuthenticationKeys(userId int) ([]data_types.SoftLayer_User_Customer_ApiAuthentication, error) {
	path := fmt.Sprintf("%s/%d/%s.json", slucs.GetName(), userId, "getApiAuthenticationKeys")
	responseBytes, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "GET", &bytes.Buffer{})
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#getApiAuthenticationKeys, error message '%s'",
			slucs.GetName(), err.Error(),
		)
		return nil, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#getApiAuthenticationKeys, HTTP error code: '%d'",
			slucs.GetName(), errorCode,
		)
		return nil, errors.New(errorMessage)
	}

	apiKeys := []data_types.SoftLayer_User_Customer_ApiAuthentication{}
	err = json.Unmarshal(responseBytes, &apiKeys)
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: failed to decode JSON response from %s#getApiAuthenticationKeys, err message '%s'",
			slucs.GetName(), err.Error(),
		)
		err := errors.New(errorMessage)
		return nil, err
	}

	return apiKeys, nil
}

/*
 * https://api.softlayer.com/rest/v3/SoftLayer_User_Customer/630887/addApiAuthenticationKey.json
 */
func (slucs *softlayer_user_customer_service) AddApiAuthenticationKey(userId int) error {
	path := fmt.Sprintf("%s/%d/%s.json", slucs.GetName(), userId, "addApiAuthenticationKey")
	_, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "GET", &bytes.Buffer{})
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#addApiAuthenticationKey, error message '%s'",
			slucs.GetName(), err.Error(),
		)
		return errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#addApiAuthenticationKey, HTTP error code: '%d'",
			slucs.GetName(), errorCode,
		)
		return errors.New(errorMessage)
	}

	return nil
}

/*
 * https://api.softlayer.com/rest/v3/SoftLayer_User_Customer/630887/removeApiAuthenticationKey/871423.json
 */
func (slucs *softlayer_user_customer_service) RemoveApiAuthenticationKey(userId int, apiKeys []data_types.SoftLayer_User_Customer_ApiAuthentication) (bool, error) {
	// Even though a whole api auth key structure is passed as input parameter, one softlayer user login can only have one api auth key.
	// Extract the api auth key id.
	apiAuthKeyId := apiKeys[0].Id
	path := fmt.Sprintf("%s/%d/%s/%d.json", slucs.GetName(), userId, "removeApiAuthenticationKey", apiAuthKeyId)
	response, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "GET", &bytes.Buffer{})
	if err != nil {
		errorMessage := fmt.Sprintf("softlayer-go: could not %s, error message: '%s'", path, err.Error())
		return false, errors.New(errorMessage)
	}

	if res := string(response[:]); res != "true" {
		errorMessage := fmt.Sprintf("softlayer-go: could not %s. Response received: '%s'", path, res)
		return false, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not %s. HTTP error code: '%d'", path, errorCode)
		return false, errors.New(errorMessage)
	}

	return true, nil
}

func (slucs *softlayer_user_customer_service) AddBulkPortalPermission(userId int, permissions []data_types.SoftLayer_User_Customer_CustomerPermission_Permission) error {
	parameters := data_types.SoftLayer_User_Customer_Parameters{
		Parameters: []interface{}{
			permissions,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return nil
	}

	path := fmt.Sprintf("%s/%d/%s", slucs.GetName(), userId, "addBulkPortalPermission.json")
	_, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "PUT", bytes.NewBuffer(requestBody))
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#addBulkPortalPermission, error message '%s'",
			slucs.GetName(), err.Error(),
		)
		return errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#addBulkPortalPermission, HTTP error code: '%d'",
			slucs.GetName(), errorCode,
		)
		return errors.New(errorMessage)
	}

	return nil
}

func (slucs *softlayer_user_customer_service) RemoveBulkPortalPermission(userId int, permissions []data_types.SoftLayer_User_Customer_CustomerPermission_Permission) error {
	parameters := data_types.SoftLayer_User_Customer_Parameters{
		Parameters: []interface{}{
			permissions,
		},
	}

	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return nil
	}

	path := fmt.Sprintf("%s/%d/%s", slucs.GetName(), userId, "removeBulkPortalPermission.json")
	_, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "PUT", bytes.NewBuffer(requestBody))
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#removeBulkPortalPermission, error message '%s'",
			slucs.GetName(), err.Error(),
		)
		return errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#removeBulkPortalPermission, HTTP error code: '%d'",
			slucs.GetName(), errorCode,
		)
		return errors.New(errorMessage)
	}

	return nil
}

func (slucs *softlayer_user_customer_service) GetPermissions(userId int) ([]data_types.SoftLayer_User_Customer_CustomerPermission_Permission, error) {
	path := fmt.Sprintf("%s/%d/%s", slucs.GetName(), userId, "getPermissions.json")
	responseBytes, errorCode, err := slucs.client.GetHttpClient().DoRawHttpRequest(path, "GET", &bytes.Buffer{})
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#getPermissions, error message '%s'",
			slucs.GetName(), err.Error(),
		)
		return nil, errors.New(errorMessage)
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf(
			"softlayer-go: could not %s#getPermissions, HTTP error code: '%d'",
			slucs.GetName(), errorCode,
		)
		return nil, errors.New(errorMessage)
	}

	permissions := []data_types.SoftLayer_User_Customer_CustomerPermission_Permission{}
	err = json.Unmarshal(responseBytes, &permissions)
	if err != nil {
		errorMessage := fmt.Sprintf(
			"softlayer-go: failed to decode JSON response from %s#getPermissions, err message '%s'",
			slucs.GetName(), err.Error(),
		)
		err := errors.New(errorMessage)
		return nil, err
	}

	return permissions, nil
}
