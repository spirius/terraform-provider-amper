package provider

import (
	"fmt"
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
				Optional:     true,
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
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"template": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"assume_role_template": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"const": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: func(v interface{}, k string) ([]string, []error) {
								switch v.(string) {
								case "string", "list", "map":
								default:
									return nil, []error{fmt.Errorf("type can be string, list or map")}
								}
								return nil, nil
							},
						},
						"string": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"list": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"map": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     schema.TypeString,
							Default:  nil,
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

	var containerId string

	if attr, ok := d.GetOk("container_id"); ok {
		containerId = attr.(string)
	}

	pt.Consts = make(map[string]interface{})

	consts := d.Get("const").(*schema.Set).List()

	for _, c := range consts {
		raw := c.(map[string]interface{})

		pt.Consts[raw["name"].(string)] = raw[raw["type"].(string)]
	}

	return cc.AddPolicyTemplate(containerId, pt)
}
