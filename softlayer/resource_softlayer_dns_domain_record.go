package softlayer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerDnsDomainRecord() *schema.Resource {
	return &schema.Resource{
		Exists:   resourceSoftLayerDnsDomainRecordExists,
		Create:   resourceSoftLayerDnsDomainRecordCreate,
		Read:     resourceSoftLayerDnsDomainRecordRead,
		Update:   resourceSoftLayerDnsDomainRecordUpdate,
		Delete:   resourceSoftLayerDnsDomainRecordDelete,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"data": {
				Type:     schema.TypeString,
				Required: true,
			},

			"domain_id": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"expire": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"host": {
				Type:     schema.TypeString,
				Required: true,
			},

			"mx_priority": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"refresh": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"responsible_person": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"retry": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"minimum_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return 86400, nil
				},
			},

			"type": {
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
				Computed: true,
			},

			"weight": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

//  Creates DNS Domain Resource Record
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/createObject
func resourceSoftLayerDnsDomainRecordCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	opts := datatypes.Dns_Domain_ResourceRecord{
		Data:              sl.String(d.Get("data").(string)),
		DomainId:          sl.Int(d.Get("domain_id").(int)),
		Host:              sl.String(d.Get("host").(string)),
		Ttl:               sl.Int(d.Get("ttl").(int)),
		Type:              sl.String(d.Get("type").(string)),
	}

	if expire, ok := d.GetOk("expire"); ok {
		opts.Expire = sl.Int(expire.(int))
	}

	if minimum, ok := d.GetOk("minimum_ttl"); ok {
		opts.Minimum = sl.Int(minimum.(int))
	}

	if mxPriority, ok := d.GetOk("mx_priority"); ok {
		opts.MxPriority = sl.Int(mxPriority.(int))
	}

	if refresh, ok := d.GetOk("refresh"); ok {
		opts.Refresh = sl.Int(refresh.(int))
	}

	if responsiblePerson, ok := d.GetOk("responsible_person"); ok {
		opts.ResponsiblePerson = sl.String(responsiblePerson.(string))
	}

	if retry, ok := d.GetOk("retry"); ok {
		opts.Retry = sl.Int(retry.(int))
	}

	if *opts.Type == "srv" {
		if serviceName, ok := d.GetOk("service"); ok {
			opts.Service = sl.String(serviceName.(string))
		}

		if protocol, ok := d.GetOk("protocol"); ok {
			opts.Protocol = sl.String(protocol.(string))
		}

		if priority, ok := d.GetOk("priority"); ok {
			opts.Priority = sl.Int(priority.(int))
		}

		if weight, ok := d.GetOk("weight"); ok {
			opts.Weight = sl.Int(weight.(int))
		}

		if port, ok := d.GetOk("port"); ok {
			opts.Port = sl.Int(port.(int))
		}
	}

	log.Printf("[INFO] Creating DNS Resource %s Record for '%d' dns domain", *opts.Type, d.Get("id"))
	record, err := service.CreateObject(&opts)

	if err != nil {
		return fmt.Errorf("Error creating DNS Resource %s Record: %s", *opts.Type, err)
	}

	d.SetId(fmt.Sprintf("%d", *record.Id))

	log.Printf("[INFO] Dns Resource %s Record ID: %s", *opts.Type, d.Id())

	return resourceSoftLayerDnsDomainRecordRead(d, meta)
}

//  Reads DNS Domain Resource Record from SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/getObject
func resourceSoftLayerDnsDomainRecordRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}
	result, err := service.Id(id).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving DNS Resource Record: %s", err)
	}

	d.Set("data", *result.Data)
	d.Set("domain_id", *result.DomainId)
	d.Set("host", *result.Host)
	d.Set("type", *result.Type)
	d.Set("ttl", *result.Ttl)

	if result.Expire != nil {
		d.Set("expire", *result.Expire)
	}

	if result.Minimum != nil {
		d.Set("minimum_ttl", *result.Minimum)
	}

	if result.MxPriority != nil {
		d.Set("mx_priority", *result.MxPriority)
	}

	if result.Refresh != nil {
		d.Set("refresh", *result.Refresh)
	}

	if result.ResponsiblePerson != nil {
		d.Set("responsible_person", *result.ResponsiblePerson)
	}

	if result.Retry != nil {
		d.Set("retry", *result.Retry)
	}

	if *result.Type == "srv" {
		if result.Service != nil {
			d.Set("service", *result.Service)
		}

		if result.Protocol != nil {
			d.Set("protocol", *result.Protocol)
		}

		if result.Port != nil {
			d.Set("port", *result.Port)
		}

		if result.Priority != nil {
			d.Set("priority", *result.Priority)
		}

		if result.Weight != nil {
			d.Set("weight", *result.Weight)
		}
	}

	return nil
}

//  Updates DNS Domain Resource Record in SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/editObject
func resourceSoftLayerDnsDomainRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	recordId, _ := strconv.Atoi(d.Id())

	service := services.GetDnsDomainResourceRecordService(sess)
	record, err := service.Id(recordId).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving DNS Resource Record: %s", err)
	}

	recordType := d.Get("type").(string)
	hasChanged := false

	if data, ok := d.GetOk("data"); ok {
		record.Data = sl.String(data.(string))
		hasChanged = hasChanged || d.HasChange("data")
	}

	if domain_id, ok := d.GetOk("domain_id"); ok {
		record.DomainId = sl.Int(domain_id.(int))
		hasChanged = hasChanged || d.HasChange("domain_id")
	}

	if host, ok := d.GetOk("host"); ok {
		record.Host = sl.String(host.(string))
		hasChanged = hasChanged || d.HasChange("host")
	}

	if ttl, ok := d.GetOk("ttl"); ok {
		record.Ttl = sl.Int(ttl.(int))
		hasChanged = hasChanged || d.HasChange("ttl")
	}

	if expire, ok := d.GetOk("expire"); ok {
		record.Expire = sl.Int(expire.(int))
		hasChanged = hasChanged || d.HasChange("expire")
	}

	if minimum_ttl, ok := d.GetOk("minimum_ttl"); ok {
		record.Minimum = sl.Int(minimum_ttl.(int))
		hasChanged = hasChanged || d.HasChange("minimum_ttl")
	}

	if mx_priority, ok := d.GetOk("mx_priority"); ok {
		record.MxPriority = sl.Int(mx_priority.(int))
		hasChanged = hasChanged || d.HasChange("mx_priority")
	}

	if refresh, ok := d.GetOk("refresh"); ok {
		record.Refresh = sl.Int(refresh.(int))
		hasChanged = hasChanged || d.HasChange("refresh")
	}

	if contact_email, ok := d.GetOk("responsible_person"); ok {
		record.ResponsiblePerson = sl.String(contact_email.(string))
		hasChanged = hasChanged || d.HasChange("responsible_person")
	}

	if retry, ok := d.GetOk("retry"); ok {
		record.Retry = sl.Int(retry.(int))
		hasChanged = hasChanged || d.HasChange("retry")
	}

	if recordType == "srv" {
		if service, ok := d.GetOk("service"); ok {
			record.Service = sl.String(service.(string))
			hasChanged = hasChanged || d.HasChange("service")
		}

		if priority, ok := d.GetOk("priority"); ok {
			record.Priority = sl.Int(priority.(int))
			hasChanged = hasChanged || d.HasChange("priority")
		}

		if protocol, ok := d.GetOk("protocol"); ok {
			record.Protocol = sl.String(protocol.(string))
			hasChanged = hasChanged || d.HasChange("protocol")
		}

		if port, ok := d.GetOk("port"); ok {
			record.Port = sl.Int(port.(int))
			hasChanged = hasChanged || d.HasChange("port")
		}

		if weight, ok := d.GetOk("weight"); ok {
			record.Weight = sl.Int(weight.(int))
			hasChanged = hasChanged || d.HasChange("weight")
		}
	}

	if hasChanged {
		_, err = service.Id(recordId).EditObject(&record)
		if err != nil {
			return fmt.Errorf("Error editing DNS Resoource Record: %s", err)
		}
	}

	return nil
}

//  Deletes DNS Domain Resource Record in SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/deleteObject
func resourceSoftLayerDnsDomainRecordDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = service.Id(id).DeleteObject()

	if err != nil {
		return fmt.Errorf("Error deleting DNS Resource Record: %s", err)
	}

	return nil
}

// Exists function is called by refresh
// if the entity is absent - it is deleted from the .tfstate file
func resourceSoftLayerDnsDomainRecordExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	record, err := service.Id(id).GetObject()

	return err == nil && record.Id != nil && *record.Id == id, nil
}
