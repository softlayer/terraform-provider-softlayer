package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/softlayer/terraform-provider-softlayer/softlayer"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: softlayer.Provider,
	})
}
