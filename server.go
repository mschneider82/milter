// A Go library for milter support
package milter

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var defaultServer Server

// MilterFactory initializes milter options
// multiple options can be set using a bitmask
type MilterFactory func() (SessionHandler, OptAction, OptProtocol, RequestMacros)

// Close server listener and wait worked process
func Close() {
	defaultServer.Close()
}

// Server Milter for handling and processing incoming connections
// support panic handling via ErrHandler
// couple of func(error) could be provided for handling error
type Server struct {
	listener      net.Listener
	milterFactory MilterFactory
	errHandlers   []func(error)
	logger        CustomLogger
	wg            sync.WaitGroup
	quit          chan struct{}
	exited        chan struct{}
}

// New generates a new Server
func New(milterfactory MilterFactory, lopt ListenerOption, opts ...Option) *Server {
	server := &Server{
		milterFactory: milterfactory,
		logger:        stdoutLogger{},
		wg:            sync.WaitGroup{},
	}
	lopt.lapply(server)
	for _, opt := range opts {
		opt.apply(server)
	}
	return server
}

// RequestMacros - Also known as SetSymList: the list of macros that the milter wants to receive from the MTA for a protocol Stage (stages has the prefix SMFIM_).
// if nil, then there are not Macros requested and the default macros from MTA are used.
type RequestMacros map[Stage][]Macro

// Close for graceful shutdown
// Stop accepting new connections
// And wait until processing connections ends
func (s *Server) Close() error {
	close(s.quit)
	s.listener.Close()
	<-s.exited
	return nil
}

// Run starts milter server via provided listener
func (s *Server) Run() error {
	if s.listener == nil {
		return ErrNoListenAddr
	}
	s.quit = make(chan struct{})
	s.exited = make(chan struct{})
	for {
		select {
		case <-s.quit:
			s.wg.Wait()
			close(s.exited)
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				s.logger.Printf("Error: Failed to accept connection: %s", err.Error())
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if conn == nil {
				s.logger.Printf("Error: conn is nil")
				continue
			}
			s.wg.Add(1)
			go func() {
				defer handlePanic(s.errHandlers)
				defer s.wg.Done()
				s.handleCon(conn)
			}()
		}
	}
}

// Handle incoming connections
func (s *Server) handleCon(conn net.Conn) {
	// create milter object
	milter, actions, protocol, requestmacros := s.milterFactory()

	session := milterSession{
		actions:  actions,
		protocol: protocol,
		sock:     conn,
		milter:   milter,
		logger:   s.logger,
		symlists: requestmacros,
	}
	// handle connection commands
	session.HandleMilterCommands()
}

// Recover panic from session and call handle with occurred error
// If no any handle provided panics will not recovered
func handlePanic(handlers []func(error)) {
	var err error

	if handlers == nil {
		return
	}

	r := recover()
	switch r.(type) {
	case nil:
		return
	case error:
		err = r.(error)
	default:
		err = errors.New(fmt.Sprint(r))
	}
	for _, f := range handlers {
		f(err)
	}
}
