package softlayer

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerDnsDomain() *schema.Resource {
	return &schema.Resource{
		Exists:   resourceSoftLayerDnsDomainExists,
		Create:   resourceSoftLayerDnsDomainCreate,
		Read:     resourceSoftLayerDnsDomainRead,
		Update:   resourceSoftLayerDnsDomainUpdate,
		Delete:   resourceSoftLayerDnsDomainDelete,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"update_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"target": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerDnsDomainCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainService(sess)

	// prepare creation parameters
	opts := datatypes.Dns_Domain{
		Name: sl.String(d.Get("name").(string)),
	}

	// Create dns domain zone with default A record based on the target value
	opts.ResourceRecords = []datatypes.Dns_Domain_ResourceRecord{
		{
			Data: sl.String(d.Get("target").(string)),
			Host: sl.String("@"),
			Ttl:  sl.Int(86400),
			Type: sl.String("a"),
		},
	}

	// create Dns_Domain object
	response, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating Dns Domain: %s", err)
	}

	// populate id
	id := *response.Id
	d.SetId(strconv.Itoa(id))
	log.Printf("[INFO] Created Dns Domain: %d", id)

	// read remote state
	return resourceSoftLayerDnsDomainRead(d, meta)
}

func resourceSoftLayerDnsDomainRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainService(sess)

	dnsId, _ := strconv.Atoi(d.Id())

	// retrieve remote object state
	dns_domain, err := service.Id(dnsId).Mask(
		"id,name,updateDate,resourceRecords",
	).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving Dns Domain %d: %s", dnsId, err)
	}

	// populate fields
	d.Set("name", *dns_domain.Name)
	if dns_domain.Serial != nil {
		d.Set("serial", *dns_domain.Serial)
	}

	if dns_domain.UpdateDate != nil {
		d.Set("update_date", *dns_domain.UpdateDate)
	}

	// find a record with host @; that will have the current target.
	for _, record := range dns_domain.ResourceRecords {
		if *record.Type == "a" && *record.Host == "@" {
			d.Set("target", *record.Data)
			break
		}
	}

	return nil
}

func resourceSoftLayerDnsDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	// If the target has been updated, find the corresponding dns record and update its data.
	sess := meta.(*session.Session)
	domainId, _ := strconv.Atoi(d.Id())

	if !d.HasChange("target") {
		return nil
	}

	newTarget := d.Get("target").(string)

	// retrieve domain state
	domainService := services.GetDnsDomainService(sess)
	domain, err := domainService.Id(domainId).Mask(
		"id,name,updateDate,resourceRecords",
	).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving DNS resource %d: %s", domainId, err)
	}

	// find a record with host @; that will have the current target.
	var record datatypes.Dns_Domain_ResourceRecord
	for _, record = range domain.ResourceRecords {
		if *record.Type == "a" && *record.Host == "@" {
			break
		}
	}

	if record.Id == nil {
		return fmt.Errorf("Could not find DNS target record for domain %s (%d)",
			sl.Get(domain.Name), sl.Get(domain.Id))
	}

	record.Data = sl.String(newTarget)

	_, err = services.GetDnsDomainResourceRecordService(sess).
		Id(*record.Id).EditObject(&record)

	if err != nil {
		return fmt.Errorf("Error editing DNS target record for domain %s (%d): %s",
			sl.Get(domain.Name), sl.Get(domain.Id), err)
	}

	return nil
}

func resourceSoftLayerDnsDomainDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainService(sess)

	dnsId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Dns Domain: %s", err)
	}

	log.Printf("[INFO] Deleting Dns Domain: %d", dnsId)
	result, err := service.Id(dnsId).DeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting Dns Domain: %s", err)
	}

	if !result {
		return errors.New("Error deleting Dns Domain")
	}

	d.SetId("")
	return nil
}

func resourceSoftLayerDnsDomainExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainService(sess)

	dnsId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(dnsId).GetObject()
	return err == nil && result.Id != nil && *result.Id == dnsId, nil
}
