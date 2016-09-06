package softlayer

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func resourceSoftLayerSSHKey() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerSSHKeyCreate,
		Read:     resourceSoftLayerSSHKeyRead,
		Update:   resourceSoftLayerSSHKeyUpdate,
		Delete:   resourceSoftLayerSSHKeyDelete,
		Exists:   resourceSoftLayerSSHKeyExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"fingerprint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceSoftLayerSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetSecuritySshKeyService(sess)

	// Build up our creation options
	opts := datatypes.Security_Ssh_Key{
		Label: sl.String(d.Get("name").(string)),
		Key:   sl.String(d.Get("public_key").(string)),
	}

	if notes, ok := d.GetOk("notes"); ok {
		opts.Notes = sl.String(notes.(string))
	}

	res, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating SSH Key: %s", err)
	}

	d.SetId(strconv.Itoa(*res.Id))
	log.Printf("[INFO] SSH Key: %d", *res.Id)

	return resourceSoftLayerSSHKeyRead(d, meta)
}

func resourceSoftLayerSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetSecuritySshKeyService(sess)

	keyId, _ := strconv.Atoi(d.Id())

	key, err := service.Id(keyId).GetObject()
	if err != nil {
		// If the key is somehow already destroyed, mark as
		// succesfully gone
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving SSH key: %s", err)
	}

	d.Set("id", *key.Id)
	d.Set("name", *key.Label)
	d.Set("public_key", strings.TrimSpace(*key.Key))
	d.Set("fingerprint", *key.Fingerprint)

	if key.Notes != nil {
		d.Set("notes", *key.Notes)
	}

	return nil
}

func resourceSoftLayerSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetSecuritySshKeyService(sess)

	keyId, _ := strconv.Atoi(d.Id())

	key, err := service.Id(keyId).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving SSH key: %s", err)
	}

	if d.HasChange("name") {
		key.Label = sl.String(d.Get("name").(string))
	}

	if d.HasChange("notes") {
		key.Notes = sl.String(d.Get("notes").(string))
	}

	_, err = service.Id(keyId).EditObject(&key)
	if err != nil {
		return fmt.Errorf("Error editing SSH key: %s", err)
	}
	return nil
}

func resourceSoftLayerSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetSecuritySshKeyService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting SSH Key: %s", err)
	}

	log.Printf("[INFO] Deleting SSH key: %d", id)
	_, err = service.Id(id).DeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting SSH key: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceSoftLayerSSHKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetSecuritySshKeyService(sess)

	keyId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(keyId).GetObject()
	return result.Id != nil && err == nil && *result.Id == keyId, nil
}
