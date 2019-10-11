package ldap

import (
	"github.com/eaciit/asn1-ber"
)

// PasswordModifyRequest implements the payload and encoding specified in
// https://tools.ietf.org/html/rfc3062
type PasswordModifyRequest struct {
	UserIdentity string
	OldPasswd    string
	NewPasswd    string
}

// Encode the PasswordModifyRequest into a ber.Packet
func (r *PasswordModifyRequest) Encode() (*ber.Packet, error) {
	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationExtendedRequest), nil, "PasswordModifyRequest")
	p.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, 0, "1.3.6.1.4.1.4203.1.11.1", "Password Modify Request"))

	octetString := ber.Encode(ber.ClassContext, ber.TypePrimitive, 1, nil, "Octet String")
	value := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PasswordModifyRequestValue")

	if r.UserIdentity != "" {
		userIdentity := ber.NewString(ber.ClassContext, ber.TypePrimitive, 0, string(r.UserIdentity), "userIdentity")
		value.AppendChild(userIdentity)
	}

	if r.OldPasswd != "" {
		oldPasswd := ber.NewString(ber.ClassContext, ber.TypePrimitive, 1, string(r.OldPasswd), "oldPasswd")
		value.AppendChild(oldPasswd)
	}

	if r.NewPasswd != "" {
		newPasswd := ber.NewString(ber.ClassContext, ber.TypePrimitive, 2, string(r.NewPasswd), "newPasswd")
		value.AppendChild(newPasswd)
	}

	octetString.AppendChild(value)
	p.AppendChild(octetString)

	return p, nil
}

func (l *Connection) Passwd(req *PasswordModifyRequest) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}

	encodedReq, err := req.Encode()
	if err != nil {
		return err
	}

	packet, err := requestBuildPacket(messageID, encodedReq, nil)

	return l.sendReqRespPacket(messageID, packet)
}
