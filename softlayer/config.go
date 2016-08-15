package softlayer

import (
	"log"

	slclient "github.com/TheWeatherCompany/softlayer-go/client"
	"github.com/TheWeatherCompany/softlayer-go/softlayer"
)

type Config struct {
	Username string
	ApiKey   string
}

type Client struct {
	accountService                              softlayer.SoftLayer_Account_Service
	virtualGuestService                         softlayer.SoftLayer_Virtual_Guest_Service
	sshKeyService                               softlayer.SoftLayer_Security_Ssh_Key_Service
	productOrderService                         softlayer.SoftLayer_Product_Order_Service
	billingItemService                          softlayer.SoftLayer_Billing_Item_Service
	dnsDomainResourceRecordService              softlayer.SoftLayer_Dns_Domain_ResourceRecord_Service
	dnsDomainService                            softlayer.SoftLayer_Dns_Domain_Service
	networkApplicationDeliveryControllerService softlayer.SoftLayer_Network_Application_Delivery_Controller_Service
	loadBalancerService                         softlayer.SoftLayer_Load_Balancer_Service
	loadBalancerServiceGroupService             softlayer.SoftLayer_Load_Balancer_Service_Group_Service
	securityCertificateService                  softlayer.SoftLayer_Security_Certificate_Service
	userCustomerService                         softlayer.SoftLayer_User_Customer_Service
	provisioningHookService                     softlayer.SoftLayer_Provisioning_Hook_Service
	scalePolicyService                          softlayer.SoftLayer_Scale_Policy_Service
	scalePolicyTriggerService                   softlayer.SoftLayer_Scale_Policy_Trigger_Service
	scaleGroupService                           softlayer.SoftLayer_Scale_Group_Service
	scaleNetworkVlanService                     softlayer.SoftLayer_Scale_Network_Vlan_Service
}

func (c *Config) Client() (*Client, error) {
	slc := slclient.NewSoftLayerClient(c.Username, c.ApiKey)

	accountService, err := slc.GetSoftLayer_Account_Service()

	if err != nil {
		return nil, err
	}

	virtualGuestService, err := slc.GetSoftLayer_Virtual_Guest_Service()

	if err != nil {
		return nil, err
	}

	networkApplicationDeliveryControllerService, err := slc.GetSoftLayer_Network_Application_Delivery_Controller_Service()

	if err != nil {
		return nil, err
	}

	loadBalancerService, err := slc.GetSoftLayer_Load_Balancer_Service()

	if err != nil {
		return nil, err
	}

	loadBalancerServiceGroupService, err := slc.GetSoftLayer_Load_Balancer_Service_Group_Service()

	if err != nil {
		return nil, err
	}

	sshKeyService, err := slc.GetSoftLayer_Security_Ssh_Key_Service()

	if err != nil {
		return nil, err
	}

	productOrderService, err := slc.GetSoftLayer_Product_Order_Service()

	if err != nil {
		return nil, err
	}

	billingItemService, err := slc.GetSoftLayer_Billing_Item_Service()

	if err != nil {
		return nil, err
	}

	dnsDomainService, err := slc.GetSoftLayer_Dns_Domain_Service()

	if err != nil {
		return nil, err
	}

	dnsDomainResourceRecordService, err := slc.GetSoftLayer_Dns_Domain_ResourceRecord_Service()

	if err != nil {
		return nil, err
	}

	provisioningHookService, err := slc.GetSoftLayer_Provisioning_Hook_Service()

	if err != nil {
		return nil, err
	}

	securityCertificateService, err := slc.GetSoftLayer_Security_Certificate_Service()

	if err != nil {
		return nil, err
	}

	userCustomerService, err := slc.GetSoftLayer_User_Customer_Service()

	if err != nil {
		return nil, err
	}

	scalePolicyService, err := slc.GetSoftLayer_Scale_Policy_Service()

	if err != nil {
		return nil, err
	}

	scalePolicyTriggerService, err := slc.GetSoftLayer_Scale_Policy_Trigger_Service()

	if err != nil {
		return nil, err
	}

	scaleGroupService, err := slc.GetSoftLayer_Scale_Group_Service()

	if err != nil {
		return nil, err
	}

	scaleNetworkVlanService, err := slc.GetSoftLayer_Scale_Network_Vlan_Service()

	if err != nil {
		return nil, err
	}

	client := &Client{
		accountService:                              accountService,
		virtualGuestService:                         virtualGuestService,
		sshKeyService:                               sshKeyService,
		productOrderService:                         productOrderService,
		billingItemService:                          billingItemService,
		dnsDomainService:                            dnsDomainService,
		dnsDomainResourceRecordService:              dnsDomainResourceRecordService,
		networkApplicationDeliveryControllerService: networkApplicationDeliveryControllerService,
		loadBalancerService:                         loadBalancerService,
		loadBalancerServiceGroupService:             loadBalancerServiceGroupService,
		securityCertificateService:                  securityCertificateService,
		userCustomerService:                         userCustomerService,
		provisioningHookService:                     provisioningHookService,
		scalePolicyService:                          scalePolicyService,
		scalePolicyTriggerService:                   scalePolicyTriggerService,
		scaleGroupService:                           scaleGroupService,
		scaleNetworkVlanService:                     scaleNetworkVlanService,
	}

	log.Println("[INFO] Created SoftLayer client")

	return client, nil
}
