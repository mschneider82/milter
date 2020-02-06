package milter_test

import (
	"io/ioutil"
	"log"

	"github.com/mschneider82/milter"
)

// A Session embetted the SessionHandler Interface
type Session struct {
	milter.DefaultSession
}

// Body in this case we just want to show some chars of the mail
// All other Interactions are done by DefaultSession implementation
func (s *Session) Body(m *milter.Modifier) (milter.Response, error) {
	b, _ := ioutil.ReadAll(s.Message)
	log.Printf("Mail's first 100 chars: %s", string(b[0:100]))
	return milter.RespAccept, nil
}

func Example() {
	panichandler := func(e error) {
		log.Printf("Panic happend: %s\n", e.Error())
	}

	setsymlist := make(milter.RequestMacros)
	setsymlist[milter.SMFIM_CONNECT] = []milter.Macro{milter.MACRO_DAEMON_NAME, milter.Macro("{client_addr}")}

	milterfactory := func() (milter.SessionHandler, milter.OptAction, milter.OptProtocol, milter.RequestMacros) {
		return &Session{},
			milter.OptAllActions, // BitMask for wanted Actions
			0, // BitMask for unwanted SMTP Parts; 0 = nothing to opt out
			setsymlist // optional: can be nil
	}

	m := milter.New(milterfactory,
		milter.WithTCPListener("127.0.0.1:12349"),
		milter.WithLogger(milter.StdOutLogger),
		milter.WithPanicHandler(panichandler),
	)
	err := m.Run()
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	defer m.Close()
}
