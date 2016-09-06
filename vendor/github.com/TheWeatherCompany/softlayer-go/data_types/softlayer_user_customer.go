package data_types

type SoftLayer_User_Customer_Parameters struct {
	Parameters []interface{} `json:"parameters"`
}

type SoftLayer_User_Customer struct {
	Address1              string                                                  `json:"address1,omitempty"`
	Address2              string                                                  `json:"address2,omitempty"`
	ApiAuthenticationKeys []SoftLayer_User_Customer_ApiAuthentication             `json:"apiAuthenticationKeys,omitempty"`
	City                  string                                                  `json:"city,omitempty"`
	CompanyName           string                                                  `json:"companyName,omitempty"`
	Country               string                                                  `json:"country,omitempty"`
	DisplayName           string                                                  `json:"displayName,omitempty"`
	Email                 string                                                  `json:"email,omitempty"`
	FirstName             string                                                  `json:"firstName,omitempty"`
	Id                    int                                                     `json:"id,omitempty"`
	LastName              string                                                  `json:"lastName,omitempty"`
	ParentId              int                                                     `json:"parentId,omitempty"`
	Permissions           []SoftLayer_User_Customer_CustomerPermission_Permission `json:"permissions"`
	State                 string                                                  `json:"state,omitempty"`
	Timezone              int                                                     `json:"timezoneId,omitempty"`
	UserStatus            int                                                     `json:"userStatusId,omitempty"`
	Username              string                                                  `json:"username,omitempty"`
}

type SoftLayer_User_Customer_ApiAuthentication struct {
	Id                int    `json:"id"`
	AuthenticationKey string `json:"authenticationKey"`
	UserId            int    `json:"userId"`
}

type SoftLayer_User_Customer_CustomerPermission_Permission struct {
	Key     string `json:"key,omitempty"`
	KeyName string `json:"keyName"`
	Name    string `json:"name,omitempty"`
}

// Deleting a user is actually setting the user status to cancel_pending (1021)
type SoftLayer_User_Customer_Delete struct {
	UserStatus int `json:"userStatusId"`
}
