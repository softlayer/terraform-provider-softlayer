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

func resourceSoftLayerProvisioningHook() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerProvisioningHookCreate,
		Read:     resourceSoftLayerProvisioningHookRead,
		Update:   resourceSoftLayerProvisioningHookUpdate,
		Delete:   resourceSoftLayerProvisioningHookDelete,
		Exists:   resourceSoftLayerProvisioningHookExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"uri": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerProvisioningHookCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetProvisioningHookService(sess)

	opts := datatypes.Provisioning_Hook{
		Name: sl.String(d.Get("name").(string)),
		Uri:  sl.String(d.Get("uri").(string)),
	}

	hook, err := service.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating Provisioning Hook: %s", err)
	}

	d.SetId(strconv.Itoa(*hook.Id))
	log.Printf("[INFO] Provisioning Hook ID: %d", *hook.Id)

	return resourceSoftLayerProvisioningHookRead(d, meta)
}

func resourceSoftLayerProvisioningHookRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetProvisioningHookService(sess)

	hookId, _ := strconv.Atoi(d.Id())

	hook, err := service.Id(hookId).GetObject()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Provisioning Hook: %s", err)
	}

	d.SetId(strconv.Itoa(*hook.Id))
	d.Set("name", *hook.Name)
	d.Set("uri", *hook.Uri)

	return nil
}

func resourceSoftLayerProvisioningHookUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetProvisioningHookService(sess)

	hookId, _ := strconv.Atoi(d.Id())

	opts := datatypes.Provisioning_Hook{}

	if d.HasChange("name") {
		opts.Name = sl.String(d.Get("name").(string))
	}

	if d.HasChange("uri") {
		opts.Uri = sl.String(d.Get("uri").(string))
	}

	_, err := service.Id(hookId).EditObject(&opts)

	if err != nil {
		return fmt.Errorf("Error editing Provisioning Hook: %s", err)
	}
	return nil
}

func resourceSoftLayerProvisioningHookDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetProvisioningHookService(sess)

	hookId, err := strconv.Atoi(d.Id())
	log.Printf("[INFO] Deleting Provisioning Hook: %d", hookId)
	_, err = service.Id(hookId).DeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting Provisioning Hook: %s", err)
	}

	return nil
}

func resourceSoftLayerProvisioningHookExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetProvisioningHookService(sess)

	hookId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(hookId).GetObject()
	if err != nil {
		if apiErr, ok := err.(sl.Error); ok {
			if apiErr.StatusCode == 404 {
				return false, nil
			}
		}

		return false, fmt.Errorf("Unable to get provisioning hook: %s", err)
	}

	return result.Id != nil && *result.Id == hookId, nil
}
