package netspy

import (
	"net/url"
	"strings"
)

type Rule struct {
	allow []string
	deny  []string
}

func NewRule(allow []string, deny []string) Rule {
	return Rule{allow: allow, deny: deny}
}

func (rule *Rule) Pass(parsedURL *url.URL) bool {
	path := parsedURL.Path
	for _, allow := range rule.allow {
		if !strings.Contains(path, allow) {
			return false
		}
	}

	for _, deny := range rule.deny {
		if strings.Contains(path, deny) {
			return false
		}
	}

	return true
}
