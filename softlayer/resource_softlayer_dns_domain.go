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
				Type:     schema.TypeInt,
				Computed: true,
			},

			"update_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"records": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"record_data": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"domain_id": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},

						"expire": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"host": {
							Type:     schema.TypeString,
							Required: true,
						},

						"minimum_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"mx_priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"refresh": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"contact_email": {
							Type:     schema.TypeString,
							Required: true,
						},

						"retry": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"ttl": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"record_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"service": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
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

	if records, ok := d.GetOk("records"); ok {
		opts.ResourceRecords = prepareRecords(records.([]interface{}))
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

func prepareRecords(raw_records []interface{}) []datatypes.Dns_Domain_ResourceRecord {
	sl_records := make([]datatypes.Dns_Domain_ResourceRecord, 0)
	for _, raw_record := range raw_records {
		var sl_record datatypes.Dns_Domain_ResourceRecord
		record := raw_record.(map[string]interface{})

		sl_record.Data = sl.String(record["record_data"].(string))
		sl_record.DomainId = sl.Int(record["domain_id"].(int))
		sl_record.Expire = sl.Int(record["expire"].(int))
		sl_record.Host = sl.String(record["host"].(string))
		sl_record.Minimum = sl.Int(record["minimum_ttl"].(int))
		sl_record.MxPriority = sl.Int(record["mx_priority"].(int))
		sl_record.Refresh = sl.Int(record["refresh"].(int))
		sl_record.ResponsiblePerson = sl.String(record["contact_email"].(string))
		sl_record.Retry = sl.Int(record["retry"].(int))
		sl_record.Ttl = sl.Int(record["ttl"].(int))
		sl_record.Type = sl.String(record["record_type"].(string))

		sl_records = append(sl_records, sl_record)
	}

	return sl_records
}

func resourceSoftLayerDnsDomainRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainService(sess)

	dnsId, _ := strconv.Atoi(d.Id())

	// retrieve remote object state
	dns_domain, err := service.Id(dnsId).Mask(
		"id,name,serial,updateDate," +
			"resourceRecords[data,domainId,expire,host,minimum,mxPriority,refresh,retry,ttl,type]",
	).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving Dns Domain %d: %s", dnsId, err)
	}

	// populate fields
	d.Set("id", *dns_domain.Id)
	d.Set("name", *dns_domain.Name)
	d.Set("serial", *dns_domain.Serial)
	d.Set("update_date", *dns_domain.UpdateDate)
	d.Set("records", read_resource_records(dns_domain.ResourceRecords))

	return nil
}

func read_resource_records(list []datatypes.Dns_Domain_ResourceRecord) []map[string]interface{} {
	records := make([]map[string]interface{}, 0, len(list))
	for _, record := range list {
		r := make(map[string]interface{})
		r["record_data"] = *record.Data
		r["domain_id"] = *record.DomainId
		r["expire"] = *record.Expire
		r["host"] = *record.Host
		r["minimum_ttl"] = *record.Minimum
		r["mx_priority"] = *record.MxPriority
		r["refresh"] = *record.Refresh
		r["retry"] = *record.Retry
		r["ttl"] = *record.Ttl
		r["record_type"] = *record.Type
		records = append(records, r)
	}
	return records
}

func resourceSoftLayerDnsDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO - update is not supported - implement delete-create?
	return errors.New("Not implemented. Update Dns Domain is currently unsupported")
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
	return result.Id != nil && err == nil && *result.Id == dnsId, nil
}
