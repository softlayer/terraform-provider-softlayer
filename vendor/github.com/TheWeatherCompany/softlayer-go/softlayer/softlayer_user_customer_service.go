package softlayer

import (
	"github.com/TheWeatherCompany/softlayer-go/data_types"
)

type SoftLayer_User_Customer_Service interface {
	Service

	CreateObject(template data_types.SoftLayer_User_Customer, password string) (data_types.SoftLayer_User_Customer, error)
	GetObject(userId int) (data_types.SoftLayer_User_Customer, error)
	EditObject(userId int, template data_types.SoftLayer_User_Customer) (bool, error)
	DeleteObject(userId int) (bool, error)

	AddApiAuthenticationKey(userId int) error
	GetApiAuthenticationKeys(userId int) ([]data_types.SoftLayer_User_Customer_ApiAuthentication, error)
	RemoveApiAuthenticationKey(userId int, apiKeys []data_types.SoftLayer_User_Customer_ApiAuthentication) (bool, error)

	AddBulkPortalPermission(userId int, permissions []data_types.SoftLayer_User_Customer_CustomerPermission_Permission) error
	RemoveBulkPortalPermission(userId int, permissions []data_types.SoftLayer_User_Customer_CustomerPermission_Permission) error
	GetPermissions(userId int) ([]data_types.SoftLayer_User_Customer_CustomerPermission_Permission, error)
}
