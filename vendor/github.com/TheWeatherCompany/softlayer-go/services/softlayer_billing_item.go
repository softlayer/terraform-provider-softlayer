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

type softLayer_Billing_Item_Service struct {
	client softlayer.Client
}

func NewSoftLayer_Billing_Item_Service(client softlayer.Client) *softLayer_Billing_Item_Service {
	return &softLayer_Billing_Item_Service{
		client: client,
	}
}

func (slbi *softLayer_Billing_Item_Service) GetName() string {
	return "SoftLayer_Billing_Item"
}

func (slbi *softLayer_Billing_Item_Service) CancelService(billingId int) (bool, error) {
	response, errorCode, err := slbi.client.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/cancelService.json", slbi.GetName(), billingId), "GET", new(bytes.Buffer))
	if err != nil {
		return false, err
	}

	if res := string(response[:]); res != "true" {
		return false, nil
	}

	if common.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Billing_Item#CancelService, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	return true, err
}

// Checks the status of any order. Requires an order receipt.
// Returns whether the order has been completely processed or not (true/false), along with
// the billing order item, and an error.
// You can also pass a list of target bill order statuses to look for. The function will return true if the bill order
// status matches any of those, in addition of matching against 'COMPLETE', which is always the final order state.
func (slbi *softLayer_Billing_Item_Service) CheckOrderStatus(receipt *data_types.SoftLayer_Container_Product_Order_Receipt, targetStatuses []string) (bool, data_types.SoftLayer_Billing_Order_Item, error) {
	response, httpCode, err :=
		slbi.client.GetHttpClient().DoRawHttpRequest(
			fmt.Sprintf(
				"SoftLayer_Billing_Order_Item/%d/getObject.json?objectMask=mask[id,billingItem[id,provisionTransaction[id,transactionStatus[name]]]]",
				receipt.PlacedOrder.Items[0].Id,
			),
			"GET", new(bytes.Buffer),
		)

	if err != nil {
		return false, data_types.SoftLayer_Billing_Order_Item{}, err
	}

	if common.IsHttpErrorCode(httpCode) {
		errorMessage := fmt.Sprintf("softlayer-go: SoftLayer_Billing_Order_Item#getObject, HTTP error code: '%d'", httpCode)
		return false, data_types.SoftLayer_Billing_Order_Item{}, errors.New(errorMessage)
	}

	billingOrderItem := data_types.SoftLayer_Billing_Order_Item{}
	err = json.Unmarshal(response, &billingOrderItem)
	if err != nil {
		return false, data_types.SoftLayer_Billing_Order_Item{}, errors.New(
			fmt.Sprintf(
				"softlayer-go: Unmarshaling response from SoftLayer_Billing_Order_Item#getObject: %s", err))
	}

	done := false
	billOrderStatus := billingOrderItem.BillingItem.ProvisionTransaction.TransactionStatus.Name
	for _, status := range targetStatuses {
		if status == billOrderStatus {
			done = true
			break
		}
	}

	if billOrderStatus == "COMPLETE" {
		done = true
	}

	return done, billingOrderItem, nil
}
