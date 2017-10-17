package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/spirius/terraform-provider-amper/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider})
}
