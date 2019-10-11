package ldap

// LDAP Modification operation codes
type ModificationCode uint8

//go:generate stringer -type=ModificationCode

const (
	ModAdd	ModificationCode      = 0
	ModDelete	ModificationCode  = 1
	ModReplace  ModificationCode	 = 2
	// Modify-Increment Extension [https://tools.ietf.org/html/rfc4525]
	ModIncrement ModificationCode = 3
)
