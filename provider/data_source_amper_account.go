package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/spirius/terraform-provider-amper/amper"
)

func dataSourceAmperAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAmperAccountRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"short_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
		},
	}
}

func dataSourceAmperAccountRead(d *schema.ResourceData, meta interface{}) error {
	cc := meta.(*amper.Kernel)

	account := &amper.Account{
		ID:        d.Get("account_id").(string),
		Name:      d.Get("name").(string),
		ShortName: d.Get("short_name").(string),
	}

	d.SetId(d.Get("account_id").(string))

	return cc.AddAccount(account)
}
