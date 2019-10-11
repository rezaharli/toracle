package ldap

import (
	"strings"
)

// DnReplace replaces runes in AssertionValues as defined in 
// RFC 4514 [https://www.ietf.org/rfc/rfc4514.txt]
func DnReplace(value string) string {
	if value == "" {
		return ""
	}

	r := strings.NewReplacer(dnReplacements...)
	value = r.Replace(value)
	
	// escape leading space and #
	if value[0] == ' ' || value[0] == '#' {
		value = "\\" + value
	}
	
	// escape space and # at the end
	if value[len(value)-1] == ' ' || value[len(value)-1] == '#' {
		value = value[0:len(value)-1] + "\\" + string(value[len(value)-1])
	}
	
	return value
}

// FilterReplace replaces runes in AssertionValues as defined in 
// RFC 4515 [https://www.ietf.org/rfc/rfc4515.txt]
func FilterReplace(value string) string {
	r := strings.NewReplacer(filterReplacements...)
	return r.Replace(value)
}

// default replacements to perform for dn attribute values
// this is used as initializing argument for a strings.Replacer
var dnReplacements = []string{
	`\`,	`\\`,
	`,`,	`\,`,
	`+`,	`\+`,
	`"`,	`\"`,
	`<`,	`\<`,
	`>`,	`\>`,
	`;`,	`\;`,
	`=`,	`\=`,
	"\000",	`\00`}	

// default replacements to perform for values used in filters
// this is used as initializing argument for a strings.Replacer
var filterReplacements = []string{
	`*`,	`\2a`,
	`(`,	`\28`,
	`)`,	`\29`,
	`\`,	`\5c`,
	"\000",	`\00`}
