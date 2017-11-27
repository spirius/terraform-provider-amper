package provider

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mitchellh/go-homedir"
	"github.com/spirius/fc"
	"io"
	"os"
)

func dataSourceAmperFc() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAmperFcRead,

		Schema: map[string]*schema.Schema{
			"input": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"input_file"},
			},
			"input_file": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"input"},
			},
			"filter": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"args": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"output": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
				StateFunc: func(in interface{}) string {
					return getContentSha256Base64(in.(string))
				},
			},
			"output_size": {
				Type:     schema.TypeInt,
				Computed: true,
				ForceNew: true,
			},
			"output_sha": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: "SHA1 checksum of output",
			},
			"output_base64sha256": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: "Base64 Encoded SHA256 checksum of output",
			},
			"output_md5": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: "MD5 of output",
			},
		},
	}
}

type ByFilterIndex []interface{}

func (a ByFilterIndex) Len() int      { return len(a) }
func (a ByFilterIndex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFilterIndex) Less(i, j int) bool {
	return a[i].(map[string]interface{})["index"].(int) < a[j].(map[string]interface{})["index"].(int)
}

func dataSourceAmperFcRead(d *schema.ResourceData, meta interface{}) (err error) {
	var input io.Reader

	if attr, ok := d.GetOk("input"); ok {
		buf := &bytes.Buffer{}
		buf.WriteString(attr.(string))
		input = buf
	} else if attr, ok := d.GetOk("input_file"); ok {
		source := attr.(string)
		path, err := homedir.Expand(source)
		if err != nil {
			return fmt.Errorf("Error expanding homedir in source (%s): %s", source, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening S3 bucket object source (%s): %s", source, err)
		}
		input = file
	} else {
		return fmt.Errorf("one of 'input', 'input_file' must be specified")
	}

	pipeline := fc.DefaultFC.NewPipeline()

	filters := d.Get("filter").([]interface{})

	if len(filters) < 2 {
		return fmt.Errorf("at least 2 filters must be specified")
	}

	input_filter := filters[0].(map[string]interface{})

	err = pipeline.SetInputFilter(input_filter["name"].(string), resourceGetStringListFromList(input_filter["args"].([]interface{}))...)

	if err != nil {
		return fmt.Errorf("cannot set input filter: %s", err)
	}

	output_filter := filters[len(filters)-1].(map[string]interface{})

	err = pipeline.SetOutputFilter(output_filter["name"].(string), resourceGetStringListFromList(output_filter["args"].([]interface{}))...)

	if err != nil {
		return fmt.Errorf("cannot set input filter: %s", err)
	}

	for i := 1; i < len(filters)-1; i++ {
		filter := filters[i].(map[string]interface{})

		pipeline.AddFilter(filter["name"].(string), resourceGetStringListFromList(filter["args"].([]interface{}))...)
	}

	var output bytes.Buffer

	if err = pipeline.Process(input, &output); err != nil {
		return fmt.Errorf("cannot process: %s", err)
	}

	sha1, base64sha256, md5 := getContentShas(output.Bytes())

	d.Set("output_sha", sha1)
	d.Set("output_base64sha256", base64sha256)
	d.Set("output_md5", md5)

	d.Set("output_size", output.Len())
	d.Set("output", output.String())
	d.SetId(sha1)

	return nil
}
