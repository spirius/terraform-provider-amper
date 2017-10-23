package amper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/Masterminds/sprig"
)

type ServiceRoleTemplate struct {
	Name               string
	Template           *string
	AssumeRoleTemplate *string
}

type PolicyTemplate struct {
	sync.Mutex

	amper     *Kernel
	container *Container
	notFound  bool

	// Key is the uniqie identifier of policy template
	Key string

	// Template is pointer to template's content.
	// If it's nil, template will be fetched from StateBucket
	Template *string

	// Vars contains list of required variables for rendering this template
	Vars []string

	// Scope defines AWS IAM services, covered by this template.
	// Formate is same, as for Action field in IAM Policy Statement.
	Scope []string

	ServiceRole *ServiceRoleTemplate
}

func (pt *PolicyTemplate) render(name string, txt *string, vars map[string]interface{}) (*IAMPolicyDoc, error) {
	tpl, err := template.
		New(name).
		Funcs(sprig.TxtFuncMap()).
		Parse(*txt)

	if err != nil {
		return nil, err
	}

	var policyBuf bytes.Buffer

	if err = tpl.Execute(&policyBuf, vars); err != nil {
		return nil, err
	}

	a := &IAMPolicyDoc{}

	if err := json.Unmarshal(policyBuf.Bytes(), a); err != nil {
		return nil, err
	}

	return a, nil
}

func (pt *PolicyTemplate) renderTemplate(c *Container, account *Account, vars map[string]string) (*IAMPolicyDoc, error) {
	pt.Lock()
	defer pt.Unlock()

	if pt.notFound {
		return nil, nil
	}

	if pt.Template == nil {
		tpl, err := pt.fetchTemplate()

		if err != nil {
			return nil, err
		}

		if tpl == nil {
			pt.notFound = true
			return nil, nil
		}

		pt.Template = tpl
	}

	templateVars := map[string]interface{}{
		"container": c,
		"account":   account,
		"vars":      vars,
	}

	return pt.render(fmt.Sprintf("container=%s,template=%s,account=%s", c.ID, pt.Key, account.Name), pt.Template, templateVars)
}

func (pt *PolicyTemplate) renderServiceRole(c *Container, account *Account, vars map[string]string) (*IAMPolicyDoc, error) {
	if pt.ServiceRole == nil {
		return nil, nil
	}

	templateVars := map[string]interface{}{
		"container": c,
		"account":   account,
		"vars":      vars,
	}

	return pt.render(fmt.Sprintf("service_role:container=%s,template=%s,account=%s", c.ID, pt.Key, account.Name), pt.ServiceRole.Template, templateVars)
}

func (pt *PolicyTemplate) renderServiceAssumeRole(c *Container, account *Account, vars map[string]string) (*IAMPolicyDoc, error) {
	if pt.ServiceRole == nil {
		return nil, nil
	}

	templateVars := map[string]interface{}{
		"container": c,
		"account":   account,
		"vars":      vars,
	}

	return pt.render(fmt.Sprintf("service_role:container=%s,template=%s,account=%s", c.ID, pt.Key, account.Name), pt.ServiceRole.AssumeRoleTemplate, templateVars)
}

func (pt *PolicyTemplate) fetchTemplate() (*string, error) {
	var key = fmt.Sprintf(pt.amper.KeyFormat, pt.container.ID, pt.Key)

	objInfo, err := pt.amper.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(pt.amper.StateBucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed reading S3 object '%s': %s", key, err)
	}

	if *objInfo.ContentLength > 65536 {
		return nil, fmt.Errorf("failed reading S3 object '%s', too big %dbytes", key, *objInfo.ContentLength)
	}

	out, err := pt.amper.S3.GetObject(&s3.GetObjectInput{
		Bucket:    aws.String(pt.amper.StateBucket),
		Key:       aws.String(key),
		VersionId: objInfo.VersionId,
	})

	if err != nil {
		return nil, fmt.Errorf("failed reading S3 object '%s': %s", key, err)
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(out.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed reading content of S3 object '%s': %s", key, err)
	}

	return aws.String(buf.String()), nil
}
