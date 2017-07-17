package softlayer

import (
	"errors"
	"os"
	"time"

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
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The timeout (in seconds) to set for any SoftLayer API calls made.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"softlayer_ssh_key":        dataSourceSoftLayerSSHKey(),
			"softlayer_image_template": dataSourceSoftLayerImageTemplate(),
			"softlayer_vlan":           dataSourceSoftLayerVlan(),
			"softlayer_dns_domain":     dataSourceSoftLayerDnsDomain(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"softlayer_virtual_guest":               resourceSoftLayerVirtualGuest(),
			"softlayer_bare_metal":                  resourceSoftLayerBareMetal(),
			"softlayer_ssh_key":                     resourceSoftLayerSSHKey(),
			"softlayer_dns_domain_record":           resourceSoftLayerDnsDomainRecord(),
			"softlayer_dns_domain":                  resourceSoftLayerDnsDomain(),
			"softlayer_lb_vpx":                      resourceSoftLayerLbVpx(),
			"softlayer_lb_vpx_vip":                  resourceSoftLayerLbVpxVip(),
			"softlayer_lb_vpx_service":              resourceSoftLayerLbVpxService(),
			"softlayer_lb_vpx_ha":                   resourceSoftLayerLbVpxHa(),
			"softlayer_lb_local":                    resourceSoftLayerLbLocal(),
			"softlayer_lb_local_service_group":      resourceSoftLayerLbLocalServiceGroup(),
			"softlayer_lb_local_service":            resourceSoftLayerLbLocalService(),
			"softlayer_security_certificate":        resourceSoftLayerSecurityCertificate(),
			"softlayer_user":                        resourceSoftLayerUser(),
			"softlayer_objectstorage_account":       resourceSoftLayerObjectStorageAccount(),
			"softlayer_provisioning_hook":           resourceSoftLayerProvisioningHook(),
			"softlayer_scale_policy":                resourceSoftLayerScalePolicy(),
			"softlayer_scale_group":                 resourceSoftLayerScaleGroup(),
			"softlayer_basic_monitor":               resourceSoftLayerBasicMonitor(),
			"softlayer_vlan":                        resourceSoftLayerVlan(),
			"softlayer_global_ip":                   resourceSoftLayerGlobalIp(),
			"softlayer_fw_hardware_dedicated":       resourceSoftLayerFwHardwareDedicated(),
			"softlayer_fw_hardware_dedicated_rules": resourceSoftLayerFwHardwareDedicatedRules(),
			"softlayer_file_storage":                resourceSoftLayerFileStorage(),
			"softlayer_block_storage":               resourceSoftLayerBlockStorage(),
			"softlayer_dns_secondary":               resourceSoftLayerDnsSecondary(),
			"softlayer_subnet":                      resourceSoftLayerSubnet(),
		},

		ConfigureFunc: providerConfigure,
	}
}

type ProviderConfig interface {
	SoftLayerSession() *session.Session
}

type providerConfig struct {
	Session *session.Session
}

func (config providerConfig) SoftLayerSession() *session.Session {
	return config.Session
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	sess := session.Session{
		UserName: d.Get("username").(string),
		APIKey:   d.Get("api_key").(string),
		Endpoint: d.Get("endpoint_url").(string),
	}

	if rawTimeout, ok := d.GetOk("timeout"); ok {
		timeout := rawTimeout.(int)
		sess.Timeout = time.Duration(timeout)
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

	return providerConfig{Session: &sess}, nil
}
