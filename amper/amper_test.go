package amper

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ = fmt.Printf

func getS3(t *testing.T) *s3.S3 {
	var awsConfig = &aws.Config{
		Region: aws.String("eu-central-1"),
		Credentials: credentials.NewChainCredentials([]credentials.Provider{
			&credentials.SharedCredentialsProvider{
				Profile: "sandbox-direct",
			},
		}),
	}

	sess, err := session.NewSession(awsConfig)

	if err != nil {
		t.Fatal(err)
	}

	return s3.New(sess)
}

func TestStaticPolicy(t *testing.T) {
	var (
		amper        *Kernel
		root, c1     *Container
		IAM, SQS *PolicyTemplate
		err          error
		acc1, acc2   *Account
	)

	amper = NewKernel(getS3(t), "vahe-test-bucket-sandbox")

	acc1 = &Account{
		ID:        "023123123",
		Name:      "sub-account-1",
		ShortName: "sa1",
	}

	if err = amper.AddAccount(acc1); err != nil {
		t.Fatal(err)
	}

	acc2 = &Account{
		ID:        "0123123",
		Name:      "sub-account-2",
		ShortName: "sa2",
	}

	if err = amper.AddAccount(acc2); err != nil {
		t.Fatal(err)
	}

	if root, err = amper.NewContainer("root"); err != nil {
		t.Fatal(err)
	}

	IAM = &PolicyTemplate{
		Key:   "iam",
		Scope: []string{"iam:*", "ec2:*"},
	}
	IAM.Template = aws.String(`{
  "Statement": [{
    "Effect": "Allow",
    "Action": "*",
    "Resource": "aws:arn:{{ .account.ID }}",
    "Condition": {
      "IpAddress": {
        "aws:SourceIp": [
          "203.0.113.0/24",
          "2001:DB8:1234:5678::/64"
        ]
      }
    }
  }]
}`)

	if err = root.AddPolicyTemplate(IAM); err != nil {
		t.Fatal(err)
	}

	SQS = &PolicyTemplate{
		Key:   "sqs",
		Scope: []string{"ec2:*", "beh:*"},
	}
	SQS.Template = aws.String(`{
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sqs:*"],
    "Resource": ["*"]
  }]
}`)

	if err = root.AddPolicyTemplate(SQS); err != nil {
		t.Fatal(err)
	}

	if c1, err = amper.NewContainer("c1"); err != nil {
		t.Fatal(err)
	}

	if _, err = c1.AddAttachment("iam", "sub-account-2", nil); err != nil {
		t.Fatal(err)
	}

	if _, err = c1.AddAttachment("sqs", "sub-account-2", nil); err != nil {
		t.Fatal(err)
	}

	policy, err, _ := c1.Policy()

	if err != nil {
		t.Fatal(err)
	}

	policy.dump()
}

func TestS3PolicyTemplate(t *testing.T) {
	var (
		amper        *Kernel
		acc          *Account
		root, c1     *Container
		err          error
		IAM, SQS *PolicyTemplate
	)
	amper = NewKernel(getS3(t), "vahe-test-bucket-sandbox")

	acc = &Account{
		ID:        "023123123",
		Name:      "sub-account-1",
		ShortName: "sa1",
	}

	if err = amper.AddAccount(acc); err != nil {
		t.Fatal(err)
	}

	if root, err = amper.NewContainer("root"); err != nil {
		t.Fatal(err)
	}

	IAM = &PolicyTemplate{
		Key:   "iam",
		Scope: []string{"iam:*", "ec2:*"},
	}

	if err = root.AddPolicyTemplate(IAM); err != nil {
		t.Fatal(err)
	}

	SQS = &PolicyTemplate{
		Key:   "sqs",
		Scope: []string{"sqs:*"},
	}
	SQS.Template = aws.String(`{
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "sqs:*Visibility*",
        "sqs:*Permission*",
        "sqs:*Queue",
        "sqs:ListDeadLetterSourceQueues",
        "sqs:*Message*"
      ],
      "NotResource": ["arn:aws:sqs:*:*:{{ .container.ID }}_*"]
    }
  ]
}`)

	if err = root.AddPolicyTemplate(SQS); err != nil {
		t.Fatal(err)
	}

	if c1, err = amper.NewContainer("c1"); err != nil {
		t.Fatal(err)
	}

	if _, err = c1.AddAttachment("iam", "sub-account-1", nil); err != nil {
		t.Fatal(err)
	}

	if _, err = c1.AddAttachment("sqs", "sub-account-1", nil); err != nil {
		t.Fatal(err)
	}

	policy, err, _ := c1.Policy()

	if err != nil {
		t.Fatal(err)
	}

	policy.dump()
}
