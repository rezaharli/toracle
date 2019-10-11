package ldap

import (
	"github.com/eaciit/asn1-ber"
)

//ModifyDNRequest ::= [APPLICATION 12] SEQUENCE {
//entry           LDAPDN,
//newrdn          RelativeLDAPDN,
//deleteoldrdn    BOOLEAN,
//newSuperior     [0] LDAPDN OPTIONAL }
//
//ModifyDNResponse ::= [APPLICATION 13] LDAPResult

type ModDnRequest struct {
	DN            string
	NewRDN        string
	DeleteOldDn   bool
	NewSuperiorDN string
	Controls      []Control
}

//Untested.
func (l *Connection) ModDn(req *ModDnRequest) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}

	encodedModDn := encodeModDnRequest(req)

	packet, err := requestBuildPacket(messageID, encodedModDn, req.Controls)
	if err != nil {
		return err
	}

	return l.sendReqRespPacket(messageID, packet)
}

func encodeModDnRequest(req *ModDnRequest) (p *ber.Packet) {
	p = ber.Encode(ber.ClassApplication, ber.TypeConstructed,
		ber.Tag(ApplicationModifyDNRequest), nil, ApplicationModifyDNRequest.String())
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, req.DN, "LDAPDN"))
	p.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, req.NewRDN, "NewRDN"))
	p.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, req.DeleteOldDn, "deleteoldrdn"))
	if len(req.NewSuperiorDN) > 0 {
		p.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive,
			ber.TagEOC, req.NewSuperiorDN, "NewSuperiorDN"))
	}
	return
}
