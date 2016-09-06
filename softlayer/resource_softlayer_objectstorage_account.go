package softlayer

import (
	"fmt"
	"log"

	datatypes "github.com/TheWeatherCompany/softlayer-go/data_types"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"time"
)

func resourceSoftLayerObjectStorageAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerObjectStorageAccountCreate,
		Read:   resourceSoftLayerObjectStorageAccountRead,
		Update: resourceSoftLayerObjectStorageAccountUpdate,
		Delete: resourceSoftLayerObjectStorageAccountDelete,
		Exists: resourceSoftLayerObjectStorageAccountExists,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_note": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSoftLayerObjectStorageAccountCreate(d *schema.ResourceData, meta interface{}) error {
	accountService := meta.(*Client).accountService

	// Check if an object storage account exists
	objectStorageAccounts, err := accountService.GetHubNetworkStorage()
	if err != nil {
		return fmt.Errorf("resource_softlayer_objectstorage_account: Error on create: %s", err)
	}

	if len(objectStorageAccounts) == 0 {
		// Order the account
		productOrderService := meta.(*Client).productOrderService

		receipt, err := productOrderService.PlaceOrder(datatypes.SoftLayer_Container_Product_Order{
			ComplexType: "SoftLayer_Container_Product_Order_Network_Storage_Hub",
			Quantity:    1,
			PackageId:   0,
			Prices: []datatypes.SoftLayer_Product_Item_Price{
				{Id: 30920},
			},
		})
		if err != nil {
			return fmt.Errorf(
				"resource_softlayer_objectstorage_account: Error ordering account: %s", err)
		}

		// Wait for the object storage account order to complete.
		billingOrderItem, err := WaitForOrderCompletion(&receipt, meta)
		if err != nil {
			return fmt.Errorf(
				"Error waiting for object storage account order (%d) to complete: %s", receipt.OrderId, err)
		}

		// Get accountName using filter on hub network storage
		objectStorageAccounts, err = accountService.GetHubNetworkStorageByFilter(
			fmt.Sprintf(`{"billingItem":{"id":{"operation":%d}}}`, billingOrderItem.BillingItem.Id),
		)
		if err != nil {
			return fmt.Errorf("resource_softlayer_objectstorage_account: Error on retrieving new: %s", err)
		}

		if len(objectStorageAccounts) == 0 {
			return fmt.Errorf("resource_softlayer_objectstorage_account: Failed to create object storage account.")
		}
	}

	// Get account name and set as the Id
	d.SetId(objectStorageAccounts[0].Username)
	d.Set("name", objectStorageAccounts[0].Username)

	return nil
}

func WaitForOrderCompletion(receipt *datatypes.SoftLayer_Container_Product_Order_Receipt, meta interface{}) (datatypes.SoftLayer_Billing_Order_Item, error) {
	log.Printf("Waiting for billing order %d to have zero active transactions", receipt.OrderId)
	var billingOrderItem datatypes.SoftLayer_Billing_Order_Item

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "in progress"},
		Target:  []string{"complete"},
		Refresh: func() (interface{}, string, error) {
			billingItemService := meta.(*Client).billingItemService
			var err error
			var completed bool
			completed, billingOrderItem, err = billingItemService.CheckOrderStatus(receipt, []string{})
			if err != nil {
				return nil, "", err
			}
			if completed {
				return billingOrderItem, "complete", nil
			} else {
				return billingOrderItem, "in progress", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return billingOrderItem, err
}

func resourceSoftLayerObjectStorageAccountRead(d *schema.ResourceData, meta interface{}) error {
	accountService := meta.(*Client).accountService
	accountName := d.Id()
	d.Set("name", accountName)

	// Check if an object storage account exists
	objectStorageAccounts, err := accountService.GetHubNetworkStorageByFilter(
		fmt.Sprintf(`{"username":{"operation":"%s"}}`, accountName))
	if err != nil {
		return fmt.Errorf("resource_softlayer_objectstorage_account: Error on Read: %s", err)
	}

	for _, objectStorageAccount := range objectStorageAccounts {
		if objectStorageAccount.Username == accountName {
			return nil
		}
	}

	return fmt.Errorf("resource_softlayer_objectstorage_account: Could not find account %s", accountName)
}

func resourceSoftLayerObjectStorageAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	// Nothing to update for now. Not supported.
	return nil
}

func resourceSoftLayerObjectStorageAccountDelete(d *schema.ResourceData, meta interface{}) error {
	// Delete is not supported for now.
	return nil
}

func resourceSoftLayerObjectStorageAccountExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceSoftLayerObjectStorageAccountRead(d, meta)

	return err == nil, err
}
