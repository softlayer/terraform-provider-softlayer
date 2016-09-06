package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TheWeatherCompany/softlayer-go/common"
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type softLayer_Load_Balancer_Service_Group struct {
	client softlayer.Client
}

func NewSoftLayer_Load_Balancer_Service_Group_Service(client softlayer.Client) *softLayer_Load_Balancer_Service_Group {
	return &softLayer_Load_Balancer_Service_Group{
		client: client,
	}
}

func (slnadcsgs *softLayer_Load_Balancer_Service_Group) GetName() string {
	return "SoftLayer_Network_Application_Delivery_Controller_LoadBalancer_Service_Group"
}

func (slnadcsgs *softLayer_Load_Balancer_Service_Group) GetObject(id int, objectMask []string) (datatypes.Softlayer_Service_Group, error) {
	response, errorCode, err := slnadcsgs.client.GetHttpClient().DoRawHttpRequestWithObjectMask(
		fmt.Sprintf("%s/%d/getObject.json", slnadcsgs.GetName(), id),
		objectMask,
		"GET",
		new(bytes.Buffer))

	if err != nil {
		return datatypes.Softlayer_Service_Group{}, err
	}

	if common.IsHttpErrorCode(errorCode) {
		return datatypes.Softlayer_Service_Group{},
			fmt.Errorf("softlayer-go: could not %s#getObject, HTTP error code: '%d'", slnadcsgs.GetName(), errorCode)
	}

	serviceGroup := datatypes.Softlayer_Service_Group{}
	err = json.Unmarshal(response, &serviceGroup)
	if err != nil {
		return datatypes.Softlayer_Service_Group{}, err
	}

	return serviceGroup, nil
}
