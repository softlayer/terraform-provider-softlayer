package softlayer

type SoftLayer_Scale_Network_Vlan_Service interface {
	Service

	DeleteObject(scaleNetworkVlanId int) (bool, error)
}
