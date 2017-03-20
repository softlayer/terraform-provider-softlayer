package softlayer

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerDnsSecondary() *schema.Resource {
	return &schema.Resource{
		Exists:   resourceSoftLayerDnsSecondaryExists,
		Create:   resourceSoftLayerDnsSecondaryCreate,
		Read:     resourceSoftLayerDnsSecondaryRead,
		Update:   resourceSoftLayerDnsSecondaryUpdate,
		Delete:   resourceSoftLayerDnsSecondaryDelete,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"masterIpAddress": {
				Type:     schema.TypeString,
				Required: true,
			},

			"transferFrequency": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"zoneName": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"statusId": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"statusText": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSoftLayerDnsSecondaryCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetDnsSecondaryService(sess)

	// prepare creation parameters
	opts := datatypes.Dns_Secondary{
		MasterIpAddress:   sl.String(d.Get("masterIpAddress").(string)),
		TransferFrequency: sl.Int(d.Get("transferFrequency").(int)),
		ZoneName:          sl.String(d.Get("zoneName").(string)),
	}

	// create Dns_Secondary object
	response, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating Dns Secondary Zone: %s", err)
	}

	// populate id
	id := *response.Id
	d.SetId(strconv.Itoa(id))
	log.Printf("[INFO] Created Dns Secondary Zone: %d", id)

	// read remote state
	return resourceSoftLayerDnsSecondaryRead(d, meta)
}

func resourceSoftLayerDnsSecondaryRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetDnsSecondaryService(sess)

	dnsId, _ := strconv.Atoi(d.Id())

	// retrieve remote object state
	dns_domain_secondary, err := service.Id(dnsId).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving Dns Secondary Zone %d: %s", dnsId, err)
	}

	// populate fields
	d.Set("masterIpAddress", *dns_domain_secondary.MasterIpAddress)
	d.Set("transferFrequency", *dns_domain_secondary.TransferFrequency)
	d.Set("zoneName", *dns_domain_secondary.ZoneName)
	d.Set("statusId", *dns_domain_secondary.StatusId)
	d.Set("statusText", *dns_domain_secondary.StatusText)

	return nil
}

func resourceSoftLayerDnsSecondaryUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	domainId, _ := strconv.Atoi(d.Id())
	hasChange := false

	opts := datatypes.Dns_Secondary{}
	if d.HasChange("masterIpAddress") {
		opts.MasterIpAddress = sl.String(d.Get("masterIpAddress").(string))
		hasChange = true
	}

	if d.HasChange("transferFrequency") {
		opts.TransferFrequency = sl.Int(d.Get("transferFrequency").(int))
		hasChange = true
	}

	if hasChange {
		service := services.GetDnsSecondaryService(sess)
		_, err := service.Id(domainId).EditObject(&opts)

		if err != nil {
			return fmt.Errorf("Error editing DNS secondary zone (%d): %s", domainId, err)
		}
	}

	return nil
}

func resourceSoftLayerDnsSecondaryDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetDnsSecondaryService(sess)

	dnsId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Dns Secondary Zone: %s", err)
	}

	log.Printf("[INFO] Deleting Dns Secondary Zone: %d", dnsId)
	result, err := service.Id(dnsId).DeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting Dns Secondary Zone: %s", err)
	}

	if !result {
		return errors.New("Error deleting Dns Secondary Zone")
	}

	d.SetId("")
	return nil
}

func resourceSoftLayerDnsSecondaryExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetDnsSecondaryService(sess)

	dnsId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(dnsId).GetObject()
	return err == nil && result.Id != nil && *result.Id == dnsId, nil
}
