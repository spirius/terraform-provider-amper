package provider

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testPolicyTemplateConstsConfig = `
data "amper_container" "test" {
  name = "test"
}

data "amper_policy_template" "test" {
  key = "test"

  container_id = "${data.amper_container.test.id}"

  scope = ["ec2:*"]

  template = <<EOF
{"Statement":[{
  "Sid": "{{ .testvar }}",
  "Effect": "Allow",
  "Action": "*",
  "Resource": "*"
}]}
EOF

  const {
    name = "testvar"
    type = "list"
    list = ["asd"]
  }

  const {
    name = "testvar2"
    type = "map"
    map = {
      aa = "bb"
    }
  }
}
`

func TestPolicyTemplateConsts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"amper": testProvider,
		},
		Steps: []resource.TestStep{
			{
				Config: testPolicyTemplateConstsConfig,
			},
		},
	})
}
