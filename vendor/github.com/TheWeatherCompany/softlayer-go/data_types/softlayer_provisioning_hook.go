package data_types

type SoftLayer_Provisioning_Hook struct {
	Id         int    `json:"id,omitempty"`
	AccountId  int    `json:"accountId,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
	ModifyDate string `json:"modifyDate,omitempty"`
	Name       string `json:"name"`
	TypeId     int    `json:"typeId"`
	Uri        string `json:"uri"`
}

type SoftLayer_Provisioning_Hook_Parameters struct {
	Parameters []SoftLayer_Provisioning_Hook `json:"parameters"`
}

type SoftLayer_Provisioning_Hook_Template struct {
	Name   string `json:"name,omitempty"`
	TypeId int    `json:"typeId,omitempty"`
	Uri    string `json:"uri,omitempty"`
}

type SoftLayer_Provisioning_Hook_Template_Parameters struct {
	Parameters []SoftLayer_Provisioning_Hook_Template `json:"parameters"`
}
