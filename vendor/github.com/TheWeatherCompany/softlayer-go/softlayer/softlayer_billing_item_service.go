package softlayer

import "github.com/TheWeatherCompany/softlayer-go/data_types"

type SoftLayer_Billing_Item_Service interface {
	Service

	CancelService(billingId int) (bool, error)
	CheckOrderStatus(receipt *data_types.SoftLayer_Container_Product_Order_Receipt, targetStatuses []string) (bool, data_types.SoftLayer_Billing_Order_Item, error)
}
