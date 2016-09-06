package softlayer

import (
	"fmt"
	"log"

	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
	"strings"
)

func resourceSoftLayerSecurityCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerSecurityCertificateCreate,
		Read:   resourceSoftLayerSecurityCertificateRead,
		Delete: resourceSoftLayerSecurityCertificateDelete,
		Exists: resourceSoftLayerSecurityCertificateExists,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},

			"certificate": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},

			"intermediate_certificate": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},

			"private_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: normalizeCert,
			},

			"common_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"organization_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"validity_begin": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"validity_days": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"validity_end": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"key_size": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"create_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"modify_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSoftLayerSecurityCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService

	if client == nil {
		return fmt.Errorf("The client is nil.")
	}

	template := datatypes.SoftLayer_Security_Certificate_Template{
		Certificate:             d.Get("certificate").(string),
		IntermediateCertificate: d.Get("intermediate_certificate").(string),
		PrivateKey:              d.Get("private_key").(string),
	}

	log.Printf("[INFO] Creating Security Certificate")

	cert, err := client.CreateSecurityCertificate(template)

	if err != nil {
		return fmt.Errorf("Error creating Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", cert.Id))

	return resourceSoftLayerSecurityCertificateRead(d, meta)
}

func resourceSoftLayerSecurityCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService
	if client == nil {
		return fmt.Errorf("The client is nil.")
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := client.GetObject(id)

	if err != nil {
		return fmt.Errorf("Unable to get Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", cert.Id))
	d.Set("certificate", cert.Certificate)
	d.Set("intermediate_certificate", cert.IntermediateCertificate)
	d.Set("private_key", cert.PrivateKey)
	d.Set("common_name", cert.CommonName)
	d.Set("organization_name", cert.OrganizationName)
	d.Set("validity_begin", cert.ValidityBegin)
	d.Set("validity_days", cert.ValidityDays)
	d.Set("validity_end", cert.ValidityEnd)
	d.Set("key_size", cert.KeySize)
	d.Set("create_date", cert.CreateDate)
	d.Set("modify_date", cert.ModifyDate)

	return nil
}

func resourceSoftLayerSecurityCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).securityCertificateService

	if client == nil {
		return fmt.Errorf("The client was nil.")
	}

	_, err := client.DeleteObject(d.Get("id").(int))

	if err != nil {
		return fmt.Errorf("Error deleting Security Certificate %s: %s", d.Get("id"), err)
	}

	return nil
}

func resourceSoftLayerSecurityCertificateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client).securityCertificateService

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := client.GetObject(id)

	if err != nil {
		return false, fmt.Errorf("Error fetching Security Cerfiticate: %s", err)
	}

	return cert.Id == id && err == nil, nil
}

func normalizeCert(cert interface{}) string {
	if cert == nil || cert == (*string)(nil) {
		return ""
	}

	switch cert.(type) {
	case string:
		return strings.TrimSpace(cert.(string))
	default:
		return ""
	}
}
