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

func resourceSoftLayerBasicMonitor() *schema.Resource {
	return &schema.Resource{
		Create:   resourceSoftLayerBasicMonitorCreate,
		Read:     resourceSoftLayerBasicMonitorRead,
		Update:   resourceSoftLayerBasicMonitorUpdate,
		Delete:   resourceSoftLayerBasicMonitorDelete,
		Exists:   resourceSoftLayerBasicMonitorExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"guest_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"query_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"response_action_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"wait_cycles": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"notified_users": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceSoftLayerBasicMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	virtualGuestService := services.GetVirtualGuestService(sess)
	monitorService := services.GetNetworkMonitorVersion1QueryHostService(sess)

	guestId := d.Get("guest_id").(int)
	ipAddress := d.Get("ip_address").(string)
	if ipAddress == "" {
		virtualGuest, err := virtualGuestService.Id(guestId).GetObject()
		if err != nil {
			return fmt.Errorf("Error looking up virtual guest %d: %s", guestId, err)
		}

		if virtualGuest.PrimaryIpAddress == nil {
			return fmt.Errorf(
				"No primary ip address found for virtual guest %d. Please specify it.", guestId)
		}

		ipAddress = *virtualGuest.PrimaryIpAddress
	}

	// Build up our creation options
	opts := datatypes.Network_Monitor_Version1_Query_Host{
		GuestId:          &guestId,
		IpAddress:        &ipAddress,
		QueryTypeId:      sl.Int(d.Get("query_type_id").(int)),
		ResponseActionId: sl.Int(d.Get("response_action_id").(int)),
	}
	if wait_cycles, ok := d.GetOk("wait_cycles"); ok {
		opts.WaitCycles = sl.Int(wait_cycles.(int))
	}

	// Create a monitor
	res, err := monitorService.CreateObject(&opts)
	if err != nil {
		return fmt.Errorf("Error creating Basic Monitor : %s", err)
	}

	d.SetId(strconv.Itoa(*res.Id))
	log.Printf("[INFO] Basic Monitor Id: %d", *res.Id)

	err = createNotifications(d, meta, guestId)
	if err != nil {
		return err
	}

	return resourceSoftLayerBasicMonitorRead(d, meta)
}

func createNotifications(d *schema.ResourceData, meta interface{}, guestId int) error {
	sess := meta.(*session.Session)
	virtualGuestService := services.GetVirtualGuestService(sess)
	notificationService := services.GetUserCustomerNotificationVirtualGuestService(sess)

	// Create a user notification
	// This represents a link between a monitored guest instance and a user account
	notificationLinks, err := virtualGuestService.Id(guestId).GetMonitoringUserNotification()
	if err != nil {
		return fmt.Errorf("Error looking up user notifications for virtual guest %d", guestId)
	}

	userNotificationOpts := datatypes.User_Customer_Notification_Virtual_Guest{
		GuestId: &guestId,
	}
	notifiedUsers := d.Get("notified_users").([]interface{})
	for _, userId := range notifiedUsers {
		userNotificationOpts.UserId = sl.Int(userId.(int))

		// Don't create the notification object if one already exists for the same user and vm
		if !notificationExists(notificationLinks, userId.(int)) {
			notificationService.CreateObject(&userNotificationOpts)
		}
	}

	return nil
}

func notificationExists(notificationLinks []datatypes.User_Customer_Notification_Virtual_Guest, userId int) bool {
	for _, link := range notificationLinks {
		if *link.UserId == userId {
			return true
		}
	}

	return false
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func resourceSoftLayerBasicMonitorRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkMonitorVersion1QueryHostService(sess)
	virtualGuestService := services.GetVirtualGuestService(sess)

	basicMonitorId, _ := strconv.Atoi(d.Id())

	basicMonitor, err := service.Id(basicMonitorId).GetObject()
	if err != nil {
		// If the monitor is somehow already destroyed, mark as
		// succesfully gone
		if apiErr, ok := err.(sl.Error); ok && apiErr.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Basic Monitor : %s", err)
	}

	guestId := *basicMonitor.GuestId

	d.Set("id", *basicMonitor.Id)
	d.Set("guest_id", guestId)
	d.Set("ip_address", strings.TrimSpace(*basicMonitor.IpAddress))
	d.Set("query_type_id", *basicMonitor.QueryTypeId)
	d.Set("response_action_id", *basicMonitor.ResponseActionId)
	d.Set("wait_cycles", *basicMonitor.WaitCycles)

	notificationLinks, err := virtualGuestService.Id(guestId).GetMonitoringUserNotification()
	if err != nil {
		return fmt.Errorf("Error looking up user notifications for virtual guest %d", guestId)
	}

	notificationUserIds := []int{}
	for _, notificationLink := range notificationLinks {
		if !contains(notificationUserIds, *notificationLink.UserId) {
			notificationUserIds = append(notificationUserIds, *notificationLink.UserId)
		}
	}
	d.Set("notified_users", notificationUserIds)

	return nil
}

func resourceSoftLayerBasicMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkMonitorVersion1QueryHostService(sess)

	basicMonitorId, _ := strconv.Atoi(d.Id())
	guestId := d.Get("guest_id").(int)

	basicMonitor, err := service.Id(basicMonitorId).GetObject()
	if err != nil {
		return fmt.Errorf("Error retrieving Basic Monitor : %s", err)
	}
	if d.HasChange("query_type_id") {
		basicMonitor.QueryTypeId = sl.Int(d.Get("query_type_id").(int))
	}
	if d.HasChange("response_action_id") {
		basicMonitor.ResponseActionId = sl.Int(d.Get("response_action_id").(int))
	}
	if d.HasChange("wait_cycles") {
		basicMonitor.WaitCycles = sl.Int(d.Get("wait_cycles").(int))
	}

	_, err = service.Id(basicMonitorId).EditObject(&basicMonitor)
	if err != nil {
		return fmt.Errorf("Error editing Basic Monitor : %s", err)
	}

	// Will only create notification objects for user/vm relationships that
	// don't exist yet.
	err = createNotifications(d, meta, guestId)
	if err != nil {
		return err
	}

	return resourceSoftLayerBasicMonitorRead(d, meta)
}

func resourceSoftLayerBasicMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(*session.Session)
	service := services.GetNetworkMonitorVersion1QueryHostService(sess)

	// Delete the basic monitor
	id, err := strconv.Atoi(d.Id())

	log.Printf("[INFO] Deleting Basic Monitor : %d", id)
	_, err = service.Id(id).DeleteObject()
	if err != nil {
		return fmt.Errorf("Error deleting Basic Monitor : %s", err)
	}

	d.SetId("")
	return nil
}

func resourceSoftLayerBasicMonitorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	sess := meta.(*session.Session)
	service := services.GetNetworkMonitorVersion1QueryHostService(sess)

	basicMonitorId, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	result, err := service.Id(basicMonitorId).GetObject()
	return err == nil && *result.Id == basicMonitorId, nil
}
