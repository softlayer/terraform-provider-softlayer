package softlayer

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSoftLayerProvisioningHook() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerProvisioningHookCreate,
		Read:   resourceSoftLayerProvisioningHookRead,
		Update: resourceSoftLayerProvisioningHookUpdate,
		Delete: resourceSoftLayerProvisioningHookDelete,
		Exists: resourceSoftLayerProvisioningHookExists,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

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
	client := meta.(*Client).provisioningHookService

	opts := datatypes.SoftLayer_Provisioning_Hook_Template{
		Name: d.Get("name").(string),
		Uri:  d.Get("uri").(string),
	}

	hook, err := client.CreateObject(opts)
	if err != nil {
		return fmt.Errorf("Error creating Provisioning Hook: %s", err)
	}

	d.SetId(strconv.Itoa(hook.Id))
	log.Printf("[INFO] Provisioning Hook ID: %d", hook.Id)

	return resourceSoftLayerProvisioningHookRead(d, meta)
}

func resourceSoftLayerProvisioningHookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).provisioningHookService

	hookId, _ := strconv.Atoi(d.Id())

	hook, err := client.GetObject(hookId)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Provisioning Hook: %s", err)
	}

	d.Set("id", hook.Id)
	d.Set("name", hook.Name)
	d.Set("typeId", hook.TypeId)
	d.Set("uri", hook.Uri)

	return nil
}

func resourceSoftLayerProvisioningHookUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).provisioningHookService

	hookId, _ := strconv.Atoi(d.Id())

	//hook, err := client.GetObject(hookId)
	//if err != nil {
	//      return fmt.Errorf("Error retrieving Provisioning Hook: %s", err)
	//}

	opts := datatypes.SoftLayer_Provisioning_Hook_Template{
	//      TypeId: d.Get("typeId").(int),
	}

	if d.HasChange("name") {
		opts.Name = d.Get("name").(string)
	}

	if d.HasChange("uri") {
		opts.Uri = d.Get("uri").(string)
	}

	_, err := client.EditObject(hookId, opts)

	if err != nil {
		return fmt.Errorf("Error editing Provisioning Hook: %s", err)
	}
	return nil
}

func resourceSoftLayerProvisioningHookDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).provisioningHookService

	hookId, err := strconv.Atoi(d.Id())
	log.Printf("[INFO] Deleting Provisioning Hook: %d", hookId)
	_, err = client.DeleteObject(hookId)
	if err != nil {
		return fmt.Errorf("Error deleting Provisioning Hook: %s", err)
	}

	return nil
}

func resourceSoftLayerProvisioningHookExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client).provisioningHookService

	if client == nil {
		return false, fmt.Errorf("The client was nil.")
	}

	hookId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := client.GetObject(hookId)
	return result.Id == hookId && err == nil, nil
}
