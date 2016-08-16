package data_types

type SoftLayer_Scale_Network_Vlan struct {
	Id            int                     `json:"id,omitempty"`
	NetworkVlan   *SoftLayer_Network_Vlan `json:"networkVlan,omitempty"`
	NetworkVlanId int                     `json:"networkVlanId,omitempty"`
}
