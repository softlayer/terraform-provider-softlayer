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
	"time"
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
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						// ICMP, GRE, AH, and ESP don't require port ranges.
						"dst_port_range_start": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"dst_port_range_end": {
							Type:     schema.TypeInt,
							Optional: true,
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
			},
		},
	}
}

func prepareRules(d *schema.ResourceData) []datatypes.Network_Firewall_Update_Request_Rule {
	ruleList := d.Get("rules").([]interface{})
	rules := make([]datatypes.Network_Firewall_Update_Request_Rule, 0)
	for i, ruleItem := range ruleList {
		ruleMap := ruleItem.(map[string]interface{})
		var rule datatypes.Network_Firewall_Update_Request_Rule
		rule.OrderValue = sl.Int(i + 1)
		rule.Action = sl.String(ruleMap["action"].(string))
		rule.SourceIpAddress = sl.String(ruleMap["src_ip_address"].(string))
		rule.SourceIpCidr = sl.Int(ruleMap["src_ip_cidr"].(int))
		rule.DestinationIpAddress = sl.String(ruleMap["dst_ip_address"].(string))
		rule.DestinationIpCidr = sl.Int(ruleMap["dst_ip_cidr"].(int))

		if ruleMap["dst_port_range_start"] != nil {
			rule.DestinationPortRangeStart = sl.Int(ruleMap["dst_port_range_start"].(int))
		}
		if ruleMap["dst_port_range_end"] != nil {
			rule.DestinationPortRangeEnd = sl.Int(ruleMap["dst_port_range_end"].(int))
		}

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
	log.Printf("[INFO] Wait one minute for applying the rules.")
	time.Sleep(time.Minute)

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
		r["action"] = *rule.Action
		r["src_ip_address"] = *rule.SourceIpAddress
		r["src_ip_cidr"] = *rule.SourceIpCidr
		r["dst_ip_address"] = *rule.DestinationIpAddress
		r["dst_ip_cidr"] = *rule.DestinationIpCidr
		if rule.DestinationPortRangeStart != nil {
			r["dst_port_range_start"] = *rule.DestinationPortRangeStart
		}
		if rule.DestinationPortRangeEnd != nil {
			r["dst_port_range_end"] = *rule.DestinationPortRangeEnd
		}
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

func appendAnyOpenRule(rules []datatypes.Network_Firewall_Update_Request_Rule, protocol string) []datatypes.Network_Firewall_Update_Request_Rule {
	ruleAnyOpen := datatypes.Network_Firewall_Update_Request_Rule{
		OrderValue:                sl.Int(len(rules) + 1),
		Action:                    sl.String("permit"),
		SourceIpAddress:           sl.String("any"),
		DestinationIpAddress:      sl.String("any"),
		DestinationPortRangeStart: sl.Int(1),
		DestinationPortRangeEnd:   sl.Int(65535),
		Protocol:                  sl.String(protocol),
		Notes:                     sl.String("terraform-default-anyopen-" + protocol),
	}
	return append(rules, ruleAnyOpen)
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
	time.Sleep(time.Minute)

	return resourceSoftLayerFwHardwareDedicatedRulesRead(d, meta)
}

func resourceSoftLayerFwHardwareDedicatedRulesDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	fwId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid firewall ID, must be an integer: %s", err)
	}

	fwContextACLId, err := getFirewallContextAccessControlListId(fwId, sess)
	if err != nil {
		return fmt.Errorf("Error during deleting of dedicated hardware firewall rules: %s", err)
	}

	ruleTemplate := datatypes.Network_Firewall_Update_Request{
		FirewallContextAccessControlListId: sl.Int(fwContextACLId),
	}

	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "tcp")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "udp")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "icmp")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "gre")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "pptp")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "ah")
	ruleTemplate.Rules = appendAnyOpenRule(ruleTemplate.Rules, "esp")

	log.Println("[INFO] Deleting dedicated hardware firewall rules")

	_, err = services.GetNetworkFirewallUpdateRequestService(sess).CreateObject(&ruleTemplate)
	if err != nil {
		return fmt.Errorf("Error during deleting of dedicated hardware firewall rules: %s", err)
	}
	time.Sleep(time.Minute)

	return nil
}

func resourceSoftLayerFwHardwareDedicatedRulesExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

	fwRulesID, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	fw, err := services.GetNetworkVlanFirewallService(sess).
		Id(fwRulesID).
		Mask("rules").
		GetObject()

	if err != nil {
		return false, fmt.Errorf("Error retrieving firewall rules: %s", err)
	}

	if len(fw.Rules) == 0 {
		return false, nil
	}

	return true, nil
}
