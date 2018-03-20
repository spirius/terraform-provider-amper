package provider

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testFcConfig = `
data "amper_fc" "test" {
  input = "{\"test\": 123}"

  filter {
    name = "json"
    args = ["asd", "123"]
  }

  filter {
    name = "json"
  }
}

output "out" {
  value = "${data.amper_fc.test.output}"
}
`

func TestFc(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"amper": testProvider,
		},
		Steps: []resource.TestStep{
			{
				Config: testFcConfig,
			},
		},
	})
}
