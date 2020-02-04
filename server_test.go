package milter

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
	"testing"

	"io"

	"github.com/mschneider82/milterclient"
)

/* TestMilter object */
type TestMilter struct {
	multipart bool
	message   *bytes.Buffer
}

// https://github.com/mschneider82/milter/blob/master/interface.go
func (e *TestMilter) Init(sid, mid string) {
	return
}

func (e *TestMilter) Disconnect() {
	return
}

func (e *TestMilter) Connect(name, value string, port uint16, ip net.IP, m *Modifier) (Response, error) {
	return RespContinue, nil
}

func (e *TestMilter) Helo(h string, m *Modifier) (Response, error) { return RespContinue, nil }

func (e *TestMilter) MailFrom(name string, m *Modifier) (Response, error) {
	return RespContinue, nil
}

func (e *TestMilter) RcptTo(name string, m *Modifier) (Response, error) {
	return RespContinue, nil
}

/* handle headers one by one */
func (e *TestMilter) Header(name, value string, m *Modifier) (Response, error) {
	// if message has multiple parts set processing flag to true
	if name == "Content-Type" && strings.HasPrefix(value, "multipart/") {
		e.multipart = true
	}
	return RespContinue, nil
}

/* at end of headers initialize message buffer and add headers to it */
func (e *TestMilter) Headers(headers textproto.MIMEHeader, m *Modifier) (Response, error) {
	// return accept if not a multipart message
	if !e.multipart {
		return RespAccept, nil
	}
	// prepare message buffer
	e.message = new(bytes.Buffer)
	// print headers to message buffer
	for k, vl := range headers {
		for _, v := range vl {
			if _, err := fmt.Fprintf(e.message, "%s: %s\n", k, v); err != nil {
				return nil, err
			}
		}
	}
	if _, err := fmt.Fprintf(e.message, "\n"); err != nil {
		return nil, err
	}
	// continue with milter processing
	return RespContinue, nil
}

// accept body chunk
func (e *TestMilter) BodyChunk(chunk []byte, m *Modifier) (Response, error) {
	// save chunk to buffer
	if _, err := e.message.Write(chunk); err != nil {
		return nil, err
	}
	return RespContinue, nil
}

/* Body is called when email message body has been sent */
func (e *TestMilter) Body(m *Modifier) (Response, error) {
	// prepare buffer
	_ = bytes.NewReader(e.message.Bytes())
	fmt.Println("size of body: ", e.message.Len())
	m.AddHeader("name", "value")
	m.AddRecipient("some.new.rcpt@example.com")
	m.ChangeFrom("new.from@example.com")
	m.ChangeHeader(0, "Subject", "New Subject")
	m.DeleteRecipient("to@example.com")
	m.InsertHeader(0, "new", "value")
	m.ReplaceBody([]byte("new body"))

	// accept message by default
	return RespAccept, nil
}

type logger struct{}

func (f logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/* myRunServer creates new Milter instance */
func myRunServer(socket net.Listener) {
	// declare milter init function
	init := func() (Milter, OptAction, OptProtocol) {
		return &TestMilter{},
			OptAddHeader | OptChangeHeader | OptChangeFrom | OptAddRcpt | OptRemoveRcpt | OptChangeBody,
			OptNoRcptTo
	}

	errhandler := func(e error) {
		fmt.Printf("Panic happend: %s\n", e.Error())
	}

	symlist := SymListFactory{}
	symlist.Set(SMFIM_CONNECT, "{daemon_name} {client_addr}")

	server := Server{
		Listener:       socket,
		MilterFactory:  init,
		ErrHandlers:    []func(error){errhandler},
		Logger:         &logger{},
		SymListFactory: &symlist,
	}
	defer server.Close()
	// start server
	err := server.RunServer()
	if err != nil {
		fmt.Printf("Error happened: %s\n", err.Error())
	}

}

/* main program */
func TestMilterClient(t *testing.T) {

	// parse commandline arguments
	protocol := "tcp"
	address := "127.0.0.1:12349"

	// bind to listening address
	socket, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatal(err)
	}

	// run server
	go myRunServer(socket)

	// run tests:
	emlFilePath := "testmail.eml"
	eml, err := os.Open(emlFilePath)
	if err != nil {
		t.Errorf("Error opening test eml file %v: %v", emlFilePath, err)
	}
	defer eml.Close()

	buf := new(bytes.Buffer)
	for i := 0; i < 1000000; i++ {
		buf.WriteString("fsdokfpsdkofksdopfkpsodfkpsdkfopsdkfposdkfposdkfposdkfopsdkfopsdkfposdfkposdffsdfsdfsdstring")
	}
	b := buf.Bytes()

	bigmailreader := io.MultiReader(eml, bytes.NewReader(b))

	msgID := milterclient.GenMtaID(12)
	last, err := milterclient.SendEml(bigmailreader, "127.0.0.1:12349", "from@unittest.de", "to@unittest.de", "", "", msgID, false, 5)
	if err != nil {
		t.Errorf("Error sending eml to milter: %v", err)
	}

	fmt.Printf("MsgId: %s, Lastmilter code: %s\n", msgID, string(last))
	if last != SMFIR_CHGFROM {
		t.Errorf("Excepted Accept from Milter, got %v", last)
	}

}
