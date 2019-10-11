package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/eaciit/asn1-ber"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Connection struct {
	IsTLS bool
	IsSSL bool
	Debug bool

	Addr                        string
	NetworkConnectTimeout       time.Duration
	ReadTimeout                 time.Duration
	AbandonMessageOnReadTimeout bool

	TlsConfig *tls.Config

	conn               net.Conn
	chanResults        map[int64]chan *ber.Packet
	lockChanResults    sync.RWMutex
	chanProcessMessage chan *messagePacket
	closeLock          sync.RWMutex
	chanMessageID      chan int64
	connected          bool
}

// NewConnection creates a new Connection object. The address is in the same format as
// used in the net package.
func NewConnection(address string) *Connection {
	return &Connection{Addr: address}
}

// Behaves like NewConnection, except that an additional parameter tlsConfig is expected.
// The resulting connection uses TLS.
func NewTLSConnection(address string, tlsConfig *tls.Config) *Connection {
	return &Connection{
		Addr:      address,
		IsTLS:     true,
		TlsConfig: tlsConfig,
	}
}

// Behaves like NewConnection, except that an additional parameter tlsConfig is expected.
// The resulting connection uses SSL.
func NewSSLConnection(address string, tlsConfig *tls.Config) *Connection {
	return &Connection{
		Addr:      address,
		IsSSL:     true,
		TlsConfig: tlsConfig,
	}
}

// Connect connects using information in Connection.
// Connection should be populated with connection information.
func (l *Connection) Connect() error {
	l.chanResults = map[int64]chan *ber.Packet{}
	l.chanProcessMessage = make(chan *messagePacket)
	l.chanMessageID = make(chan int64)

	if l.conn == nil {
		var c net.Conn
		var err error
		if l.NetworkConnectTimeout > 0 {
			c, err = net.DialTimeout("tcp", l.Addr, l.NetworkConnectTimeout)
		} else {
			c, err = net.Dial("tcp", l.Addr)
		}

		if err != nil {
			return err
		}

		if l.IsSSL {
			tlsConn := tls.Client(c, l.TlsConfig)
			err = tlsConn.Handshake()
			if err != nil {
				return err
			}
			l.conn = tlsConn
		} else {
			l.conn = c
		}
	}
	l.start()
	l.connected = true
	if l.IsTLS {
		err := l.startTLS()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Connection) start() {
	go l.reader()
	go l.processMessages()
}

// Close closes the connection.
func (l *Connection) Close() error {
	if l.Debug {
		log.Println("Starting Close()")
	}
	l.sendProcessMessage(&messagePacket{Op: MessageQuit})
	return nil
}

// Returns the next available messageID
func (l *Connection) nextMessageID() (messageID int64, ok bool) {
	messageID, ok = <-l.chanMessageID
	if l.Debug {
		log.Printf("MessageID: %d, ok: %v\n", messageID, ok)
	}
	return
}

// StartTLS sends the command to start a TLS session and then creates a new TLS Client
func (l *Connection) startTLS() error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "MessageID channel is closed.")
	}

	if l.IsSSL {
		return newError(ErrorNetwork, "Already encrypted")
	}

	tlsRequest := encodeTLSRequest()
	packet, err := requestBuildPacket(messageID, tlsRequest, nil)

	if err != nil {
		return err
	}

	err = l.sendReqRespPacket(messageID, packet)
	if err != nil {
		return err
	}

	conn := tls.Client(l.conn, l.TlsConfig)
	err = conn.Handshake()
	if err != nil {
		return err
	}
	l.IsSSL = true
	l.conn = conn

	return nil
}

func encodeTLSRequest() (tlsRequest *ber.Packet) {
	tlsRequest = ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationExtendedRequest), nil, "Start TLS")
	tlsRequest.AppendChild(ber.NewString(ber.ClassContext, ber.TypePrimitive, 0, "1.3.6.1.4.1.1466.20037", "TLS Extended Command"))
	return
}

const (
	MessageQuit     = 0
	MessageRequest  = 1
	MessageResponse = 2
	MessageFinish   = 3
)

type messagePacket struct {
	Op        int
	MessageID int64
	Packet    *ber.Packet
	Channel   chan *ber.Packet
}

func (l *Connection) getNewResultChannel(message_id int64) (out chan *ber.Packet, err error) {
	// as soon as a channel is requested add to chanResults to never miss
	// on cleanup.
	l.lockChanResults.Lock()
	defer l.lockChanResults.Unlock()

	if l.chanResults == nil {
		return nil, newError(ErrorClosing, "Connection closing/closed")
	}

	if _, ok := l.chanResults[message_id]; ok {
		errStr := fmt.Sprintf("chanResults already allocated, message_id: %d", message_id)
		return nil, newError(ErrorUnknown, errStr)
	}

	out = make(chan *ber.Packet, ResultChanBufferSize)
	l.chanResults[message_id] = out
	return
}

func (l *Connection) sendMessage(p *ber.Packet) (out chan *ber.Packet, err error) {
	message_id, ok := p.Children[0].Value.(int64)
	if !ok {
		return nil, errors.New(fmt.Sprintf("type assertion int64 for %v failed!", p.Children[0].Value))
	}
	// sendProcessMessage may not process a message on shutdown
	// getNewResultChannel adds id/chan to chan results
	out, err = l.getNewResultChannel(message_id)
	if err != nil {
		return
	}
	if l.Debug {
		log.Printf("sendMessage-> message_id: %d, out: %v\n", message_id, out)
	}

	message_packet := &messagePacket{Op: MessageRequest, MessageID: message_id, Packet: p, Channel: out}
	l.sendProcessMessage(message_packet)
	return
}

func (l *Connection) processMessages() {
	defer l.closeAllChannels()
	defer func() {
		// Close all channels, connection and quit.
		// Use closeLock to stop MessageRequests
		// and l.connected to stop any future MessageRequests.
		l.closeLock.Lock()
		defer l.closeLock.Unlock()
		l.connected = false
		// will shutdown reader.
		l.conn.Close()
	}()
	var message_id int64 = 1
	var message_packet *messagePacket

	for {
		select {
		case l.chanMessageID <- message_id:
			message_id++
		case message_packet = <-l.chanProcessMessage:
			switch message_packet.Op {
			case MessageQuit:
				if l.Debug {
					log.Printf("Shutting down\n")
				}
				return
			case MessageRequest:
				// Add to message list and write to network
				if l.Debug {
					fmt.Printf("Sending message %d\n", message_packet.MessageID)
				}
				buf := message_packet.Packet.Bytes()
				for len(buf) > 0 {
					n, err := l.conn.Write(buf)
					if err != nil {
						if l.Debug {
							fmt.Printf("Error Sending Message: %s\n", err)
						}
						return
					}
					if n == len(buf) {
						break
					}
					buf = buf[n:]
				}
			case MessageFinish:
				// Remove from message list
				if l.Debug {
					fmt.Printf("Finished message %d\n", message_packet.MessageID)
				}
				l.lockChanResults.Lock()
				delete(l.chanResults, message_packet.MessageID)
				l.lockChanResults.Unlock()
			}
		}
	}
}

func (l *Connection) closeAllChannels() {
	l.lockChanResults.Lock()
	defer l.lockChanResults.Unlock()
	for MessageID, Channel := range l.chanResults {
		if l.Debug {
			fmt.Printf("Closing channel for MessageID %d\n", MessageID)
		}
		close(Channel)
		delete(l.chanResults, MessageID)
	}
	l.chanResults = nil

	close(l.chanMessageID)
	l.chanMessageID = nil

	close(l.chanProcessMessage)
	l.chanProcessMessage = nil
}

func (l *Connection) finishMessage(MessageID int64) {
	message_packet := &messagePacket{Op: MessageFinish, MessageID: MessageID}
	l.sendProcessMessage(message_packet)
}

func (l *Connection) reader() {
	defer l.Close()
	for {
		p, err := ber.ReadPacket(l.conn)
		if err != nil {
			if l.Debug {
				fmt.Printf("ldap.reader: %s\n", err)
			}
			return
		}

		addLDAPDescriptions(p)

		message_id, ok := p.Children[0].Value.(int64)
		if !ok {
			// type assertion failed.. maybe we better stop
			return
		}

		message_packet := &messagePacket{Op: MessageResponse, MessageID: message_id, Packet: p}

		l.readerToChanResults(message_packet)
	}
}

func (l *Connection) readerToChanResults(message_packet *messagePacket) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Recovered in readerToChanResults", r)
		}
	}()
	if l.Debug {
		fmt.Printf("Receiving message %d\n", message_packet.MessageID)
	}

	// very small chance on disconnect to write to a closed channel as
	// lockChanResults is unlocked immediately hence defer above.
	// Don't lock while sending to chanResult below as that can block and hold
	// the lock.
	l.lockChanResults.RLock()
	chanResult, ok := l.chanResults[message_packet.MessageID]
	l.lockChanResults.RUnlock()

	if !ok {
		if l.Debug {
			fmt.Printf("Message Result chan not found (possible Abandon), MessageID: %d\n", message_packet.MessageID)
		}
	} else {
		// chanResult is a buffered channel of ResultChanBufferSize
		chanResult <- message_packet.Packet
	}
}

func (l *Connection) sendProcessMessage(message *messagePacket) {
	go func() {
		// multiple senders can queue on l.chanProcessMessage
		// but block on shutdown.
		l.closeLock.RLock()
		defer l.closeLock.RUnlock()
		if l.connected {
			l.chanProcessMessage <- message
		}
	}()
}
