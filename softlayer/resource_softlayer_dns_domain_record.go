package softlayer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.ibm.com/riethm/gopherlayer.git/datatypes"
	"github.ibm.com/riethm/gopherlayer.git/services"
	"github.ibm.com/riethm/gopherlayer.git/session"
	"github.ibm.com/riethm/gopherlayer.git/sl"
)

func resourceSoftLayerDnsDomainResourceRecord() *schema.Resource {
	return &schema.Resource{
		Exists:   resourceSoftLayerDnsDomainResourceRecordExists,
		Create:   resourceSoftLayerDnsDomainResourceRecordCreate,
		Read:     resourceSoftLayerDnsDomainResourceRecordRead,
		Update:   resourceSoftLayerDnsDomainResourceRecordUpdate,
		Delete:   resourceSoftLayerDnsDomainResourceRecordDelete,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"record_data": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"domain_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"expire": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"host": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"minimum_ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"mx_priority": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"refresh": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"contact_email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"retry": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},

			"record_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"priority": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"weight": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

//  Creates DNS Domain Resource Record
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/createObject
func resourceSoftLayerDnsDomainResourceRecordCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	opts := datatypes.Dns_Domain_ResourceRecord{
		Data:              sl.String(d.Get("record_data").(string)),
		DomainId:          sl.Int(d.Get("domain_id").(int)),
		Expire:            sl.Int(d.Get("expire").(int)),
		Host:              sl.String(d.Get("host").(string)),
		Minimum:           sl.Int(d.Get("minimum_ttl").(int)),
		MxPriority:        sl.Int(d.Get("mx_priority").(int)),
		Refresh:           sl.Int(d.Get("refresh").(int)),
		ResponsiblePerson: sl.String(d.Get("contact_email").(string)),
		Retry:             sl.Int(d.Get("retry").(int)),
		Ttl:               sl.Int(d.Get("ttl").(int)),
		Type:              sl.String(d.Get("record_type").(string)),
	}

	if *opts.Type == "srv" {
		opts_srv := datatypes.Dns_Domain_ResourceRecord_SrvType{
			Dns_Domain_ResourceRecord: opts,
			Service:                   sl.String(d.Get("service").(string)),
			Protocol:                  sl.String(d.Get("protocol").(string)),
			Priority:                  sl.Int(d.Get("priority").(int)),
			Weight:                    sl.Int(d.Get("weight").(int)),
			Port:                      sl.Int(d.Get("port").(int)),
		}

		service_srv := services.GetDnsDomainResourceRecordSrvTypeService(sess)
		log.Printf("[INFO] Creating DNS Resource SRV Record for '%d' dns domain", d.Get("id"))
		record, err := service_srv.CreateObject(&opts_srv)
		if err != nil {
			return fmt.Errorf("Error creating DNS Resource SRV Record: %s", err)
		}

		d.SetId(fmt.Sprintf("%d", *record.Id))
		log.Printf("[INFO] Dns Resource SRV Record ID: %s", d.Id())
	} else {
		log.Printf("[INFO] Creating DNS Resource Record for '%d' dns domain", d.Get("id"))
		record, err := service.CreateObject(&opts)

		if err != nil {
			return fmt.Errorf("Error creating DNS Resource Record: %s", err)
		}

		d.SetId(fmt.Sprintf("%d", *record.Id))

		log.Printf("[INFO] Dns Resource Record ID: %s", d.Id())
	}

	return resourceSoftLayerDnsDomainResourceRecordRead(d, meta)
}

//  Reads DNS Domain Resource Record from SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/getObject
func resourceSoftLayerDnsDomainResourceRecordRead(d *schema.ResourceData, meta interface{}) error {
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
	d.Set("domainId", *result.DomainId)
	d.Set("expire", *result.Expire)
	d.Set("host", *result.Host)
	d.Set("id", *result.Id)
	d.Set("minimum", *result.Minimum)
	d.Set("mxPriority", *result.MxPriority)
	d.Set("refresh", *result.Refresh)
	d.Set("responsiblePerson", *result.ResponsiblePerson)
	d.Set("retry", *result.Retry)
	d.Set("ttl", *result.Ttl)
	d.Set("type", *result.Type)

	if *result.Type == "srv" {
		service := services.GetDnsDomainResourceRecordSrvTypeService(sess)
		result, err := service.Id(id).GetObject()
		if err != nil {
			return fmt.Errorf("Error retrieving DNS Resource SRV Record: %s", err)
		}

		d.Set("service", *result.Service)
		d.Set("protocol", *result.Protocol)
		d.Set("port", *result.Port)
		d.Set("priority", *result.Priority)
		d.Set("weight", *result.Weight)
	}

	return nil
}

//  Updates DNS Domain Resource Record in SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/editObject
func resourceSoftLayerDnsDomainResourceRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	recordId, _ := strconv.Atoi(d.Id())

	if d.Get("type").(string) == "srv" {
		service := services.GetDnsDomainResourceRecordSrvTypeService(sess)
		record, err := service.Id(recordId).GetObject()
		if err != nil {
			return fmt.Errorf("Error retrieving DNS Resource SRV Record: %s", err)
		}

		if data, ok := d.GetOk("record_data"); ok {
			record.Data = sl.String(data.(string))
		}
		if domain_id, ok := d.GetOk("domain_id"); ok {
			record.DomainId = sl.Int(domain_id.(int))
		}
		if expire, ok := d.GetOk("expire"); ok {
			record.Expire = sl.Int(expire.(int))
		}
		if host, ok := d.GetOk("host"); ok {
			record.Host = sl.String(host.(string))
		}
		if minimum_ttl, ok := d.GetOk("minimum_ttl"); ok {
			record.Minimum = sl.Int(minimum_ttl.(int))
		}
		if mx_priority, ok := d.GetOk("mx_priority"); ok {
			record.MxPriority = sl.Int(mx_priority.(int))
		}
		if refresh, ok := d.GetOk("refresh"); ok {
			record.Refresh = sl.Int(refresh.(int))
		}
		if contact_email, ok := d.GetOk("contact_email"); ok {
			record.ResponsiblePerson = sl.String(contact_email.(string))
		}
		if retry, ok := d.GetOk("retry"); ok {
			record.Retry = sl.Int(retry.(int))
		}
		if ttl, ok := d.GetOk("ttl"); ok {
			record.Ttl = sl.Int(ttl.(int))
		}
		if record_type, ok := d.GetOk("record_type"); ok {
			record.Type = sl.String(record_type.(string))
		}
		if service, ok := d.GetOk("service"); ok {
			record.Service = sl.String(service.(string))
		}
		if priority, ok := d.GetOk("priority"); ok {
			record.Priority = sl.Int(priority.(int))
		}
		if protocol, ok := d.GetOk("protocol"); ok {
			record.Protocol = sl.String(protocol.(string))
		}
		if port, ok := d.GetOk("port"); ok {
			record.Port = sl.Int(port.(int))
		}
		if weight, ok := d.GetOk("weight"); ok {
			record.Weight = sl.Int(weight.(int))
		}

		_, err = service.Id(recordId).EditObject(&record)
		if err != nil {
			return fmt.Errorf("Error editing DNS Resoource SRV Record: %s", err)
		}
	} else {
		service := services.GetDnsDomainResourceRecordService(sess)
		record, err := service.Id(recordId).GetObject()
		if err != nil {
			return fmt.Errorf("Error retrieving DNS Resource Record: %s", err)
		}

		if data, ok := d.GetOk("record_data"); ok {
			record.Data = sl.String(data.(string))
		}
		if domain_id, ok := d.GetOk("domain_id"); ok {
			record.DomainId = sl.Int(domain_id.(int))
		}
		if expire, ok := d.GetOk("expire"); ok {
			record.Expire = sl.Int(expire.(int))
		}
		if host, ok := d.GetOk("host"); ok {
			record.Host = sl.String(host.(string))
		}
		if minimum_ttl, ok := d.GetOk("minimum_ttl"); ok {
			record.Minimum = sl.Int(minimum_ttl.(int))
		}
		if mx_priority, ok := d.GetOk("mx_priority"); ok {
			record.MxPriority = sl.Int(mx_priority.(int))
		}
		if refresh, ok := d.GetOk("refresh"); ok {
			record.Refresh = sl.Int(refresh.(int))
		}
		if contact_email, ok := d.GetOk("contact_email"); ok {
			record.ResponsiblePerson = sl.String(contact_email.(string))
		}
		if retry, ok := d.GetOk("retry"); ok {
			record.Retry = sl.Int(retry.(int))
		}
		if ttl, ok := d.GetOk("ttl"); ok {
			record.Ttl = sl.Int(ttl.(int))
		}
		if record_type, ok := d.GetOk("record_type"); ok {
			record.Type = sl.String(record_type.(string))
		}

		_, err = service.Id(recordId).EditObject(&record)
		if err != nil {
			return fmt.Errorf("Error editing DNS Resoource Record: %s", err)
		}
	}

	return nil
}

//  Deletes DNS Domain Resource Record in SL system
//  https://sldn.softlayer.com/reference/services/SoftLayer_Dns_Domain_ResourceRecord/deleteObject
func resourceSoftLayerDnsDomainResourceRecordDelete(d *schema.ResourceData, meta interface{}) error {
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
func resourceSoftLayerDnsDomainResourceRecordExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetDnsDomainResourceRecordService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	record, err := service.Id(id).GetObject()

	return err == nil && *record.Id == id, nil
}
