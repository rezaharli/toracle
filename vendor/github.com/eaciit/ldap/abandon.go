package ldap

import (
	"fmt"
	"github.com/eaciit/asn1-ber"
)

// Will return an error. Normally due to closed connection.
func (l *Connection) Abandon(abandonMessageID int64) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}

	encodedAbandon := ber.NewInteger(ber.ClassApplication, ber.TypePrimitive, ber.Tag(ApplicationAbandonRequest), abandonMessageID, ApplicationAbandonRequest.String())

	packet, err := requestBuildPacket(messageID, encodedAbandon, nil)
	if err != nil {
		return err
	}

	if l.Debug {
		ber.PrintPacket(packet)
	}

	channel, err := l.sendMessage(packet)

	if err != nil {
		return err
	}

	if channel == nil {
		return newError(ErrorNetwork, "Could not send message")
	}

	defer l.finishMessage(messageID)
	if l.Debug {
		fmt.Printf("%d: NOT waiting Abandon for response\n", messageID)
	}

	// success
	return nil
}
