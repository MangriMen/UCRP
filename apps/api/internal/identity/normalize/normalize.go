package normalize

import (
	"regexp"
	"strings"
)

var (
	digitsOnly = regexp.MustCompile(`\D`)
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func Phone(s string) string {
	d := digitsOnly.ReplaceAllString(strings.TrimSpace(s), "")
	if d == "" {
		return ""
	}
	if len(d) == 11 && (strings.HasPrefix(d, "8") || strings.HasPrefix(d, "7")) {
		return "7" + d[1:]
	}
	if len(d) == 10 {
		return "7" + d
	}
	return d
}

func Email(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func Card(s string) string {
	return digitsOnly.ReplaceAllString(strings.TrimSpace(s), "")
}

func ValidEmail(s string) bool {
	return emailRe.MatchString(s)
}
