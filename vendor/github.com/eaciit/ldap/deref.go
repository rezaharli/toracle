package ldap

type Deref uint8

const (
	NeverDerefAliases   Deref = 0
	DerefInSearching    Deref = 1
	DerefFindingBaseObj Deref = 2
	DerefAlways         Deref = 3
)
