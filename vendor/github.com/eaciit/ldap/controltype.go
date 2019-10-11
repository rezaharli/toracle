package ldap

import (
	"errors"
	"github.com/eaciit/asn1-ber"
)

type ControlType string

const (
	ControlTypeMatchedValuesRequest    ControlType = "1.2.826.0.1.3344810.2.3"
	ControlTypePermissiveModifyRequest ControlType = "1.2.840.113556.1.4.1413"
	ControlTypePaging                  ControlType = "1.2.840.113556.1.4.319"
	ControlTypeManageDsaITRequest      ControlType = "2.16.840.1.113730.3.4.2"
	ControlTypeSubtreeDeleteRequest    ControlType = "1.2.840.113556.1.4.805"
	ControlTypeNoOpRequest             ControlType = "1.3.6.1.4.1.4203.1.10.2"
	ControlTypeServerSideSortRequest   ControlType = "1.2.840.113556.1.4.473"
	ControlTypeServerSideSortResponse  ControlType = "1.2.840.113556.1.4.474"
	ControlTypeVlvRequest              ControlType = "2.16.840.1.113730.3.4.9"
	ControlTypeVlvResponse             ControlType = "2.16.840.1.113730.3.4.10"

//1.2.840.113556.1.4.473
//1.3.6.1.1.12
//1.3.6.1.1.13.1
//1.3.6.1.1.13.2
//1.3.6.1.4.1.26027.1.5.2
//1.3.6.1.4.1.42.2.27.8.5.1
//1.3.6.1.4.1.42.2.27.9.5.2
//1.3.6.1.4.1.42.2.27.9.5.8
//1.3.6.1.4.1.4203.1.10.1
//1.3.6.1.4.1.7628.5.101.1
//2.16.840.1.113730.3.4.12
//2.16.840.1.113730.3.4.16
//2.16.840.1.113730.3.4.17
//2.16.840.1.113730.3.4.18
//2.16.840.1.113730.3.4.19
//2.16.840.1.113730.3.4.3
//2.16.840.1.113730.3.4.4
//2.16.840.1.113730.3.4.5
//
)

var controlTypeStrings = map[ControlType]string{
	ControlTypeMatchedValuesRequest:    "MatchedValuesRequest",
	ControlTypePermissiveModifyRequest: "PermissiveModifyRequest",
	ControlTypePaging:                  "Paging",
	ControlTypeManageDsaITRequest:      "ManageDsaITRequest",
	ControlTypeSubtreeDeleteRequest:    "SubtreeDeleteRequest",
	ControlTypeNoOpRequest:             "NoOpRequest",
	ControlTypeServerSideSortRequest:   "ServerSideSortRequest",
	ControlTypeServerSideSortResponse:  "ServerSideSortResponse",
	ControlTypeVlvRequest:              "VlvRequest",
	ControlTypeVlvResponse:             "VlvResponse",
}

type controlTypeFn func(p *ber.Packet) (Control, error)

var controlTypeFns = map[ControlType]controlTypeFn{
	ControlTypeServerSideSortResponse: NewControlServerSideSortResponse,
	ControlTypePaging:                 NewControlPagingFromPacket,
	ControlTypeVlvResponse:            NewControlVlvResponse,
}

func (c ControlType) String() string {
	return controlTypeStrings[c]
}

func (c ControlType) function() (controlTypeFn, error) {
	f, ok := controlTypeFns[c]
	if !ok {
		return nil, errors.New("No function registered for " + c.String())
	}
	return f, nil
}
