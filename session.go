package milter

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"net"
	"net/textproto"
	"strings"
	"time"
)

// OptAction sets which actions the milter wants to perform.
// Multiple options can be set using a bitmask.
type OptAction uint32

// OptProtocol masks out unwanted parts of the SMTP transaction.
// Multiple options can be set using a bitmask.
type OptProtocol uint32

const (
	// set which actions the milter wants to perform
	OptNone           OptAction = 0x00  /* SMFIF_NONE no flags */
	OptAddHeader      OptAction = 0x01  /* SMFIF_ADDHDRS filter may add headers */
	OptChangeBody     OptAction = 0x02  /* SMFIF_CHGBODY filter may replace body */
	OptAddRcpt        OptAction = 0x04  /* SMFIF_ADDRCPT filter may add recipients */
	OptRemoveRcpt     OptAction = 0x08  /* SMFIF_DELRCPT filter may delete recipients */
	OptChangeHeader   OptAction = 0x10  /* SMFIF_CHGHDRS filter may change/delete headers */
	OptQuarantine     OptAction = 0x20  /* SMFIF_QUARANTINE filter may quarantine envelope */
	OptChangeFrom     OptAction = 0x40  /* SMFIF_CHGFROM filter may change "from" (envelope sender) */
	OptAddRcptPartial OptAction = 0x80  /* SMFIF_ADDRCPT_PAR filter may add recipients, including ESMTP parameter to the envelope.*/
	OptSetSymList     OptAction = 0x100 /* SMFIF_SETSYMLIST filter can send set of symbols (macros) that it wants */
	OptAllActions     OptAction = OptAddHeader | OptChangeBody | OptAddRcpt | OptRemoveRcpt | OptChangeHeader | OptQuarantine | OptChangeFrom | OptAddRcptPartial | OptSetSymList

	// mask out unwanted parts of the SMTP transaction
	OptNoConnect    OptProtocol = 0x01       /* SMFIP_NOCONNECT MTA should not send connect info */
	OptNoHelo       OptProtocol = 0x02       /* SMFIP_NOHELO MTA should not send HELO info */
	OptNoMailFrom   OptProtocol = 0x04       /* SMFIP_NOMAIL MTA should not send MAIL info */
	OptNoRcptTo     OptProtocol = 0x08       /* SMFIP_NORCPT MTA should not send RCPT info */
	OptNoBody       OptProtocol = 0x10       /* SMFIP_NOBODY MTA should not send body (chunk) */
	OptNoHeaders    OptProtocol = 0x20       /* SMFIP_NOHDRS MTA should not send headers */
	OptNoEOH        OptProtocol = 0x40       /* SMFIP_NOEOH MTA should not send EOH */
	OptNrHdr        OptProtocol = 0x80       /* SMFIP_NR_HDR SMFIP_NOHREPL No reply for headers */
	OptNoUnknown    OptProtocol = 0x100      /* SMFIP_NOUNKNOWN MTA should not send unknown commands */
	OptNoData       OptProtocol = 0x200      /* SMFIP_NODATA MTA should not send DATA */
	OptSkip         OptProtocol = 0x400      /* SMFIP_SKIP MTA understands SMFIS_SKIP */
	OptRcptRej      OptProtocol = 0x800      /* SMFIP_RCPT_REJ MTA should also send rejected RCPTs */
	OptNrConn       OptProtocol = 0x1000     /* SMFIP_NR_CONN No reply for connect */
	OptNrHelo       OptProtocol = 0x2000     /* SMFIP_NR_HELO No reply for HELO */
	OptNrMailFrom   OptProtocol = 0x4000     /* SMFIP_NR_MAIL No reply for MAIL */
	OptNrRcptTo     OptProtocol = 0x8000     /* SMFIP_NR_RCPT No reply for RCPT */
	OptNrData       OptProtocol = 0x10000    /* SMFIP_NR_DATA No reply for DATA */
	OptNrUnknown    OptProtocol = 0x20000    /* SMFIP_NR_UNKN No reply for UNKNOWN */
	OptNrEOH        OptProtocol = 0x40000    /* SMFIP_NR_EOH No reply for eoh */
	OptNrBody       OptProtocol = 0x80000    /* SMFIP_NR_BODY No reply for body chunk */
	OptHdrLeadSpace OptProtocol = 0x100000   /* SMFIP_HDR_LEADSPC header value leading space */
	OptMDS256K      OptProtocol = 0x10000000 /* SMFIP_MDS_256K MILTER_MAX_DATA_SIZE=256K */
	OptMDS1M        OptProtocol = 0x20000000 /* SMFIP_MDS_1M MILTER_MAX_DATA_SIZE=1M */
)

// milterSession keeps session state during MTA communication
type milterSession struct {
	actions   OptAction
	protocol  OptProtocol
	sock      io.ReadWriteCloser
	headers   textproto.MIMEHeader
	macros    map[string]string
	milter    Milter
	sessionID string
	mailID    string
	logger    Logger
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// genRandomID generates an random ID. vocals are removed to prevent dirty words which could be negative in spam score
func (c *milterSession) genRandomID(length int) string {
	var letters = []rune("bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ReadPacket reads incoming milter packet
func (c *milterSession) ReadPacket() (*Message, error) {
	// read packet length
	var length uint32
	if err := binary.Read(c.sock, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// read packet data
	data := make([]byte, length)
	if _, err := io.ReadFull(c.sock, data); err != nil {
		return nil, err
	}

	// prepare response data
	message := Message{
		Code: data[0],
		Data: data[1:],
	}

	return &message, nil
}

// WritePacket sends a milter response packet to socket stream
func (m *milterSession) WritePacket(msg *Message) error {
	buffer := bufio.NewWriter(m.sock)

	// calculate and write response length
	length := uint32(len(msg.Data) + 1)
	if err := binary.Write(buffer, binary.BigEndian, length); err != nil {
		return err
	}

	// write response code
	if err := buffer.WriteByte(msg.Code); err != nil {
		return err
	}

	// write response data
	if _, err := buffer.Write(msg.Data); err != nil {
		return err
	}

	// flush data to network socket stream
	if err := buffer.Flush(); err != nil {
		return err
	}

	return nil
}

// Process processes incoming milter commands
func (m *milterSession) Process(msg *Message) (Response, error) {
	switch msg.Code {
	case SMFIC_ABORT:
		// abort current message and start over
		m.headers = nil
		m.macros = nil
		// do not send response

		// on SMFIC_ABORT
		// Reset state to before SMFIC_MAIL and continue,
		// unless connection is dropped by MTA
		m.milter.Init(m.sessionID, m.mailID)

		return nil, nil

	case SMFIC_BODY:
		// body chunk
		return m.milter.BodyChunk(msg.Data, newModifier(m))

	case SMFIC_CONNECT:
		// new connection, get hostname
		Hostname := readCString(msg.Data)
		msg.Data = msg.Data[len(Hostname)+1:]
		// get protocol family
		protocolFamily := msg.Data[0]
		msg.Data = msg.Data[1:]
		// get port
		var Port uint16
		if protocolFamily == '4' || protocolFamily == '6' {
			if len(msg.Data) < 2 {
				return RespTempFail, nil
			}
			Port = binary.BigEndian.Uint16(msg.Data)
			msg.Data = msg.Data[2:]
			if protocolFamily == '6' {
				msg.Data = msg.Data[5:] // ipv6 is in format IPv6:XXX:XX1
			}
		}
		// get address
		Address := readCString(msg.Data)
		// convert address and port to human readable string
		family := map[byte]string{
			'U': "unknown",
			'L': "unix",
			'4': "tcp4",
			'6': "tcp6",
		}
		// run handler and return
		return m.milter.Connect(
			Hostname,
			family[protocolFamily],
			Port,
			net.ParseIP(Address),
			newModifier(m))

	case SMFIC_MACRO:
		// define macros
		m.macros = make(map[string]string)
		// convert data to Go strings
		data := decodeCStrings(msg.Data[1:])
		if len(data) != 0 {
			// store data in a map
			for i := 0; i < len(data); i += 2 {
				m.macros[data[i]] = data[i+1]
			}
		}
		// do not send response
		return nil, nil

	case SMFIC_BODYEOB:
		// End of body marker
		return m.milter.Body(newModifier(m))

	case SMFIC_HELO:
		// helo command HELO/EHLO name
		name := strings.TrimSuffix(string(msg.Data), null)
		return m.milter.Helo(name, newModifier(m))

	case SMFIC_HEADER:
		// make sure headers is initialized - Mail header
		if m.headers == nil {
			m.headers = make(textproto.MIMEHeader)
		}
		// add new header to headers map
		HeaderData := decodeCStrings(msg.Data)
		if len(HeaderData) == 2 {
			m.headers.Add(HeaderData[0], HeaderData[1])
			// call and return milter handler
			return m.milter.Header(HeaderData[0], HeaderData[1], newModifier(m))
		}

	case SMFIC_MAIL:
		// MAIL FROM: information
		m.mailID = m.genRandomID(12)
		// Call Init for a new Mail
		m.milter.Init(m.sessionID, m.mailID)
		// envelope from address
		envfrom := readCString(msg.Data)
		return m.milter.MailFrom(strings.ToLower(strings.Trim(envfrom, "<>")), newModifier(m))

	case SMFIC_EOH:
		// end of headers
		return m.milter.Headers(m.headers, newModifier(m))

	case SMFIC_OPTNEG:
		// Option negotiation - ignore request and prepare response buffer
		buffer := new(bytes.Buffer)
		// prepare response data
		for _, value := range []uint32{2, uint32(m.actions), uint32(m.protocol)} {
			if err := binary.Write(buffer, binary.BigEndian, value); err != nil {
				return nil, err
			}
		}
		// build and send packet
		return NewResponse(SMFIC_OPTNEG, buffer.Bytes()), nil

	case SMFIC_QUIT:
		// Quit milter communication
		// client requested session close
		return nil, ErrCloseSession

	case SMFIC_RCPT:
		// RCPT TO: information
		// envelope to address
		envto := readCString(msg.Data)
		return m.milter.RcptTo(strings.ToLower(strings.Trim(envto, "<>")), newModifier(m))

	case SMFIC_DATA:
		// data, ignore

	default:
		// print error and close session
		log.Printf("Unrecognized command code: %c", msg.Code)
		return nil, ErrCloseSession
	}

	// by default continue with next milter message
	return RespContinue, nil
}

// HandleMilterComands processes all milter commands in the same connection
func (m *milterSession) HandleMilterCommands() {

	defer m.sock.Close()
	defer m.milter.Disconnect()

	m.sessionID = m.genRandomID(12)

	// Call Init() for a new Session first
	m.milter.Init(m.sessionID, m.mailID)

	for {
		// ReadPacket
		msg, err := m.ReadPacket()
		if err != nil {
			if err != io.EOF {
				m.logger.Printf("Error reading milter command: %v", err)
			}
			return
		}

		// process command
		resp, err := m.Process(msg)
		if err != nil {
			if err != ErrCloseSession {
				// log error condition
				m.logger.Printf("Error performing milter command: %v", err)
			}
			return
		}

		// ignore empty responses
		if resp != nil {
			// send back response message
			if err = m.WritePacket(resp.Response()); err != nil {
				m.logger.Printf("Error writing packet: %v", err)
				return
			}
		}
	}
}
