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
	securityCertificateService                  softlayer.SoftLayer_Security_Certificate_Service
	userCustomerService                         softlayer.SoftLayer_User_Customer_Service
	provisioningHookService                     softlayer.SoftLayer_Provisioning_Hook_Service
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

	client := &Client{
		accountService:                              accountService,
		virtualGuestService:                         virtualGuestService,
		sshKeyService:                               sshKeyService,
		productOrderService:                         productOrderService,
		billingItemService:                          billingItemService,
		dnsDomainService:                            dnsDomainService,
		dnsDomainResourceRecordService:              dnsDomainResourceRecordService,
		networkApplicationDeliveryControllerService: networkApplicationDeliveryControllerService,
		securityCertificateService:                  securityCertificateService,
		userCustomerService:                         userCustomerService,
		provisioningHookService:                     provisioningHookService,
	}

	log.Println("[INFO] Created SoftLayer client")

	return client, nil
}
