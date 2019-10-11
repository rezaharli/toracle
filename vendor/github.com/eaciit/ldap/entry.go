package ldap

import (
	"strings"
)

type Entry struct {
	DN         string
	Attributes []*EntryAttribute
}

type EntryAttribute struct {
	Name   string
	Values []string
}

/*
func (req *Entry) RecordType() uint8 {
	return EntryRecord
}
*/

func NewEntry(dn string) *Entry {
	entry := &Entry{DN: dn}
	entry.Attributes = make([]*EntryAttribute, 0)
	return entry
}

// AddAttributeValue - Add a single Attr value
// no check is done for duplicate values.
func (e *Entry) AddAttributeValue(attributeName, value string) {
	// Don't add empty values.
	if value == "" {
		return
	}
	index := e.GetAttributeIndex(attributeName)
	if index == -1 {
		eAttr := EntryAttribute{Name: attributeName, Values: []string{value}}
		e.Attributes = append(e.Attributes, &eAttr)
	} else {
		e.Attributes[index].Values = append(e.Attributes[index].Values, value)
	}
}

// AddAttributeValues - Add via a name and slice of values
// no check is done for duplicate values.
func (e *Entry) AddAttributeValues(attributeName string, values []string) {
	if len(values) == 0 {
		return
	}
	index := e.GetAttributeIndex(attributeName)
	if index == -1 {
		eAttr := &EntryAttribute{Name: attributeName, Values: values}
		e.Attributes = append(e.Attributes, eAttr)
	} else {
		e.Attributes[index].Values = append(e.Attributes[index].Values, values...)
	}
}

func (e *Entry) GetAttributeValues(attributeName string) []string {
	for _, attr := range e.Attributes {
		if strings.EqualFold(attr.Name, attributeName) {
			return attr.Values
		}
	}
	return []string{}
}

// GetAttributeValue - returning an empty string is a bad idea
// some directory servers will return empty attr values (Sunone).
// Just asking for trouble.
func (e *Entry) GetAttributeValue(attributeName string) string {
	values := e.GetAttributeValues(attributeName)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (e *Entry) GetAttributeIndex(attributeName string) int {
	for i, attr := range e.Attributes {
		if strings.EqualFold(attr.Name, attributeName) {
			return i
		}
	}
	return -1
}

// TODO: Proper LDIF writer, currently just for testing...
func (e *Entry) String() string {
	ldif := "dn: " + e.DN + "\n"
	for _, attr := range e.Attributes {
		for _, val := range attr.Values {
			ldif += attr.Name + ": " + val + "\n"
		}
	}
	return ldif
}
