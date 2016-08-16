package softlayer

type SoftLayer_Scale_Policy_Trigger_Service interface {
	Service

	DeleteObject(scalePolicyTriggerId int) (bool, error)
}
