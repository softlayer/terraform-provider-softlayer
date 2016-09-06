package softlayer

import (
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
)

type SoftLayer_Load_Balancer_Service_Group_Service interface {
	Service

	GetObject(id int, objectMask []string) (datatypes.Softlayer_Service_Group, error)
}
