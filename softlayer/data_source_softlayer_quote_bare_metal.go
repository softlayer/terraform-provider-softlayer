package softlayer

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/sl"
	"strconv"
	"strings"
)

func dataSourceSoftLayerQuoteBareMetal() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSoftLayerQuoteBareMetalRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The internal id of the quote for bare metal server",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"name": {
				Description: "The name of this quote",
				Type:        schema.TypeString,
				Required:    true,
			},

			"datacenter": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_speed": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"private_network_only": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"tcp_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"package_key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"process_key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"os_key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_key_names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"redundant_network": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"unbonded_network": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"public_bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Custom bare metal server only
			"redundant_power_supply": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			// Custom bare metal server only - Order multiple RAID groups
			"storage_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"array_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"hard_drives": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Computed: true,
						},
						"array_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"partition_template_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceSoftLayerQuoteBareMetalRead(d *schema.ResourceData, meta interface{}) error {
	sess := meta.(ProviderConfig).SoftLayerSession()
	service := services.GetAccountService(sess)

	name := d.Get("name").(string)

	quotes, err := service.
		Mask("id,name,order[items[storageGroups,item],orderTopLevelItems]").
		GetActiveQuotes()
	if err != nil {
		return fmt.Errorf("Error looking up quote [%s]: %s", name, err)
	}

	for _, quote := range quotes {
		if quote.Name != nil && *quote.Name == name {
			// Build a bare metal template from the quote.
			order, err := services.GetBillingOrderQuoteService(sess).
				Id(*quote.Id).GetRecalculatedOrderContainer(nil, sl.Bool(false))
			if err != nil {
				return fmt.Errorf(
					"Encountered problem trying to get the bare metal order template from quote: %s", err)
			}
			bmPackage, err := services.GetProductPackageService(sess).
				Id(*order.PackageId).GetObject()
			if err != nil {
				return fmt.Errorf("Unable to find a package name from quote: %s", err)
			}
			if len(order.StorageGroups) > 0 {
				storageGroups := make([]map[string]interface{}, 0, len(order.StorageGroups))
				for _, sg := range order.StorageGroups {
					storageGroup := make(map[string]interface{})
					storageGroup["array_type_id"] = *sg.ArrayTypeId
					storageGroup["array_size"] = sl.Get(sg.ArraySize, 0)
					storageGroup["partition_template_id"] = sl.Get(sg.PartitionTemplateId, 0)
					storageGroup["hard_drives"] = sg.HardDrives
					storageGroups = append(storageGroups, storageGroup)
				}

				d.Set("storage_groups", storageGroups)
			}
			locationId, err := strconv.Atoi(*order.Location)
			if err != nil {
				return fmt.Errorf("Location Id should be an integer: %s", *order.Location)
			}
			dc, err := services.GetLocationDatacenterService(sess).Id(locationId).GetObject()
			if err != nil {
				return fmt.Errorf("Unable to find a data center from quote: %s", err)
			}
			d.Set("datacenter", *dc.Name)
			d.SetId(fmt.Sprintf("%d", *quote.Id))
			d.Set("package_key_name", *bmPackage.KeyName)
			d.Set("redundant_power_supply", false)
			diskMap := make(map[int]string)

			for _, item := range quote.Order.Items {
				switch *item.CategoryCode {
				case "server":
					d.Set("process_key_name", *item.Item.KeyName)
				case "os":
					d.Set("os_key_name", *item.Item.KeyName)
				case "ram":
					d.Set("memory", int(*item.Item.Capacity))
				case "bandwidth":
					d.Set("public_bandwidth", int(*item.Item.Capacity))
				case "port_speed":
					d.Set("network_speed", int(*item.Item.Capacity))
					d.Set("unbonded_network", false)
					d.Set("redundant_network", false)
					d.Set("private_network_only", false)
					if strings.Contains(*item.Item.KeyName, "UNBONDED") {
						d.Set("unbonded_network", true)
					}
					if strings.Contains(*item.Item.KeyName, "REDUNDANT") {
						d.Set("redundant_network", true)
					}
					if !strings.Contains(*item.Item.KeyName, "PUBLIC") {
						d.Set("private_network_only", true)
					}
				case "power_supply":
					d.Set("redundant_power_supply", true)
				case "monitoring":
					d.Set("tcp_monitoring", false)
					if strings.Contains(*item.Item.KeyName, "TCP") {
						d.Set("tcp_monitoring", true)
					}
				}

				if strings.HasPrefix(*item.CategoryCode, "disk") {
					diskIndex, err := strconv.Atoi(strings.Split(*item.CategoryCode, "disk")[1])
					if err == nil {
						diskMap[diskIndex] = *item.Item.KeyName
					}
				}

			}
			numberOfDisk := len(diskMap)
			if numberOfDisk > 0 {
				disks := make([]string, numberOfDisk, numberOfDisk)
				for i := 0; i < numberOfDisk; i++ {
					if len(diskMap[i]) > 0 {
						disks[i] = diskMap[i]
					} else {
						return fmt.Errorf("Unable to retrieve disk information.")
					}
				}
				d.Set("disk_key_names", disks)
			}
			return nil
		}
	}

	return fmt.Errorf("Could not find quote with name [%s]", name)
}
