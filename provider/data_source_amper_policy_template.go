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

	return cc.AddPolicyTemplate(d.Get("container_id").(string), pt)
}
