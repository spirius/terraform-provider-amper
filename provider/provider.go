package provider

import (
	"fmt"

	"github.com/spirius/terraform-provider-amper/amper"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	
	tfaws "github.com/terraform-providers/terraform-provider-aws/aws"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"profile": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				InputDefault: "us-east-1",
			},

			"assume_role": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"state_bucket": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "State file bucket",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"amper_account":         dataSourceAmperAccount(),
			"amper_container":       dataSourceAmperContainer(),
			"amper_policy_template": dataSourceAmperPolicyTemplate(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &tfaws.Config{
		AccessKey: d.Get("access_key").(string),
		SecretKey: d.Get("secret_key").(string),
		Profile:   d.Get("profile").(string),
		Region:    d.Get("region").(string),
	}

	assumeRoleList := d.Get("assume_role").(*schema.Set).List()
	if len(assumeRoleList) == 1 {
		assumeRole := assumeRoleList[0].(map[string]interface{})
		config.AssumeRoleARN = assumeRole["role_arn"].(string)
		config.AssumeRoleSessionName = assumeRole["session_name"].(string)
		config.AssumeRoleExternalID = assumeRole["external_id"].(string)

		if v := assumeRole["policy"].(string); v != "" {
			config.AssumeRolePolicy = v
		}
	}

	tfcreds, err := tfaws.GetCredentials(config)

	if err != nil {
		return nil, err
	}
	
	creds := (*credentials.Credentials)(tfcreds)

	if _, err = creds.Get(); err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoCredentialProviders" {
			return nil, fmt.Errorf("No valid credential sources found for AWS Provider")
		}

		return nil, fmt.Errorf("Error loading credentials for AWS Provider: %s", err)
	}

	awsConfig := &aws.Config{
		Credentials: creds,
		Region:      aws.String(config.Region),
	}

	sess, err := session.NewSession(awsConfig)

	if err != nil {
		return nil, errwrap.Wrapf("Error creating AWS session: {{err}}", err)
	}

	return amper.NewKernel(s3.New(sess), d.Get("state_bucket").(string)), nil
}
