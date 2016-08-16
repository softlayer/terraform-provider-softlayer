package data_types

type SoftLayer_Scale_Group_Parameters struct {
	Parameters []interface{} `json:"parameters"`
}

type SoftLayer_Scale_Group struct {
	Cooldown                   int                                 `json:"cooldown,omitempty"`
	Id                         int                                 `json:"id,omitempty"`
	LoadBalancers              []SoftLayer_Scale_LoadBalancer      `json:"loadBalancers,omitempty"`
	MaximumMemberCount         int                                 `json:"maximumMemberCount,omitempty"`
	MinimumMemberCount         int                                 `json:"minimumMemberCount,omitempty"`
	Name                       string                              `json:"name,omitempty"`
	NetworkVlans               []SoftLayer_Scale_Network_Vlan      `json:"networkVlans,omitempty"`
	Policies                   []SoftLayer_Scale_Policy            `json:"policies,omitempty"`
	RegionalGroup              *SoftLayer_Location_Group_Regional  `json:"regionalGroup,omitempty"`
	RegionalGroupId            int                                 `json:"regionalGroupId,omitempty"`
	SuspendedFlag              bool                                `json:"suspendedFlag"`
	TerminationPolicy          *SoftLayer_Scale_Termination_Policy `json:"terminationPolicy,omitempty"`
	VirtualGuestMemberTemplate SoftLayer_Virtual_Guest_Template    `json:"virtualGuestMemberTemplate,omitempty"`
	Status                     *SoftLayer_Scale_Group_Status       `json:"status,omitempty"`
}

type SoftLayer_Scale_Termination_Policy struct {
	KeyName string `json:"keyName,omitempty"`
}
