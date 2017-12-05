package softlayer

import (
	"fmt"
	"log"

	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerSecurityCertificate() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerSecurityCertificateCreate,
		Read:     resourceSoftLayerSecurityCertificateRead,
		Delete:   resourceSoftLayerSecurityCertificateDelete,
		Exists:   resourceSoftLayerSecurityCertificateExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
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
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecurityCertificateService(sess)

	template := datatypes.Security_Certificate{
		Certificate:             sl.String(d.Get("certificate").(string)),
		IntermediateCertificate: sl.String(d.Get("intermediate_certificate").(string)),
		PrivateKey:              sl.String(d.Get("private_key").(string)),
	}

	log.Printf("[INFO] Creating Security Certificate")

	cert, err := service.CreateObject(&template)

	if err != nil {
		return fmt.Errorf("Error creating Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *cert.Id))

	return resourceSoftLayerSecurityCertificateRead(d, meta)
}

func resourceSoftLayerSecurityCertificateRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecurityCertificateService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := service.Id(id).GetObject()

	if err != nil {
		return fmt.Errorf("Unable to get Security Certificate: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", *cert.Id))
	d.Set("certificate", *cert.Certificate)
	if cert.IntermediateCertificate != nil {
		d.Set("intermediate_certificate", *cert.IntermediateCertificate)
	}
	if cert.PrivateKey != nil {
		d.Set("private_key", *cert.PrivateKey)
	}
	d.Set("common_name", *cert.CommonName)
	d.Set("organization_name", *cert.OrganizationName)
	d.Set("validity_begin", *cert.ValidityBegin)
	d.Set("validity_days", *cert.ValidityDays)
	d.Set("validity_end", *cert.ValidityEnd)
	d.Set("key_size", *cert.KeySize)
	d.Set("create_date", *cert.CreateDate)
	d.Set("modify_date", *cert.ModifyDate)

	return nil
}

func resourceSoftLayerSecurityCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecurityCertificateService(sess)

	id, _ := strconv.Atoi(d.Id())
	_, err := service.Id(id).DeleteObject()

	if err != nil {
		return fmt.Errorf("Error deleting Security Certificate %s: %s", d.Id(), err)
	}

	return nil
}

func resourceSoftLayerSecurityCertificateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecurityCertificateService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	cert, err := service.Id(id).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); ok {
			if apiErr.StatusCode == 404 {
				return false, nil
			}
		}

		return false, fmt.Errorf("Error obtaining security certificate: %s", err)
	}

	return cert.Id != nil && *cert.Id == id, nil
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
