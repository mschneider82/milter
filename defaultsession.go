package milter

import (
	"bytes"
	"fmt"
	"net"
	"net/textproto"
)

// A DefaultSession can be used as a basic implementation for SessionHandler Interface
// It has already a From and Rcpts[] field, mostly used to embett in your own
// Session struct.
type DefaultSession struct {
	SID            string // Session ID
	MID            string // Mail ID
	From           string
	ClientName     string
	HeloName       string
	ClientIP       net.IP
	Rcpts          []string
	MessageHeaders textproto.MIMEHeader
	Message        *bytes.Buffer
}

// https://github.com/mschneider82/milter/blob/master/interface.go
func (e *DefaultSession) Init(sid, mid string) {
	e.SID, e.MID = sid, mid
	e.Message = new(bytes.Buffer)
	return
}

func (e *DefaultSession) Disconnect() {
	return
}

func (e *DefaultSession) Connect(name, family string, port uint16, ip net.IP, m *Modifier) (Response, error) {
	e.ClientName = name
	e.ClientIP = ip
	return RespContinue, nil
}

func (e *DefaultSession) Helo(name string, m *Modifier) (Response, error) {
	e.HeloName = name
	return RespContinue, nil
}

func (e *DefaultSession) MailFrom(from string, m *Modifier) (Response, error) {
	e.From = from
	return RespContinue, nil
}

func (e *DefaultSession) RcptTo(rcptTo string, m *Modifier) (Response, error) {
	e.Rcpts = append(e.Rcpts, rcptTo)
	return RespContinue, nil
}

/* handle headers one by one */
func (e *DefaultSession) Header(name, value string, m *Modifier) (Response, error) {
	headerLine := fmt.Sprintf("%s: %s\r\n", name, value)
	if _, err := e.Message.WriteString(headerLine); err != nil {
		return nil, err
	}
	return RespContinue, nil
}

// emptyLine is needed between Headers and Body
const emptyLine = "\r\n"

/* at end of headers initialize message buffer and add headers to it */
func (e *DefaultSession) Headers(headers textproto.MIMEHeader, m *Modifier) (Response, error) {
	// return accept if not a multipart message
	e.MessageHeaders = headers

	if _, err := e.Message.WriteString(emptyLine); err != nil {
		return nil, err
	}
	// continue with milter processing
	return RespContinue, nil
}

// accept body chunk
func (e *DefaultSession) BodyChunk(chunk []byte, m *Modifier) (Response, error) {
	// save chunk to buffer
	if _, err := e.Message.Write(chunk); err != nil {
		return nil, err
	}
	return RespContinue, nil
}

/* Body is called when email message body has been sent */
func (e *DefaultSession) Body(m *Modifier) (Response, error) {
	// accept message by default
	return RespAccept, nil
}
