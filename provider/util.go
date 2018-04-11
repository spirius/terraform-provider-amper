package provider

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var reservedWords = []string{
	"ip",
	"c",
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

func resourceGetStringListFromList(attrs []interface{}) (res []string) {
	res = make([]string, 0, len(attrs))

	for _, v := range attrs {
		res = append(res, v.(string))
	}

	return
}

func getContentShas(data []byte) (string, string, string) {
	h := sha1.New()
	h.Write(data)
	sha1 := hex.EncodeToString(h.Sum(nil))

	h256 := sha256.New()
	h256.Write(data)
	shaSum := h256.Sum(nil)
	sha256base64 := base64.StdEncoding.EncodeToString(shaSum[:])

	md5 := md5.New()
	md5.Write(data)
	md5Sum := hex.EncodeToString(md5.Sum(nil))

	return sha1, sha256base64, md5Sum
}

func getContentSha256Base64(data string) string {
	h256 := sha256.New()
	h256.Write([]byte(data))
	shaSum := h256.Sum(nil)
	return base64.StdEncoding.EncodeToString(shaSum[:])
}
