package data_types

import "time"

const (
	SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_REPEATING    = 2
	SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_ONE_TIME     = 3
	SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_RESOURCE_USE = 1
)

type SoftLayer_Scale_Policy_Parameters struct {
	Parameters []SoftLayer_Scale_Policy `json:"parameters"`
}

type SoftLayer_Scale_Policy struct {
	ScaleActions        []SoftLayer_Scale_Policy_Action              `json:"scaleActions,omitempty"`
	Cooldown            int                                          `json:"cooldown,omitempty"`
	Id                  int                                          `json:"id,omitempty"`
	Name                string                                       `json:"name,omitempty"`
	ScaleGroupId        int                                          `json:"scaleGroupId,omitempty"`
	Triggers            []SoftLayer_Scale_Policy_Trigger             `json:"triggers,omitempty"`
	OneTimeTriggers     []SoftLayer_Scale_Policy_Trigger_OneTime     `json:"oneTimeTriggers,omitempty"`
	RepeatingTriggers   []SoftLayer_Scale_Policy_Trigger_Repeating   `json:"repeatingTriggers,omitempty"`
	ResourceUseTriggers []SoftLayer_Scale_Policy_Trigger_ResourceUse `json:"resourceUseTriggers,omitempty"`
}

type SoftLayer_Scale_Policy_Action struct {
	Id        int    `json:"id,omitempty"`
	TypeId    int    `json:"typeId,omitempty"`
	Amount    int    `json:"amount,omitempty"`
	ScaleType string `json:"scaleType,omitempty"`
}

type SoftLayer_Scale_Policy_Action_Scale struct {
	SoftLayer_Scale_Policy_Action
	Amount    int    `json:"amount,omitempty"`
	ScaleType string `json:"scaleType,omitempty"`
}

type SoftLayer_Scale_Policy_Trigger struct {
	Id     int `json:"id,omitempty"`
	TypeId int `json:"typeId,omitempty"`
}

type SoftLayer_Scale_Policy_Trigger_OneTime struct {
	SoftLayer_Scale_Policy_Trigger
	Date *time.Time `json:"date,omitempty"`
}

type SoftLayer_Scale_Policy_Trigger_Repeating struct {
	SoftLayer_Scale_Policy_Trigger
	Schedule string `json:"schedule,omitempty"`
}

type SoftLayer_Scale_Policy_Trigger_ResourceUse struct {
	SoftLayer_Scale_Policy_Trigger
	Watches []SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch `json:"watches,omitempty"`
}

type SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch struct {
	Id        int    `json:"id,omitempty"`
	Metric    string `json:"metric,omitempty"`
	Operator  string `json:"operator,omitempty"`
	Period    int    `json:"period,omitempty"`
	Value     string `json:"value,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}
