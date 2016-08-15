package softlayer

import (
	"fmt"
	"log"
	"strconv"
	"bytes"
	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"time"
	"regexp"
)

const SoftLayerTimeFormat = string("2006-01-02T15:04:05-07:00")

func resourceSoftLayerScalePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerScalePolicyCreate,
		Read:   resourceSoftLayerScalePolicyRead,
		Update: resourceSoftLayerScalePolicyUpdate,
		Delete: resourceSoftLayerScalePolicyDelete,
		Exists: resourceSoftLayerScalePolicyExists,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"scale_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"scale_amount": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"cooldown": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"scale_group_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"triggers": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						// Conditionally-required fields, based on value of "type"
						"watches": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
									"metric": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"operator": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"value": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"period": &schema.Schema{
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
							Set: resourceSoftLayerScalePolicyHandlerHash,
						},

						"date": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"schedule": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceSoftLayerScalePolicyTriggerHash,
			},
		},
	}
}

func resourceSoftLayerScalePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).scalePolicyService
	var err error

	// Build up creation options
	opts := datatypes.SoftLayer_Scale_Policy{
		Name:         d.Get("name").(string),
		ScaleGroupId: d.Get("scale_group_id").(int),
		Cooldown:     d.Get("cooldown").(int),
	}

	if opts.Cooldown <= 0 || opts.Cooldown > 864000 {
		return fmt.Errorf("Error retrieving scalePolicy: %s", "cooldown must be between 0 seconds and 10 days.")
	}

	opts.ScaleActions = []datatypes.SoftLayer_Scale_Policy_Action{{
		TypeId:    1,
		Amount:    d.Get("scale_amount").(int),
		ScaleType: d.Get("scale_type").(string),
	},
	}
	if opts.ScaleActions[0].Amount <= 0 {
		return fmt.Errorf("Error retrieving scalePolicy: %s", "scale_amount should be greater than 0.")
	}
	if opts.ScaleActions[0].ScaleType != "ABSOLUTE" && opts.ScaleActions[0].ScaleType != "RELATIVE" && opts.ScaleActions[0].ScaleType != "PERCENT" {
		return fmt.Errorf("Error retrieving scalePolicy: %s", "scale_type should be ABSOLUTE, RELATIVE, or PERCENT.")
	}

	if _, ok := d.GetOk("triggers"); ok {
		err = validateTriggerTypes(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}

		opts.OneTimeTriggers,err = prepareOneTimeTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}

		opts.RepeatingTriggers,err = prepareRepeatingTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}

		opts.ResourceUseTriggers,err = prepareResourceUseTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}
	}

	log.Printf("[INFO] Creating scale policy")
	res, err := client.CreateObject(opts)
	if err != nil {
		return fmt.Errorf("Error creating Scale Policy: %s $s", err)
	}

	d.SetId(strconv.Itoa(res.Id))
	log.Printf("[INFO] Scale Polocy: %d", res.Id)

	return resourceSoftLayerScalePolicyRead(d, meta)
}

func resourceSoftLayerScalePolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).scalePolicyService
	scalePolicyId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid scale policy ID, must be an integer: %s", err)
	}

	log.Printf("[INFO] Reading Scale Polocy: %d", scalePolicyId)
	scalePolicy, err := client.GetObject(scalePolicyId)
	if err != nil {
		return fmt.Errorf("Error retrieving Scale Policy: %s", err)
	}

	d.Set("name", scalePolicy.Name)
	d.Set("cooldown", scalePolicy.Cooldown)
	d.Set("scale_group_id", scalePolicy.ScaleGroupId)
	d.Set("scale_type", scalePolicy.ScaleActions[0].ScaleType)
	d.Set("scale_amount", scalePolicy.ScaleActions[0].Amount)

	triggers := make([]map[string]interface{}, 0)
	triggers = append(triggers, readOneTimeTriggers(scalePolicy.OneTimeTriggers)...)
	triggers = append(triggers, readRepeatingTriggers(scalePolicy.RepeatingTriggers)...)
	triggers = append(triggers, readResourceUseTriggers(scalePolicy.ResourceUseTriggers)...)

	d.Set("triggers", triggers)

	return nil
}

func resourceSoftLayerScalePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).scalePolicyService
	triggerClient := meta.(*Client).scalePolicyTriggerService

	scalePolicyId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid scale policy ID, must be an integer: %s", err)
	}

	scalePolicy, err := client.GetObject(scalePolicyId)
	if err != nil {
		return fmt.Errorf("Error retrieving scalePolicy: %s", err)
	}

	var template datatypes.SoftLayer_Scale_Policy

	template.Id, _ = strconv.Atoi(d.Id())

	if d.HasChange("name") {
		template.Name = d.Get("name").(string)
	}

	if d.HasChange("scale_type") || d.HasChange("scale_amount") {
		template.ScaleActions = []datatypes.SoftLayer_Scale_Policy_Action{{
			Id:     scalePolicy.ScaleActions[0].Id,
			TypeId: 1,
		}}
	}

	if d.HasChange("scale_type") {
		template.ScaleActions[0].ScaleType = d.Get("scale_type").(string)
		if template.ScaleActions[0].ScaleType != "ABSOLUTE" && template.ScaleActions[0].ScaleType != "RELATIVE" && template.ScaleActions[0].ScaleType != "PERCENT" {
			return fmt.Errorf("Error retrieving scalePolicy: %s", "scale_type should be ABSOLUTE, RELATIVE, or PERCENT.")
		}
	}

	if d.HasChange("scale_amount") {
		template.ScaleActions[0].Amount = d.Get("scale_amount").(int)
		if template.ScaleActions[0].Amount <= 0 {
			return fmt.Errorf("Error retrieving scalePolicy: %s", "scale_amount should be greater than 0.")
		}
	}

	if d.HasChange("cooldown") {
		template.Cooldown = d.Get("cooldown").(int)
		if template.Cooldown <= 0 || template.Cooldown > 864000 {
			return fmt.Errorf("Error retrieving scalePolicy: %s", "cooldown must be between 0 seconds and 10 days.")
		}
	}

	if _, ok := d.GetOk("triggers"); ok {
		template.OneTimeTriggers,err = prepareOneTimeTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}
		template.RepeatingTriggers,err = prepareRepeatingTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}
		template.ResourceUseTriggers,err = prepareResourceUseTriggers(d)
		if err != nil {
			return fmt.Errorf("Error retrieving scalePolicy: %s", err)
		}
	}

	for _, triggerList := range scalePolicy.Triggers {
		triggerClient.DeleteObject(triggerList.Id)
	}

	time.Sleep(60)
	log.Printf("[INFO] Updating scale policy: %d", scalePolicyId)
	_, err = client.EditObject(scalePolicyId, template)

	if err != nil {
		return fmt.Errorf("Error updating scalie policy: %s", err)
	}

	return nil
}

func resourceSoftLayerScalePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).scalePolicyService

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting scale policy: %s", err)
	}

	log.Printf("[INFO] Deleting scale policy: %d", id)
	_, err = client.DeleteObject(id)
	if err != nil {
		return fmt.Errorf("Error deleting scale policy: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceSoftLayerScalePolicyExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client).scalePolicyService

	if client == nil {
		return false, fmt.Errorf("The client was nil.")
	}

	policyId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := client.GetObject(policyId)
	return result.Id == policyId && err == nil, nil
}

func validateTriggerTypes(d *schema.ResourceData) error {
	triggerLists := d.Get("triggers").(*schema.Set).List()
	for _, triggerList := range triggerLists {
		trigger := triggerList.(map[string]interface{})
		trigger_type := trigger["type"].(string)
		if trigger_type != "ONE_TIME" && trigger_type != "REPEATING" && trigger_type != "RESOURCE_USE" {
			return fmt.Errorf("Invalid trigger type: %s", trigger_type)
		}
	}
	return nil
}

func prepareOneTimeTriggers(d *schema.ResourceData) ([]datatypes.SoftLayer_Scale_Policy_Trigger_OneTime, error) {
	triggerLists := d.Get("triggers").(*schema.Set).List()
	triggers := make([]datatypes.SoftLayer_Scale_Policy_Trigger_OneTime, 0)
	for _, triggerList := range triggerLists {
		trigger := triggerList.(map[string]interface{})

		if trigger["type"].(string) == "ONE_TIME" {
			var oneTimeTrigger datatypes.SoftLayer_Scale_Policy_Trigger_OneTime
			oneTimeTrigger.TypeId = datatypes.SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_ONE_TIME
			timeStampString := trigger["date"].(string)
			timeStamp, err := time.Parse(SoftLayerTimeFormat, timeStampString)
			if err != nil {
				return nil, err
			}

			// SoftLayer triggers only accept EST time zone
			isEST,_ := regexp.MatchString("-05:00$", timeStampString)
			if !isEST {
				return nil, fmt.Errorf("The time zone should be an EST(-05:00).")
			}

			oneTimeTrigger.Date = &timeStamp
			triggers = append(triggers, oneTimeTrigger)
		}
	}
	return triggers, nil
}

func prepareRepeatingTriggers(d *schema.ResourceData) ([]datatypes.SoftLayer_Scale_Policy_Trigger_Repeating, error) {
	triggerLists := d.Get("triggers").(*schema.Set).List()
	triggers := make([]datatypes.SoftLayer_Scale_Policy_Trigger_Repeating, 0)
	for _, triggerList := range triggerLists {
		trigger := triggerList.(map[string]interface{})

		if trigger["type"].(string) == "REPEATING" {
			var repeatingTrigger datatypes.SoftLayer_Scale_Policy_Trigger_Repeating
			repeatingTrigger.TypeId = datatypes.SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_REPEATING
			repeatingTrigger.Schedule = trigger["schedule"].(string)
			triggers = append(triggers, repeatingTrigger)
		}
	}
	return triggers, nil
}

func prepareResourceUseTriggers(d *schema.ResourceData) ([]datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse, error) {
	triggerLists := d.Get("triggers").(*schema.Set).List()
	triggers := make([]datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse, 0)
	for _, triggerList := range triggerLists {
		trigger := triggerList.(map[string]interface{})

		if trigger["type"].(string) == "RESOURCE_USE" {
			var resourceUseTrigger datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse
			var err error
			resourceUseTrigger.TypeId = datatypes.SOFTLAYER_SCALE_POLICY_TRIGGER_TYPE_ID_RESOURCE_USE
			resourceUseTrigger.Watches, err = prepareWatches(trigger["watches"].(*schema.Set))
			if err != nil {
				return nil, err
			}
			triggers = append(triggers, resourceUseTrigger)
		}
	}
	return triggers, nil
}

func prepareWatches(d *schema.Set) ([]datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch, error) {
	watchLists := d.List()
	watches := make([]datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch, 0)
	for _, watcheList := range watchLists {
		var watch datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch
		watchMap := watcheList.(map[string]interface{})

		watch.Metric = watchMap["metric"].(string)
		if watch.Metric != "host.cpu.percent" && watch.Metric != "host.network.backend.in.rate" && watch.Metric != "host.network.backend.out.rate" && watch.Metric != "host.network.frontend.in.rate" && watch.Metric != "host.network.frontend.out.rate"{
			return nil, fmt.Errorf("Invalid metric : %s", watch.Metric)
		}

		watch.Operator = watchMap["operator"].(string)
		if watch.Operator !=">" && watch.Operator !="<" {
			return nil, fmt.Errorf("Invalid operator : %s", watch.Operator)
		}

		watch.Period = watchMap["period"].(int)
		if watch.Period <= 0 {
			return nil, fmt.Errorf("period shoud be greater than 0.")
		}

		watch.Value = watchMap["value"].(string)

		// Autoscale only support EWMA algorithm.
		watch.Algorithm = "EWMA"

		watches = append(watches, watch)
	}
	return watches, nil
}

func readOneTimeTriggers(list []datatypes.SoftLayer_Scale_Policy_Trigger_OneTime) []map[string]interface{} {
	triggers := make([]map[string]interface{}, 0, len(list))
	for _, trigger := range list {
		t := make(map[string]interface{})
		t["id"] = trigger.Id
		t["type"] = "ONE_TIME"
		t["date"] = trigger.Date.Format(SoftLayerTimeFormat)
		triggers = append(triggers, t)
	}
	return triggers
}

func readRepeatingTriggers(list []datatypes.SoftLayer_Scale_Policy_Trigger_Repeating) []map[string]interface{} {
	triggers := make([]map[string]interface{}, 0, len(list))
	for _, trigger := range list {
		t := make(map[string]interface{})
		t["id"] = trigger.Id
		t["type"] = "REPEATING"
		t["schedule"] = trigger.Schedule
		triggers = append(triggers, t)
	}
	return triggers
}

func readResourceUseTriggers(list []datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse) []map[string]interface{} {
	triggers := make([]map[string]interface{}, 0, len(list))
	for _, trigger := range list {
		t := make(map[string]interface{})
		t["id"] = trigger.Id
		t["type"] = "RESOURCE_USE"
		t["watches"] = readResourceUseWatches(trigger.Watches)
		triggers = append(triggers, t)
	}
	return triggers
}

func readResourceUseWatches(list []datatypes.SoftLayer_Scale_Policy_Trigger_ResourceUse_Watch) []map[string]interface{} {
	watches := make([]map[string]interface{}, 0, len(list))
	for _, watch := range list {
		w := make(map[string]interface{})
		w["id"] = watch.Id
		w["metric"] = watch.Metric
		w["operator"] = watch.Operator
		w["period"] = watch.Period
		w["value"] = watch.Value
		watches = append(watches, w)
	}
	return watches
}

func resourceSoftLayerScalePolicyTriggerHash(v interface{}) int {
	var buf bytes.Buffer
	trigger := v.(map[string]interface{})
	if trigger["type"].(string) == "ONE_TIME" {
		buf.WriteString(fmt.Sprintf("%s-", trigger["type"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", trigger["date"].(string)))
	}
	if trigger["type"].(string) == "REPEATING" {
		buf.WriteString(fmt.Sprintf("%s-", trigger["type"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", trigger["schedule"].(string)))
	}
	if trigger["type"].(string) == "RESOURCE_USE" {
		buf.WriteString(fmt.Sprintf("%s-", trigger["type"].(string)))
		for _, watchList := range trigger["watches"].(*schema.Set).List() {
			watch := watchList.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", watch["metric"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", watch["operator"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", watch["value"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", watch["period"].(int)))
		}
	}
	return hashcode.String(buf.String())
}

func resourceSoftLayerScalePolicyHandlerHash(v interface{}) int {
	var buf bytes.Buffer
	watch := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", watch["metric"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", watch["operator"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", watch["value"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", watch["period"].(int)))
	return hashcode.String(buf.String())
}
