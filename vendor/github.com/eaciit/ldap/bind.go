package ldap

import (
	"github.com/eaciit/asn1-ber"
)

/*
Simple bind to the server. If using a timeout you should close the connection
on a bind failure.
*/
func (l *Connection) Bind(username, password string) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}

	encodedBind := encodeSimpleBindRequest(username, password)

	packet, err := requestBuildPacket(messageID, encodedBind, nil)
	if err != nil {
		return err
	}

	return l.sendReqRespPacket(messageID, packet)

}

func encodeSimpleBindRequest(username, password string) (bindRequest *ber.Packet) {
	bindRequest = ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationBindRequest), nil, "Bind Request")
	bindRequest.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 3, "Version"))
	bindRequest.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, username, "User Name"))
	bindRequest.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, 0, password, "Password"))
	return
}
