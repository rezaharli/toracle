package ldap

type Scope uint8

const (
	ScopeBaseObject   Scope = 0
	ScopeSingleLevel  Scope = 1
	ScopeWholeSubtree Scope = 2
)
