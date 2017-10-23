package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/spirius/terraform-provider-amper/amper"
)

func dataSourceAmperPolicyTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAmperPolicyTemplateRead,

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				//ValidateFunc: validateName,
			},
			"container_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"template": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vars": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"scope": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"service_role": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"template": {
							Type:     schema.TypeString,
							Required: true,
						},
						"assume_role_template": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAmperPolicyTemplateRead(d *schema.ResourceData, meta interface{}) error {
	cc := meta.(*amper.Kernel)

	pt := &amper.PolicyTemplate{
		Key: d.Get("key").(string),
	}

	if attr, ok := d.GetOk("template"); ok {
		pt.Template = aws.String(attr.(string))
	}

	if attr := d.Get("vars").(*schema.Set); attr.Len() > 0 {
		pt.Vars = make([]string, 0, attr.Len())

		for _, v := range attr.List() {
			pt.Vars = append(pt.Vars, v.(string))
		}
	}

	if attr := d.Get("scope").(*schema.Set); attr.Len() > 0 {
		pt.Scope = make([]string, 0, attr.Len())

		for _, v := range attr.List() {
			pt.Scope = append(pt.Scope, v.(string))
		}
	}

	d.SetId(d.Get("key").(string))

	serviceRole := d.Get("service_role").(*schema.Set).List()

	if len(serviceRole) == 1 {
		l := serviceRole[0].(map[string]interface{})

		pt.ServiceRole = &amper.ServiceRoleTemplate{
			Name:               l["name"].(string),
			Template:           aws.String(l["template"].(string)),
			AssumeRoleTemplate: aws.String(l["assume_role_template"].(string)),
		}
	}

	return cc.AddPolicyTemplate(d.Get("container_id").(string), pt)
}
