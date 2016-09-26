package softlayer

import (
	"errors"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/softlayer/softlayer-go/session"
)

func Provider() terraform.ResourceProvider {
	defaultSoftLayerSession := session.New()
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return defaultSoftLayerSession.UserName, nil
				},
				Description: "The user name for SoftLayer API operations.",
			},
			"api_key": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return defaultSoftLayerSession.APIKey, nil
				},
				Description: "The API key for SoftLayer API operations.",
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return defaultSoftLayerSession.Endpoint, nil
				},
				Description: "The endpoint url for the SoftLayer API.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"softlayer_ssh_key": dataSourceSoftLayerSSHKey(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"softlayer_virtual_guest":          resourceSoftLayerVirtualGuest(),
			"softlayer_ssh_key":                resourceSoftLayerSSHKey(),
			"softlayer_dns_domain_record":      resourceSoftLayerDnsDomainRecord(),
			"softlayer_dns_domain":             resourceSoftLayerDnsDomain(),
			"softlayer_lb_vpx":                 resourceSoftLayerLbVpx(),
			"softlayer_lb_vpx_vip":             resourceSoftLayerLbVpxVip(),
			"softlayer_lb_vpx_service":         resourceSoftLayerLbVpxService(),
			"softlayer_lb_local":               resourceSoftLayerLbLocal(),
			"softlayer_lb_local_service_group": resourceSoftLayerLbLocalServiceGroup(),
			"softlayer_lb_local_service":       resourceSoftLayerLbLocalService(),
			"softlayer_security_certificate":   resourceSoftLayerSecurityCertificate(),
			"softlayer_user":                   resourceSoftLayerUser(),
			"softlayer_objectstorage_account":  resourceSoftLayerObjectStorageAccount(),
			"softlayer_provisioning_hook":      resourceSoftLayerProvisioningHook(),
			"softlayer_scale_policy":           resourceSoftLayerScalePolicy(),
			"softlayer_scale_group":            resourceSoftLayerScaleGroup(),
			"softlayer_basic_monitor":          resourceSoftLayerBasicMonitor(),
			"softlayer_vlan":                   resourceSoftLayerVlan(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	sess := session.Session{
		UserName: d.Get("username").(string),
		APIKey:   d.Get("api_key").(string),
		Endpoint: d.Get("endpoint_url").(string),
	}

	if sess.UserName == "" || sess.APIKey == "" {
		return nil, errors.New(
			"No SoftLayer credentials were found. Please ensure you have specified" +
				" them in the provider or in the environment (see the documentation).",
		)
	}

	if os.Getenv("TF_LOG") != "" {
		sess.Debug = true
	}

	return &sess, nil
}
