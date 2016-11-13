package softlayer

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/minsikl/netscaler-nitro-go/client"
	dt "github.com/minsikl/netscaler-nitro-go/datatypes"
	"github.com/minsikl/netscaler-nitro-go/op"
	"github.com/softlayer/softlayer-go/session"
	"log"
	"strconv"
	"strings"
	"time"
)

func resourceSoftLayerLbVpxHa() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerLbVpxHaCreate,
		Read:     resourceSoftLayerLbVpxHaRead,
		Delete:   resourceSoftLayerLbVpxHaDelete,
		Exists:   resourceSoftLayerLbVpxHaExists,
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"primary_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"secondary_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"stay_secondary": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func configureHA(nClient1 *client.NitroClient, nClient2 *client.NitroClient, staySecondary bool) error {
	// 1. VPX2 : Sync password
	systemuserReq2 := dt.SystemuserReq{
		Systemuser: &dt.Systemuser{
			Username: op.String("root"),
			Password: op.String(nClient1.Password),
		},
	}
	err := nClient2.Update(&systemuserReq2)
	if err != nil {
		return err
	}
	nClient2.Password = nClient1.Password

	// 2. VPX1 : Register hanode
	hanodeReq1 := dt.HanodeReq{
		Hanode: &dt.Hanode{
			Id:        op.Int(2),
			Ipaddress: op.String(nClient2.IpAddress),
		},
	}

	err = nClient1.Add(&hanodeReq1)
	if err != nil {
		return err
	}

	// Wait 5 secs to make VPX1 a primary node.
	time.Sleep(time.Second * 5)

	// 3. VPX2 : Register hanode
	hanodeReq2 := dt.HanodeReq{
		Hanode: &dt.Hanode{
			Id:        op.Int(2),
			Ipaddress: op.String(nClient1.IpAddress),
		},
	}
	err = nClient2.Add(&hanodeReq2)
	if err != nil {
		return err
	}

	// Update STAYSECONDARY
	if staySecondary {
		stay := dt.HanodeReq{
			Hanode: &dt.Hanode{
				Hastatus: op.String("STAYSECONDARY"),
			},
		}
		err = nClient2.Update(&stay)
		if err != nil {
			return err
		}
	}

	// 4. VPX1 : Register rpcnode
	nsrpcnode1 := dt.NsrpcnodeReq{
		Nsrpcnode: &dt.Nsrpcnode{
			Ipaddress: op.String(nClient1.IpAddress),
			Password:  op.String(nClient1.Password),
		},
	}
	err = nClient1.Update(&nsrpcnode1)
	if err != nil {
		return err
	}
	nsrpcnode1.Nsrpcnode.Ipaddress = op.String(nClient2.IpAddress)
	err = nClient1.Update(&nsrpcnode1)
	if err != nil {
		return err
	}

	// 5. VPX2 : Register rpcnode
	nsrpcnode2 := dt.NsrpcnodeReq{
		Nsrpcnode: &dt.Nsrpcnode{
			Ipaddress: op.String(nClient1.IpAddress),
			Password:  op.String(nClient1.Password),
		},
	}
	err = nClient2.Update(&nsrpcnode2)
	if err != nil {
		return err
	}
	nsrpcnode2.Nsrpcnode.Ipaddress = op.String(nClient2.IpAddress)
	err = nClient2.Update(&nsrpcnode2)
	if err != nil {
		return err
	}

	// 6. VPX1 : Sync files
	hafiles := dt.HafilesReq{
		Hafiles: &dt.Hafiles{
			[]string{"all"},
		},
	}
	err = nClient1.Add(&hafiles, "action=sync")
	if err != nil {
		return err
	}

	return nil
}

func deleteHA(nClient1 *client.NitroClient, nClient2 *client.NitroClient) error {
	// 1. VPX2 : Delete hanode
	err := nClient2.Delete(&dt.HanodeReq{}, "2")
	if err != nil {
		return err
	}

	// 2. VPX1 : Delete hanode
	err = nClient1.Delete(&dt.HanodeReq{}, "2")
	if err != nil {
		return err
	}
	return nil
}

func parseHAId(id string) (int, int, error) {
	if len(id) < 1 {
		return 0, 0, fmt.Errorf("Failed to parse id : Unable to get netscaler Ids")
	}
	idList := strings.Split(id, ":")
	if len(idList) != 2 || len(idList[0]) < 1 || len(idList[1]) < 1 {
		return 0, 0, fmt.Errorf("Failed to parse id : Invalid HA ID")
	}
	primaryId, err := strconv.Atoi(idList[0])
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to parse id : Unable to get a primaryId %s", err)
	}
	secondaryId, err := strconv.Atoi(idList[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to parse id : Unable to get a secondaryId %s", err)
	}
	return primaryId, secondaryId, nil
}

func resourceSoftLayerLbVpxHaCreate(d *schema.ResourceData, meta interface{}) error {
	primaryId := d.Get("primary_id").(int)
	secondaryId := d.Get("secondary_id").(int)
	staySecondary := false
	if stay, ok := d.GetOk("stay_secondary"); ok {
		staySecondary = stay.(bool)
	}

	nClientPrimary, err := getNitroClient(meta.(*session.Session), primaryId)
	if err != nil {
		return fmt.Errorf("Error getting primary netscaler information ID: %d", primaryId)
	}

	nClientSecondary, err := getNitroClient(meta.(*session.Session), secondaryId)
	if err != nil {
		return fmt.Errorf("Error getting secondary netscaler information ID: %d", secondaryId)
	}

	err = configureHA(nClientPrimary, nClientSecondary, staySecondary)
	if err != nil {
		return fmt.Errorf("Error configuration HA %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%d:%d", primaryId, secondaryId))

	log.Printf("[INFO] Netscaler HA ID: %s", d.Id())

	return resourceSoftLayerLbVpxHaRead(d, meta)
}

func resourceSoftLayerLbVpxHaRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceSoftLayerLbVpxHaDelete(d *schema.ResourceData, meta interface{}) error {
	primaryId, secondaryId, err := parseHAId(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting HA %s", err.Error())
	}
	nClientPrimary, err := getNitroClient(meta.(*session.Session), primaryId)
	if err != nil {
		return fmt.Errorf("Error getting primary netscaler information ID: %d", primaryId)
	}
	nClientSecondary, err := getNitroClient(meta.(*session.Session), secondaryId)
	if err != nil {
		return fmt.Errorf("Error getting secondary netscaler information ID: %d", secondaryId)
	}

	secondaryPassword := nClientSecondary.Password
	nClientSecondary.Password = nClientPrimary.Password
	err = deleteHA(nClientPrimary, nClientSecondary)
	if err != nil {
		return fmt.Errorf("Error deleting HA %s", err.Error())
	}

	// Restore password of the secondary VPX
	systemuserReq := dt.SystemuserReq{
		Systemuser: &dt.Systemuser{
			Username: op.String("root"),
			Password: op.String(secondaryPassword),
		},
	}
	err = nClientSecondary.Update(&systemuserReq)
	if err != nil {
		return err
	}

	return nil
}

func resourceSoftLayerLbVpxHaExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}
