package softlayer

import (
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
)

type SoftLayer_Provisioning_Hook_Service interface {
	Service

	CreateObject(template datatypes.SoftLayer_Provisioning_Hook_Template) (datatypes.SoftLayer_Provisioning_Hook, error)
	GetObject(id int) (datatypes.SoftLayer_Provisioning_Hook, error)
	EditObject(id int, template datatypes.SoftLayer_Provisioning_Hook_Template) (bool, error)
	DeleteObject(id int) (bool, error)
}
