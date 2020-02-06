package milter_test

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

	"github.com/mschneider82/milter"
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

func (e *TestMilter) Connect(name, value string, port uint16, ip net.IP, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (e *TestMilter) Helo(h string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (e *TestMilter) MailFrom(name string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

func (e *TestMilter) RcptTo(name string, m *milter.Modifier) (milter.Response, error) {
	return milter.RespContinue, nil
}

/* handle headers one by one */
func (e *TestMilter) Header(name, value string, m *milter.Modifier) (milter.Response, error) {
	// if message has multiple parts set processing flag to true
	if name == "Content-Type" && strings.HasPrefix(value, "multipart/") {
		e.multipart = true
	}
	return milter.RespContinue, nil
}

/* at end of headers initialize message buffer and add headers to it */
func (e *TestMilter) Headers(headers textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	// return accept if not a multipart message
	if !e.multipart {
		return milter.RespAccept, nil
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
	return milter.RespContinue, nil
}

// accept body chunk
func (e *TestMilter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {
	// save chunk to buffer
	if _, err := e.message.Write(chunk); err != nil {
		return nil, err
	}
	return milter.RespContinue, nil
}

/* Body is called when email message body has been sent */
func (e *TestMilter) Body(m *milter.Modifier) (milter.Response, error) {
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
	return milter.RespAccept, nil
}

type logger struct{}

func (f logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

/* main program */
func TestMilterClient(t *testing.T) {
	panichandler := func(e error) {
		fmt.Printf("Panic happend: %s\n", e.Error())
	}

	setsymlist := make(milter.RequestMacros)
	setsymlist[milter.SMFIM_CONNECT] = []milter.Macro{milter.MACRO_DAEMON_NAME, milter.Macro("{client_addr}")}

	milterfactory := func() (milter.SessionHandler, milter.OptAction, milter.OptProtocol, milter.RequestMacros) {
		return &TestMilter{},
			milter.OptAddHeader | milter.OptChangeHeader | milter.OptChangeFrom | milter.OptAddRcpt | milter.OptRemoveRcpt | milter.OptChangeBody,
			0,
			setsymlist
	}
	m := milter.New(milterfactory,
		milter.WithTCPListener("127.0.0.1:12349"),
		milter.WithLogger(milter.StdOutLogger),
		milter.WithPanicHandler(panichandler),
	)
	go func() {
		err := m.Run()
		if err != nil {
			log.Fatalf("Error: %s", err.Error())
		}
	}()
	defer m.Close()

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
	if last != milter.SMFIR_CHGFROM {
		t.Errorf("Excepted Accept from Milter, got %v", last)
	}
}
