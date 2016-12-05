package softlayer

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"log"
)

const (
	aclMask = "name,firewallInterfaces[name,firewallContextAccessControlLists]"
)

func resourceSoftLayerFwHardwareDedicatedRules() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerFwHardwareDedicatedRulesCreate,
		Read:     resourceSoftLayerFwHardwareDedicatedRulesRead,
		Update:   resourceSoftLayerFwHardwareDedicatedRulesUpdate,
		Delete:   resourceSoftLayerFwHardwareDedicatedRulesDelete,
		Exists:   resourceSoftLayerFwHardwareDedicatedRulesExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"firewall_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"order_value": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
						"src_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"src_ip_cidr": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"dst_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"dst_ip_cidr": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"dst_port_range_start": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"dst_port_range_end": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"notes": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: func(v interface{}) int {
					rule := v.(map[string]interface{})
					return rule["order_value"].(int)
				},
			},
		},
	}
}

func prepareRules(d *schema.ResourceData) []datatypes.Network_Firewall_Update_Request_Rule {
	ruleList := d.Get("rules").(*schema.Set).List()
	rules := make([]datatypes.Network_Firewall_Update_Request_Rule, 0)
	for _, ruleItem := range ruleList {
		ruleMap := ruleItem.(map[string]interface{})
		var rule datatypes.Network_Firewall_Update_Request_Rule
		rule.OrderValue = sl.Int(ruleMap["order_value"].(int))
		rule.Action = sl.String(ruleMap["action"].(string))
		rule.SourceIpAddress = sl.String(ruleMap["src_ip_address"].(string))
		rule.SourceIpCidr = sl.Int(ruleMap["src_ip_cidr"].(int))
		rule.DestinationIpAddress = sl.String(ruleMap["dst_ip_address"].(string))
		rule.DestinationIpCidr = sl.Int(ruleMap["dst_ip_cidr"].(int))
		rule.DestinationPortRangeStart = sl.Int(ruleMap["dst_port_range_start"].(int))
		rule.DestinationPortRangeEnd = sl.Int(ruleMap["dst_port_range_end"].(int))
		rule.Protocol = sl.String(ruleMap["protocol"].(string))
		if len(ruleMap["notes"].(string)) > 0 {
			rule.Notes = sl.String(ruleMap["notes"].(string))
		}
		rules = append(rules, rule)
	}
	return rules
}

func getFirewallContextAccessControlListId(fwId int, sess *session.Session) (int, error) {
	service := services.GetNetworkVlanFirewallService(sess)
	vlan, err := service.Id(fwId).Mask(aclMask).GetNetworkVlans()

	if err != nil {
		return 0, err
	}

	for _, fwInterface := range vlan[0].FirewallInterfaces {
		if fwInterface.Name != nil &&
			*fwInterface.Name == "outside" &&
			len(fwInterface.FirewallContextAccessControlLists) > 0 &&
			fwInterface.FirewallContextAccessControlLists[0].Id != nil {
			return *fwInterface.FirewallContextAccessControlLists[0].Id, nil
		}
	}
	return 0, fmt.Errorf("No firewallContextAccessControlListId.")
}

func resourceSoftLayerFwHardwareDedicatedRulesCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	fwId := d.Get("firewall_id").(int)
	rules := prepareRules(d)

	fwContextACLId, err := getFirewallContextAccessControlListId(fwId, sess)
	if err != nil {
		return fmt.Errorf("Error during creation of dedicated hardware firewall rules: %s", err)
	}

	ruleTemplate := datatypes.Network_Firewall_Update_Request{
		FirewallContextAccessControlListId: sl.Int(fwContextACLId),
		Rules: rules,
	}

	log.Println("[INFO] Creating dedicated hardware firewall rules")

	_, err = services.GetNetworkFirewallUpdateRequestService(sess).CreateObject(&ruleTemplate)
	if err != nil {
		return fmt.Errorf("Error during creation of dedicated hardware firewall rules: %s", err)
	}

	d.SetId(strconv.Itoa(fwId))

	log.Printf("[INFO] Firewall rules ID: %s", d.Id())

	return resourceSoftLayerFwHardwareDedicatedRulesRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedRulesRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	fwRulesID, _ := strconv.Atoi(d.Id())

	fw, err := services.GetNetworkVlanFirewallService(sess).
		Id(fwRulesID).
		Mask("rules").
		GetObject()

	if err != nil {
		return fmt.Errorf("Error retrieving firewall rules: %s", err)
	}

	rules := make([]map[string]interface{}, 0, len(fw.Rules))
	for _, rule := range fw.Rules {
		r := make(map[string]interface{})
		r["order_value"] = *rule.OrderValue
		r["action"] = *rule.Action
		r["src_ip_address"] = *rule.SourceIpAddress
		r["src_ip_cidr"] = *rule.SourceIpCidr
		r["dst_ip_address"] = *rule.DestinationIpAddress
		r["dst_ip_cidr"] = *rule.DestinationIpCidr
		r["dst_port_range_start"] = *rule.DestinationPortRangeStart
		r["dst_port_range_end"] = *rule.DestinationPortRangeEnd
		r["protocol"] = *rule.Protocol
		if len(*rule.Notes) > 0 {
			r["notes"] = *rule.Notes
		}
		rules = append(rules, r)
	}

	d.Set("firewall_id", fwRulesID)
	d.Set("rules", rules)

	return nil
}

func resourceSoftLayerFwHardwareDedicatedRulesUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	fwId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid firewall ID, must be an integer: %s", err)
	}
	rules := prepareRules(d)

	fwContextACLId, err := getFirewallContextAccessControlListId(fwId, sess)
	if err != nil {
		return fmt.Errorf("Error during updating of dedicated hardware firewall rules: %s", err)
	}

	ruleTemplate := datatypes.Network_Firewall_Update_Request{
		FirewallContextAccessControlListId: sl.Int(fwContextACLId),
		Rules: rules,
	}

	log.Println("[INFO] Updating dedicated hardware firewall rules")

	_, err = services.GetNetworkFirewallUpdateRequestService(sess).CreateObject(&ruleTemplate)
	if err != nil {
		return fmt.Errorf("Error during updating of dedicated hardware firewall rules: %s", err)
	}
	return resourceSoftLayerFwHardwareDedicatedRulesRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedRulesDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	fwService := services.GetNetworkVlanFirewallService(sess)

	fwID, _ := strconv.Atoi(d.Id())

	// Get billing item associated with the firewall
	billingItem, err := fwService.Id(fwID).GetBillingItem()

	if err != nil {
		return fmt.Errorf("Error while looking up billing item associated with the firewall: %s", err)
	}

	if billingItem.Id == nil {
		return fmt.Errorf("Error while looking up billing item associated with the firewall: No billing item for ID:%d", fwID)
	}

	success, err := services.GetBillingItemService(sess).Id(*billingItem.Id).CancelService()
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("SoftLayer reported an unsuccessful cancellation")
	}

	return nil
}

func resourceSoftLayerFwHardwareDedicatedRulesExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

	fwID, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = services.GetNetworkVlanFirewallService(sess).
		Id(fwID).
		GetObject()

	if err != nil {
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error retrieving firewall information: %s", err)
	}

	return true, nil
}
