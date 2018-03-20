package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

var testProvider *schema.Provider

func init() {
	testProvider = Provider().(*schema.Provider)
}
