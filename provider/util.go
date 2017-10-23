package provider

import (
	"fmt"
	"regexp"
	"strings"
)

var reservedWords = []string{
	// AWS service names in ARN
	"apigateway",
	"appstream",
	"artifact",
	"autoscaling",
	"aws-portal",
	"acm",
	"clouddirectory",
	"cloudformation",
	"cloudfront",
	"cloudhsm",
	"cloudsearch",
	"cloudtrail",
	"cloudwatch",
	"events",
	"logs",
	"codebuild",
	"codecommit",
	"codedeploy",
	"codepipeline",
	"codestar",
	"cognito-idp",
	"cognito-identity",
	"cognito-sync",
	"config",
	"datapipeline",
	"dms",
	"devicefarm",
	"directconnect",
	"ds",
	"dynamodb",
	"ec2",
	"ecr",
	"ecs",
	"ssm",
	"elasticbeanstalk",
	"elasticfilesystem",
	"elasticloadbalancing",
	"elasticmapreduce",
	"elastictranscoder",
	"elasticache",
	"es",
	"gamelift",
	"glacier",
	"glue",
	"health",
	"iam",
	"importexport",
	"inspector",
	"iot",
	"kms",
	"kinesisanalytics",
	"firehose",
	"kinesis",
	"lambda",
	"lightsail",
	"machinelearning",
	"aws-marketplace",
	"aws-marketplace-management",
	"mobileanalytics",
	"mobilehub",
	"opsworks",
	"opsworks-cm",
	"organizations",
	"polly",
	"redshift",
	"rds",
	"route53",
	"route53domains",
	"sts",
	"servicecatalog",
	"ses",
	"sns",
	"sqs",
	"s3",
	"swf",
	"sdb",
	"states",
	"storagegateway",
	"support",
	"trustedadvisor",
	"ec2",
	"waf",
	"workdocs",
	"workmail",
	"workspaces",

	// Custom reserved words
	"ip",
}

var reservedWordsRegexp *regexp.Regexp

func init() {
	w := strings.Join(reservedWords, "|")
	reservedWordsRegexp = regexp.MustCompile(fmt.Sprintf(`(-(?:%s)-|^(?:%s)-|-(?:%s)$)`, w, w, w))
}

func validateContainerName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	ws, errors = validateName(v, k)

	matches := reservedWordsRegexp.FindStringSubmatch(value)

	if len(matches) > 0 {
		errors = append(errors, fmt.Errorf(
			"reserved word '%s' used in %q: %q", matches[1], k, value))
	}
	return
}

func validateName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}
	if !regexp.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lower-case alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	if regexp.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen: %q", k, value))
	}
	if regexp.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain multiple hyphens in sequence: %q", k, value))
	}
	return
}
