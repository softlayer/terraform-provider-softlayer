package softlayer

import (
	"github.com/TheWeatherCompany/softlayer-go/data_types"
)

type SoftLayer_Scale_Group_Service interface {
	Service

	CreateObject(template data_types.SoftLayer_Scale_Group) (data_types.SoftLayer_Scale_Group, error)
	GetNetworkVlans(groupId int, objectMask []string, objectFilter string) ([]data_types.SoftLayer_Scale_Network_Vlan, error)
	GetObject(groupId int, objectMask []string) (data_types.SoftLayer_Scale_Group, error)
	EditObject(groupId int, template data_types.SoftLayer_Scale_Group) (bool, error)
	ForceDeleteObject(scaleGroupId int) (bool, error)
}
