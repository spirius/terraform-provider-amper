package provider

import (
	"fmt"
	"regexp"
)

func validateName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}
	if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
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
	return
}
