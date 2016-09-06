package data_types

type SoftLayer_Load_Balancer_Virtual_Server_Update_Parameters struct {
	Parameters []Softlayer_Load_Balancer_Virtual_Server_Parameters `json:"parameters"`
}

type Softlayer_Load_Balancer_Virtual_Server_Parameters struct {
	VirtualServers []*Softlayer_Load_Balancer_Virtual_Server `json:"virtualServers"`
}

type Softlayer_Load_Balancer_Virtual_Server struct {
	Id               int                        `json:"id,omitempty"`
	Allocation       int                        `json:"allocation,omitempty"`
	Port             int                        `json:"port,omitempty"`
	ServiceGroups    []*Softlayer_Service_Group `json:"serviceGroups"`
	VirtualIpAddress *SoftLayer_Load_Balancer   `json:"virtualIpAddress,omitempty"`
}

type Softlayer_Service_Group struct {
	Id              int                                     `json:"id,omitempty"`
	RoutingMethodId int                                     `json:"routingMethodId"`
	RoutingTypeId   int                                     `json:"routingTypeId"`
	RoutingMethod   string                                  `json:"routingMethod,omitempty"`
	RoutingType     string                                  `json:"routingId,omitempty"`
	Services        []*Softlayer_Service                    `json:"services"`
	VirtualServer   *Softlayer_Load_Balancer_Virtual_Server `json:"virtualServer,omitempty"`
}

type Softlayer_Service struct {
	Id              int                          `json:"id,omitempty"`
	Enabled         int                          `json:"enabled"`
	Port            int                          `json:"port"`
	IpAddressId     int                          `json:"ipAddressId"`
	HealthChecks    []*Softlayer_Health_Check    `json:"healthChecks"`
	GroupReferences []*Softlayer_Group_Reference `json:"groupReferences"`
}

type Softlayer_Health_Check struct {
	HealthCheckTypeId int    `json:"healthCheckTypeId"`
	HealthCheckType   string `json:"healthCheckType,omitempty"`
}

type Softlayer_Group_Reference struct {
	Weight int `json:"weight"`
}
