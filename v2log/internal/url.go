package internal

import "strings"

var company = [...]string{"1"}
var nsfw = [...]string{}

func InCompany(domain string) bool {
	for i := range company {
		if domain == company[i] {
			return true
		}
		if company[i][0] == '*' && strings.HasSuffix(domain, company[i][1:]) {
			return true
		}
	}
	return false
}
func InNSFW(domain string) bool {
	for i := range nsfw {
		if domain == nsfw[i] {
			return true
		}
		if nsfw[i][0] == '*' && strings.HasSuffix(domain, nsfw[i][1:]) {
			return true
		}
	}
	return false
}
