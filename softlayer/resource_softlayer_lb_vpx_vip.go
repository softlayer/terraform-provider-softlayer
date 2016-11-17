package softlayer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/helpers/network"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"

	"github.com/minsikl/netscaler-nitro-go/client"
	dt "github.com/minsikl/netscaler-nitro-go/datatypes"
	"github.com/minsikl/netscaler-nitro-go/op"
)

const (
	VPX_VERSION_10_1 = "10.1"
)

var (
	// Load balancing algorithm mapping tables

	lbMethodMapFromSLtoVPX105 = map[string][2]string{
		"rr":    {"NONE", "ROUNDROBIN"},
		"sr":    {"NONE", "LEASTRESPONSETIME"},
		"lc":    {"NONE", "LEASTCONNECTION"},
		"pi":    {"SOURCEIP", "ROUNDROBIN"},
		"pi-sr": {"SOURCEIP", "LEASTRESPONSETIME"},
		"pi-lc": {"SOURCEIP", "LEASTCONNECTION"},
		"ic":    {"COOKIEINSERT", "ROUNDROBIN"},
		"ic-sr": {"COOKIEINSERT", "LEASTRESPONSETIME"},
		"ic-lc": {"COOKIEINSERT", "LEASTCONNECTION"},
	}

	lbMethodMapFromVPX105toSL = map[[2]string]string{
		{"NONE", "ROUNDROBIN"}:                "rr",
		{"NONE", "LEASTRESPONSETIME"}:         "sr",
		{"NONE", "LEASTCONNECTION"}:           "lc",
		{"SOURCEIP", "ROUNDROBIN"}:            "pi",
		{"SOURCEIP", "LEASTRESPONSETIME"}:     "pi-sr",
		{"SOURCEIP", "LEASTCONNECTION"}:       "pi-lc",
		{"COOKIEINSERT", "ROUNDROBIN"}:        "ic",
		{"COOKIEINSERT", "LEASTRESPONSETIME"}: "ic-sr",
		{"COOKIEINSERT", "LEASTCONNECTION"}:   "ic-lc",
	}
)

func resourceSoftLayerLbVpxVip() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbVpxVipCreate,
		Read:     resourceSoftLayerLbVpxVipRead,
		Update:   resourceSoftLayerLbVpxVipUpdate,
		Delete:   resourceSoftLayerLbVpxVipDelete,
		Exists:   resourceSoftLayerLbVpxVipExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"nad_controller_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"load_balancing_method": {
				Type:     schema.TypeString,
				Required: true,
			},

			"persistence": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// name field is actually used as an ID in SoftLayer
			// http://sldn.softlayer.com/reference/services/SoftLayer_Network_Application_Delivery_Controller/updateLiveLoadBalancer
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"source_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// security_certificate_id is only acceptable with SSL type
			"security_certificate_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"virtual_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSoftLayerLbVpxVipCreate(d *schema.ResourceData, meta interface{}) error {
	version, err := getVPXVersion(d.Get("nad_controller_id").(int), meta.(ProviderConfig).SoftLayerSession())
	if err != nil {
		return fmt.Errorf("Error creating Virtual Ip Address: %s", err)
	}

	if version == VPX_VERSION_10_1 {
		return resourceSoftLayerLbVpxVipCreate101(d, meta)
	}

	return resourceSoftLayerLbVpxVipCreate105(d, meta)
}

func resourceSoftLayerLbVpxVipRead(d *schema.ResourceData, meta interface{}) error {
	nadcId, _, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error Reading Virtual IP Address: %s", err)
	}

	version, err := getVPXVersion(nadcId, meta.(ProviderConfig).SoftLayerSession())
	if err != nil {
		return fmt.Errorf("Error Reading Virtual Ip Address: %s", err)
	}

	if version == VPX_VERSION_10_1 {
		return resourceSoftLayerLbVpxVipRead101(d, meta)
	}

	return resourceSoftLayerLbVpxVipRead105(d, meta)
}

func resourceSoftLayerLbVpxVipUpdate(d *schema.ResourceData, meta interface{}) error {
	nadcId, _, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error updating Virtual IP Address: %s", err)
	}

	version, err := getVPXVersion(nadcId, meta.(ProviderConfig).SoftLayerSession())
	if err != nil {
		return fmt.Errorf("Error updating Virtual Ip Address: %s", err)
	}

	if version == VPX_VERSION_10_1 {
		return resourceSoftLayerLbVpxVipUpdate101(d, meta)
	}

	return resourceSoftLayerLbVpxVipUpdate105(d, meta)
}

func resourceSoftLayerLbVpxVipDelete(d *schema.ResourceData, meta interface{}) error {
	nadcId, _, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address: %s", err)
	}

	version, err := getVPXVersion(nadcId, meta.(ProviderConfig).SoftLayerSession())
	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address: %s", err)
	}

	if version == VPX_VERSION_10_1 {
		return resourceSoftLayerLbVpxVipDelete101(d, meta)
	}

	return resourceSoftLayerLbVpxVipDelete105(d, meta)
}

func resourceSoftLayerLbVpxVipExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	nadcId, _, err := parseId(d.Id())
	if err != nil {
		return false, fmt.Errorf("Error in exists: %s", err)
	}

	version, err := getVPXVersion(nadcId, meta.(ProviderConfig).SoftLayerSession())
	if err != nil {
		return false, fmt.Errorf("Error in exists: %s", err)
	}

	if version == VPX_VERSION_10_1 {
		return resourceSoftLayerLbVpxVipExists101(d, meta)
	}

	return resourceSoftLayerLbVpxVipExists105(d, meta)
}

func parseId(id string) (int, string, error) {
	if len(id) < 1 {
		return 0, "", fmt.Errorf("Failed to parse id %s: Unable to get a VIP ID", id)
	}

	idList := strings.Split(id, ":")
	if len(idList) != 2 || len(idList[0]) < 1 || len(idList[1]) < 1 {
		return 0, "", fmt.Errorf("Failed to parse id %s: Invalid VIP ID", id)
	}

	nadcId, err := strconv.Atoi(idList[0])
	if err != nil {
		return 0, "", fmt.Errorf("Failed to parse id : Unable to get a VIP ID %s", err)
	}

	vipName := idList[1]
	return nadcId, vipName, nil
}

func resourceSoftLayerLbVpxVipCreate101(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("security_certificate_id"); ok {
		return fmt.Errorf("Error creating Virtual Ip Address: security_certificate_id is not supported with VPX 10.1.")
	}

	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	vipName := d.Get("name").(string)

	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		LoadBalancingMethod: sl.String(d.Get("load_balancing_method").(string)),
		Name:                sl.String(vipName),
		SourcePort:          sl.Int(d.Get("source_port").(int)),
		Type:                sl.String(d.Get("type").(string)),
		VirtualIpAddress:    sl.String(d.Get("virtual_ip_address").(string)),
	}

	log.Printf("[INFO] Creating Virtual Ip Address %s", *template.VirtualIpAddress)

	var err error
	var successFlag bool

	for count := 0; count < 10; count++ {
		successFlag, err = service.Id(nadcId).CreateLiveLoadBalancer(&template)
		log.Printf("[INFO] Creating Virtual Ip Address %s successFlag : %t", *template.VirtualIpAddress, successFlag)

		if err != nil && strings.Contains(err.Error(), "already exists") {
			log.Printf("[INFO] Creating Virtual Ip Address %s error : %s. Ingore the error.", *template.VirtualIpAddress, err.Error())
			successFlag = true
			err = nil
			break
		}

		if err != nil && strings.Contains(err.Error(), "Operation already in progress") {
			log.Printf("[INFO] Creating Virtual Ip Address %s error : %s. Retry in 10 secs", *template.VirtualIpAddress, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error creating Virtual Ip Address: %s", err)
	}

	if !successFlag {
		return errors.New("Error creating Virtual Ip Address")
	}

	d.SetId(fmt.Sprintf("%d:%s", nadcId, vipName))

	log.Printf("[INFO] Netscaler VPX VIP ID: %s", d.Id())

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipCreate105(d *schema.ResourceData, meta interface{}) error {
	nadcId := d.Get("nad_controller_id").(int)
	nClient, err := getNitroClient(meta.(ProviderConfig).SoftLayerSession(), nadcId)
	if err != nil {
		return fmt.Errorf("Error getting netscaler information ID: %d", nadcId)
	}

	vipName := d.Get("name").(string)
	vipType := d.Get("type").(string)
	securityCertificateId := d.Get("security_certificate_id").(int)

	lbvserverReq := dt.LbvserverReq{
		Lbvserver: &dt.Lbvserver{
			Name:        op.String(vipName),
			Ipv46:       op.String(d.Get("virtual_ip_address").(string)),
			Port:        op.Int(d.Get("source_port").(int)),
			ServiceType: op.String(vipType),
		},
	}

	if len(d.Get("persistence").(string)) > 0 {
		lbvserverReq.Lbvserver.Lbmethod = op.String(d.Get("persistence").(string))
	}
	lbMethodPair := lbMethodMapFromSLtoVPX105[d.Get("load_balancing_method").(string)]
	if len(lbMethodPair[1]) > 0 {
		if len(lbMethodPair[0]) > 0 {
			lbvserverReq.Lbvserver.Persistencetype = &lbMethodPair[0]
		} else {
			lbvserverReq.Lbvserver.Persistencetype = op.String("NONE")
		}
		lbvserverReq.Lbvserver.Lbmethod = &lbMethodPair[1]
	}

	log.Printf("[INFO] Creating Virtual Ip Address %s", *lbvserverReq.Lbvserver.Ipv46)

	// security_certificated_id is only available when type is 'SSL'
	if securityCertificateId > 0 && vipType != "SSL" {
		return fmt.Errorf("Error creating VIP : security_certificated_id is only available when type is 'SSL'")
	} else if securityCertificateId == 0 && vipType == "SSL" {
		return fmt.Errorf("Error creating VIP : 'SSL' type requires security_certificated_id.")

	}

	// Create a virtual server
	err = nClient.Add(&lbvserverReq)
	if err != nil {
		return err
	}

	// Configure security_certificate for SSL Offload.
	if vipType == "SSL" {
		// Delete the previous security certificate.
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)

		err = configureSecurityCertificate(nClient, meta.(ProviderConfig).SoftLayerSession(), vipName, securityCertificateId)

		if err != nil {
			// Rollback VIP creation and return an error.
			resourceSoftLayerLbVpxVipDelete105(d, meta)
			return err
		}
	}

	d.SetId(fmt.Sprintf("%d:%s", nadcId, vipName))

	log.Printf("[INFO] Netscaler VPX VIP ID: %s", d.Id())

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipRead101(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	vip, err := network.GetNadcLbVipByName(sess, nadcId, vipName)
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : while looking up a virtual ip address : %s", err)
	}

	d.Set("nad_controller_id", nadcId)
	if vip.LoadBalancingMethod != nil {
		d.Set("load_balancing_method", *vip.LoadBalancingMethod)
	}

	if vip.Name != nil {
		d.Set("name", *vip.Name)
	}

	if vip.SourcePort != nil {
		d.Set("source_port", *vip.SourcePort)
	}

	if vip.Type != nil {
		d.Set("type", *vip.Type)
	}

	if vip.VirtualIpAddress != nil {
		d.Set("virtual_ip_address", *vip.VirtualIpAddress)
	}

	return nil
}

func resourceSoftLayerLbVpxVipRead105(d *schema.ResourceData, meta interface{}) error {
	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	nClient, err := getNitroClient(meta.(ProviderConfig).SoftLayerSession(), nadcId)
	if err != nil {
		return fmt.Errorf("Error getting netscaler information ID: %d", nadcId)
	}

	// Read a virtual server
	vip := dt.LbvserverRes{}
	err = nClient.Get(&vip, vipName)
	if err != nil {
		fmt.Printf("Error getting VIP information : %s", err.Error())
	}

	d.Set("nad_controller_id", nadcId)
	if vip.Lbvserver[0].Lbmethod != nil {
		d.Set("load_balancing_method", *vip.Lbvserver[0].Lbmethod)
	}

	if vip.Lbvserver[0].Name != nil {
		d.Set("name", *vip.Lbvserver[0].Name)
	}

	if vip.Lbvserver[0].Port != nil {
		d.Set("source_port", *vip.Lbvserver[0].Port)
	}

	if vip.Lbvserver[0].ServiceType != nil {
		d.Set("type", *vip.Lbvserver[0].ServiceType)
	}

	if vip.Lbvserver[0].Persistencetype != nil {
		if *vip.Lbvserver[0].Persistencetype == "NONE" {
			d.Set("persistence", nil)
		} else {
			d.Set("persistence", *vip.Lbvserver[0].Persistencetype)
		}
	}

	lbMethod := lbMethodMapFromVPX105toSL[[2]string{*vip.Lbvserver[0].Persistencetype, *vip.Lbvserver[0].Lbmethod}]
	if len(lbMethod) > 0 {
		d.Set("load_balancing_method", lbMethod)
	}

	if vip.Lbvserver[0].Ipv46 != nil {
		d.Set("virtual_ip_address", *vip.Lbvserver[0].Ipv46)
	}

	// Read a security certificate information
	securityCertificateId, err := getSecurityCertificateId(nClient, vipName)
	if err == nil {
		d.Set("security_certificate_id", securityCertificateId)
	} else {
		if _, ok := d.GetOk("security_certificate_id"); ok {
			d.Set("security_certificate_id", 0)
		}
	}

	return nil
}

func resourceSoftLayerLbVpxVipUpdate101(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId := d.Get("nad_controller_id").(int)
	template := datatypes.Network_LoadBalancer_VirtualIpAddress{
		Name: sl.String(d.Get("name").(string)),
	}

	if d.HasChange("load_balancing_method") {
		template.LoadBalancingMethod = sl.String(d.Get("load_balancing_method").(string))
	}

	if d.HasChange("virtual_ip_address") {
		template.VirtualIpAddress = sl.String(d.Get("virtual_ip_address").(string))
	}

	var err error

	for count := 0; count < 10; count++ {
		var successFlag bool
		successFlag, err = service.Id(nadcId).UpdateLiveLoadBalancer(&template)
		log.Printf("[INFO]  Updating Virtual Ip Address successFlag : %t", successFlag)

		if err != nil && strings.Contains(err.Error(), "Operation already in progress") {
			log.Printf("[INFO] Updating Virtual Ip Address error : %s. Retry in 10 secs", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error updating Virtual Ip Address: %s", err)
	}

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipUpdate105(d *schema.ResourceData, meta interface{}) error {
	nadcId := d.Get("nad_controller_id").(int)
	nClient, err := getNitroClient(meta.(ProviderConfig).SoftLayerSession(), nadcId)
	if err != nil {
		return fmt.Errorf("Error getting netscaler information ID: %d", nadcId)
	}

	lbvserverReq := dt.LbvserverReq{
		Lbvserver: &dt.Lbvserver{
			Name: op.String(d.Get("name").(string)),
		},
	}

	if d.HasChange("load_balancing_method") || d.HasChange("persistence") {
		lbvserverReq.Lbvserver.Persistencetype = op.String(d.Get("persistence").(string))
		lbvserverReq.Lbvserver.Lbmethod = op.String(d.Get("load_balancing_method").(string))

		lbMethodPair := lbMethodMapFromSLtoVPX105[d.Get("load_balancing_method").(string)]
		if len(lbMethodPair[1]) > 0 {
			if len(lbMethodPair[0]) > 0 {
				lbvserverReq.Lbvserver.Persistencetype = &lbMethodPair[0]
			} else {
				lbvserverReq.Lbvserver.Persistencetype = op.String("NONE")
			}
			lbvserverReq.Lbvserver.Lbmethod = &lbMethodPair[1]
		}
	}

	if d.HasChange("virtual_ip_address") {
		lbvserverReq.Lbvserver.Ipv46 = sl.String(d.Get("virtual_ip_address").(string))
	}

	// Update the virtual server
	err = nClient.Update(&lbvserverReq)
	if err != nil {
		return fmt.Errorf("Error updating Virtual Ip Address: " + err.Error())
	}

	return resourceSoftLayerLbVpxVipRead(d, meta)
}

func resourceSoftLayerLbVpxVipDelete101(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetNetworkApplicationDeliveryControllerService(sess)

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	for count := 0; count < 10; count++ {
		var successFlag bool
		successFlag, err = service.Id(nadcId).DeleteLiveLoadBalancer(
			&datatypes.Network_LoadBalancer_VirtualIpAddress{Name: sl.String(vipName)},
		)
		log.Printf("[INFO] Deleting Virtual Ip Address %s successFlag : %t", vipName, successFlag)

		if err != nil &&
			(strings.Contains(err.Error(), "Operation already in progress") ||
				strings.Contains(err.Error(), "No Service")) {
			log.Printf("[INFO] Deleting Virtual Ip Address %s Error : %s  Retry in 10 secs", vipName, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}

		// Check if the resource is already deleted.
		if err != nil && strings.Contains(err.Error(), "Unable to find object with unknown identifier of") {
			log.Printf("[INFO] Deleting Virtual Ip Address %s Error : %s . Ignore the error.", vipName, err.Error())
			err = nil
		}

		break
	}

	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address %s: %s", vipName, err)
	}

	return nil
}

func resourceSoftLayerLbVpxVipDelete105(d *schema.ResourceData, meta interface{}) error {
	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	nClient, err := getNitroClient(meta.(ProviderConfig).SoftLayerSession(), nadcId)
	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address %s: %s", vipName, err)
	}

	// Delete a virtual server
	err = nClient.Delete(&dt.LbvserverReq{}, vipName)
	if err != nil {
		return fmt.Errorf("Error deleting Virtual Ip Address %s: %s", vipName, err)
	}

	// Delete a security certificate
	securityCertificateId, err := getSecurityCertificateId(nClient, vipName)
	if err == nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
	}

	return nil
}

func resourceSoftLayerLbVpxVipExists101(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(ProviderConfig).SoftLayerSession()

	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return false, fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	vip, err := network.GetNadcLbVipByName(sess, nadcId, vipName)

	return vip != nil && err == nil && *vip.Name == vipName, nil
}

func resourceSoftLayerLbVpxVipExists105(d *schema.ResourceData, meta interface{}) (bool, error) {
	nadcId, vipName, err := parseId(d.Id())
	if err != nil {
		return false, fmt.Errorf("softlayer_lb_vpx : %s", err)
	}

	nClient, err := getNitroClient(meta.(ProviderConfig).SoftLayerSession(), nadcId)
	if err != nil {
		return false, err
	}

	// Read a virtual server
	vip := dt.LbvserverRes{}
	err = nClient.Get(&vip, vipName)

	if err != nil && strings.Contains(err.Error(), "No such resource") {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func getNitroClient(sess *session.Session, nadcId int) (*client.NitroClient, error) {
	service := services.GetNetworkApplicationDeliveryControllerService(sess)
	nadc, err := service.Id(nadcId).Mask("managementIpAddress,password[password]").GetObject()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving netscaler: %s", err)
	}
	return client.NewNitroClient("http", *nadc.ManagementIpAddress, dt.CONFIG,
		"root", *nadc.Password.Password, true), nil
}

func configureSecurityCertificate(nClient *client.NitroClient, sess *session.Session, vipName string, securityCertificateId int) error {
	// Read security_certificate
	service := services.GetSecurityCertificateService(sess)
	cert, err := service.Id(securityCertificateId).GetObject()
	if err != nil {
		return fmt.Errorf("Unable to get Security Certificate: %s", err)
	}

	certName := vipName + "_" + strconv.Itoa(securityCertificateId)
	certFileName := certName + ".cert"
	keyFileName := certName + ".key"

	// Delete previous security certificate
	deleteSecurityCertificate(nClient, vipName, securityCertificateId)

	// Upload security_certificate
	certReq := dt.SystemfileReq{
		Systemfile: &dt.Systemfile{
			Filename:     op.String(certFileName),
			Filecontent:  op.String(base64.StdEncoding.EncodeToString([]byte(*cert.Certificate))),
			Filelocation: op.String("/nsconfig/ssl/"),
			Fileencoding: op.String("BASE64"),
		},
	}

	err = nClient.Add(&certReq)
	if err != nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
		return err
	}

	keyReq := dt.SystemfileReq{
		Systemfile: &dt.Systemfile{
			Filename:     op.String(keyFileName),
			Filecontent:  op.String(base64.StdEncoding.EncodeToString([]byte(*cert.PrivateKey))),
			Filelocation: op.String("/nsconfig/ssl/"),
			Fileencoding: op.String("BASE64"),
		},
	}

	err = nClient.Add(&keyReq)
	if err != nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
		return err
	}

	// Enable SSL

	sslFeature := dt.NsfeatureReq{
		Nsfeature: &dt.Nsfeature{
			Feature: []string{"ssl"},
		},
	}

	err = nClient.Enable(&sslFeature, true)
	if err != nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
		return err
	}

	// Register SSL

	sslCertKey := dt.SslcertkeyReq{
		Sslcertkey: &dt.Sslcertkey{
			Certkey: op.String(certName),
			Cert:    op.String(certFileName),
			Key:     op.String(keyFileName),
		},
	}

	err = nClient.Add(&sslCertKey)
	if err != nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
		return err
	}

	// Bind security_certificate

	sslBind := dt.SslvserverSslcertkeyBindingReq{
		SslvserverSslcertkeyBinding: &dt.SslvserverSslcertkeyBinding{
			Vservername: op.String(vipName),
			Certkeyname: op.String(certName),
		},
	}

	err = nClient.Add(&sslBind)
	if err != nil {
		deleteSecurityCertificate(nClient, vipName, securityCertificateId)
		return err
	}
	return nil
}

func deleteSecurityCertificate(nClient *client.NitroClient, vipName string, securityCertificateId int) {
	certName := vipName + "_" + strconv.Itoa(securityCertificateId)
	certFileName := certName + ".cert"
	keyFileName := certName + ".key"

	// Delete sslvserversslcertkeybinding
	nClient.Delete(&dt.SslvserverSslcertkeyBindingReq{}, vipName, "args=certkeyname:"+certName)

	// Delete sslcertkey
	nClient.Delete(&dt.SslcertkeyReq{}, certName)

	// Delete cert
	nClient.Delete(&dt.SystemfileReq{}, certFileName, "args=fileLocation:"+"%2Fnsconfig%2Fssl%2F")

	// Delete key
	nClient.Delete(&dt.SystemfileReq{}, keyFileName, "args=fileLocation:"+"%2Fnsconfig%2Fssl%2F")
}

func getSecurityCertificateId(nClient *client.NitroClient, vipName string) (int, error) {
	securityCertificateId := 0
	res := dt.SslcertkeyRes{}
	err := nClient.Get(&res, "")
	if err != nil {
		return 0, fmt.Errorf("Error getting securityCertificateId information : %s", err.Error())
	}

	//CertKey name is consisted of `vipName`_`securityCertificateId`.
	for _, sslCertKey := range res.Sslcertkey {
		sslCertKeyArr := strings.Split(*sslCertKey.Certkey, "_")
		if len(sslCertKeyArr) < 2 || !strings.HasPrefix(*sslCertKey.Certkey, vipName+"_") {
			continue
		}

		securityCertificateId, err = strconv.Atoi(sslCertKeyArr[len(sslCertKeyArr)-1])
		if err != nil {
			continue
		} else {
			return securityCertificateId, nil
		}
	}
	return 0, fmt.Errorf("Error getting securityCertificateId information : No security certificate for %s", vipName)
}
