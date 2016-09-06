package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.ibm.com/emerging-tech/terraform-provider-softlayer/softlayer"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: softlayer.Provider,
	})
}
