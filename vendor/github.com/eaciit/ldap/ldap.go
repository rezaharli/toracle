// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides LDAP client functions.
package ldap

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/eaciit/asn1-ber"
	"github.com/eaciit/toolkit"
)

const (
	DefaultTimeout       = 60 * time.Minute
	ResultChanBufferSize = 5 // buffer items in each chanResults default: 5
)

// Adds descriptions to an LDAP Response packet for debugging
func addLDAPDescriptions(packet *ber.Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = newError(ErrorDebugging, "Cannot process packet to add descriptions")
		}
	}()
	packet.Description = "LDAP Response"
	packet.Children[0].Description = "Message ID"

	application := ApplicationCode(packet.Children[1].Tag)
	packet.Children[1].Description = application.String()

	switch application {
	case ApplicationBindRequest:
		addRequestDescriptions(packet)
	case ApplicationBindResponse:
		addDefaultLDAPResponseDescriptions(packet)
	case ApplicationUnbindRequest:
		addRequestDescriptions(packet)
	case ApplicationSearchRequest:
		addRequestDescriptions(packet)
	case ApplicationSearchResultEntry:
		packet.Children[1].Children[0].Description = "Object Name"
		packet.Children[1].Children[1].Description = "Attributes"
		for _, child := range packet.Children[1].Children[1].Children {
			child.Description = "Attribute"
			child.Children[0].Description = "Attribute Name"
			child.Children[1].Description = "Attribute Values"
			for _, grandchild := range child.Children[1].Children {
				grandchild.Description = "Attribute Value"
			}
		}
		if len(packet.Children) == 3 {
			addControlDescriptions(packet.Children[2])
		}
	case ApplicationSearchResultDone:
		addDefaultLDAPResponseDescriptions(packet)
	case ApplicationModifyRequest:
		addRequestDescriptions(packet)
	case ApplicationModifyResponse:
	case ApplicationAddRequest:
		addRequestDescriptions(packet)
	case ApplicationAddResponse:
	case ApplicationDelRequest:
		addRequestDescriptions(packet)
	case ApplicationDelResponse:
	case ApplicationModifyDNRequest:
		addRequestDescriptions(packet)
	case ApplicationModifyDNResponse:
	case ApplicationCompareRequest:
		addRequestDescriptions(packet)
	case ApplicationCompareResponse:
	case ApplicationAbandonRequest:
		addRequestDescriptions(packet)
	case ApplicationSearchResultReference:
	case ApplicationExtendedRequest:
		addRequestDescriptions(packet)
	case ApplicationExtendedResponse:
	}

	return nil
}

func addControlDescriptions(packet *ber.Packet) {
	packet.Description = "Controls"
	for _, child := range packet.Children {
		child.Description = "Control"

		// FIXME: this is hacky, but like the original implementation in the asn1-ber packet previously used
		var descValue string
		switch t := child.Children[0].Value.(type) {
		case string:
			descValue = t
		case []byte:
			descValue = string(t)
		default:
			descValue = ""
		}

		child.Children[0].Description = fmt.Sprintf("Control Type (%v)", ControlType(descValue))
		value := child.Children[1]
		if len(child.Children) == 3 {
			child.Children[1].Description = "Criticality"
			value = child.Children[2]
		}
		value.Description = "Control Value"

		switch ControlType(descValue) {
		case ControlTypePaging:
			value.Description += " (Paging)"
			if value.Value != nil {
				value_children := ber.DecodePacket(value.Data.Bytes())
				value.Data.Truncate(0)
				value.Value = nil
				value_children.Children[1].Value = value_children.Children[1].Data.Bytes()
				value.AppendChild(value_children)
			}
			value.Children[0].Description = "Real Search Control Value"
			value.Children[0].Children[0].Description = "Paging Size"
			value.Children[0].Children[1].Description = "Cookie"
		}
	}
}

func addRequestDescriptions(packet *ber.Packet) {
	packet.Description = "LDAP Request"
	packet.Children[0].Description = "Message ID"
	packet.Children[1].Description = ApplicationCode(packet.Children[1].Tag).String()
	if len(packet.Children) == 3 {
		addControlDescriptions(packet.Children[2])
	}
}

func addDefaultLDAPResponseDescriptions(packet *ber.Packet) {
	code, ok := packet.Children[1].Children[0].Value.(int64)
	if !ok {
		log.Printf("%T\n", packet.Children[1].Children[0].Value)
		log.Println("type assertion failed in ldap.go 125")
		code = 212
	}

	resultCode := ResultCode(code)

	packet.Children[1].Children[0].Description = "Result Code (" + resultCode.String() + ")"
	packet.Children[1].Children[1].Description = "Matched DN"
	packet.Children[1].Children[2].Description = "Error Message"
	if len(packet.Children[1].Children) > 3 {
		packet.Children[1].Children[3].Description = "Referral"
	}
	if len(packet.Children) == 3 {
		addControlDescriptions(packet.Children[2])
	}
}

func DebugBinaryFile(FileName string) error {
	file, err := ioutil.ReadFile(FileName)
	if err != nil {
		return err
	}
	ber.PrintBytes(os.Stdout, file, "")
	packet := ber.DecodePacket(file)
	addLDAPDescriptions(packet)
	ber.PrintPacket(packet)

	return nil
}

type Error struct {
	sText      string
	ResultCode ResultCode
}

func (e *Error) Error() string {
	return fmt.Sprintf("LDAP Result Code %d %q: %s", e.ResultCode, e.ResultCode.String(), e.sText)
}

func newError(resultCode ResultCode, sText string) error {
	return &Error{ResultCode: resultCode, sText: sText}
}

func getResultCode(packet *ber.Packet) (ResultCode, string) {
	if packet == nil {
		return ErrorNetwork, "Empty packet"
	} else if len(packet.Children) >= 2 {
		response := packet.Children[1]
		if response == nil {
			return ErrorNetwork, "Empty response in packet"
		}

		if response.ClassType == ber.ClassApplication && response.TagType == ber.TypeConstructed && len(response.Children) >= 3 {
			// Children[1].Children[2] is the diagnosticMessage which is guaranteed to exist as seen here: https://tools.ietf.org/html/rfc4511#section-4.1.9
			resultCode := toolkit.ToInt(response.Children[0].Value, toolkit.RoundingAuto)
			message := toolkit.ToString(response.Children[2].Value)
			return ResultCode(resultCode), message
		}
	}

	return ErrorNetwork, "Invalid packet format"
}
