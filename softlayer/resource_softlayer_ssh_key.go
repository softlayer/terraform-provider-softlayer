package softlayer

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
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
			"label": {
				Type:     schema.TypeString,
				Required: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},

			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"notes": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceSoftLayerSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecuritySshKeyService(sess)

	// First check if the key exits by fingerprint
	// If so, set the Id (and fingerprint), but update notes and label (if any)
	key := d.Get("public_key").(string)
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return errors.New("Invalid public key specified")
	}

	k, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("Error decoding the public key: %s", err)
	}

	fp := md5.Sum([]byte(k))
	prints := make([]string, len(fp))
	for i, b := range fp {
		prints[i] = fmt.Sprintf("%02x", b)
	}
	fingerprint := strings.Join(prints, ":")

	keys, err := services.GetAccountService(sess).
		Filter(filter.Path("sshKeys.fingerprint").Eq(fingerprint).Build()).
		GetSshKeys()
	if err == nil && len(keys) > 0 {
		slKey := keys[0]
		id := *slKey.Id
		slKey.Id = nil
		d.SetId(fmt.Sprintf("%d", id))
		d.Set("fingerprint", fingerprint)
		editKey := false

		notes := d.Get("notes").(string)
		if notes != "" && (slKey.Notes == nil || notes != *slKey.Notes) {
			slKey.Notes = sl.String(notes)
			editKey = true
		} else if slKey.Notes != nil {
			d.Set("notes", *slKey.Notes)
		}

		label := d.Get("label").(string)
		if label != *slKey.Label {
			slKey.Label = sl.String(d.Get("label").(string))
			editKey = true
		}

		if editKey {
			_, err = service.Id(id).EditObject(&slKey)
			return err
		}

		return nil
	} // End of "Import"

	// Build up our creation options
	opts := datatypes.Security_Ssh_Key{
		Label: sl.String(d.Get("label").(string)),
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
	sess := meta.(ProviderConfig).SoftLayerSession()
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

	d.SetId(strconv.Itoa(*key.Id))
	d.Set("label", *key.Label)
	d.Set("public_key", *key.Key)
	d.Set("fingerprint", *key.Fingerprint)

	if key.Notes != nil {
		d.Set("notes", *key.Notes)
	}

	return nil
}

func resourceSoftLayerSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecuritySshKeyService(sess)

	keyId, _ := strconv.Atoi(d.Id())

	key, err := service.Id(keyId).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving SSH key: %s", err)
	}

	if d.HasChange("label") {
		key.Label = sl.String(d.Get("label").(string))
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
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecuritySshKeyService(sess)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting SSH Key: %s", err)
	}

	log.Printf("[INFO] Deleting SSH key: %d", id)
	tries := 0
	for {
		_, err = service.Id(id).DeleteObject()
		if err != nil {
			if strings.Contains(err.Error(), "it is currently being used") && tries < 5 {
				log.Printf("SSH key %d is still being used. Waiting to delete...\n", id)
				time.Sleep(1 * time.Minute)
				tries = tries + 1
				continue
			}
		}
		break
	}

	d.SetId("")
	return err
}

func resourceSoftLayerSSHKeyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetSecuritySshKeyService(sess)

	keyId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(keyId).GetObject()
	return result.Id != nil && err == nil && *result.Id == keyId, nil
}
