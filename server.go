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

// MilterInit initializes milter options
// multiple options can be set using a bitmask
type MilterInit func() (Milter, OptAction, OptProtocol)

// RunServer provides a convenient way to start a milter server
// Handlers provide way to handle errors from panics
// With nil handlers panics not recovered
func RunServer(server net.Listener, logger Logger, init MilterInit, handlers ...func(error)) error {
	defaultServer.Listener = server
	defaultServer.MilterFactory = init
	defaultServer.ErrHandlers = handlers
	defaultServer.Logger = logger
	return defaultServer.RunServer()
}

// Close server listener and wait worked process
func Close() {
	defaultServer.Close()
}

// Server Milter for handling and processing incoming connections
// support panic handling via ErrHandler
// couple of func(error) could be provided for handling error
type Server struct {
	Listener       net.Listener
	MilterFactory  MilterInit
	ErrHandlers    []func(error)
	Logger         Logger
	SymListFactory *SymListFactory // SymListFactory Optional: Set the list of macros that the milter wants to receive from the MTA for a protocol stage.
	sync.WaitGroup
	quit   chan struct{}
	exited chan struct{}
}

// SymListFactory Factory to Set the list of macros that the milter wants to receive from the MTA for a protocol Stage (stages has the prefix SMFIM_).
type SymListFactory struct {
	m map[Stage]string
}

// Set the list of macros that the milter wants to receive from the MTA for a protocol Stage (stages has the prefix SMFIM_).
// list of macros (separated by space). Example: "{rcpt_mailer} {rcpt_host}"
func (s *SymListFactory) Set(stage Stage, macros string) {
	if s.m == nil {
		s.m = make(map[Stage]string)
	}
	s.m[stage] = macros
}

// Close for graceful shutdown
// Stop accepting new connections
// And wait until processing connections ends
func (s *Server) Close() error {
	close(s.quit)
	s.Listener.Close()
	<-s.exited
	return nil
}

// RunServer starts milter server via provided listener
func (s *Server) RunServer() error {
	if s.Listener == nil {
		return errors.New("no listen addr specified")
	}
	s.quit = make(chan struct{})
	s.exited = make(chan struct{})
	for {
		select {
		case <-s.quit:
			s.Wait()
			close(s.exited)
			return nil
		default:
			conn, err := s.Listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				s.Logger.Printf("Error: Failed to accept connection: %s", err.Error())
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if conn == nil {
				s.Logger.Printf("Error: conn is nil")
				continue
			}
			s.Add(1)
			go func() {
				defer handlePanic(s.ErrHandlers)
				defer s.Done()
				s.handleCon(conn)
			}()
		}
	}
}

// Handle incoming connections
func (s *Server) handleCon(conn net.Conn) {
	// create milter object
	milter, actions, protocol := s.MilterFactory()
	var symlists map[Stage]string
	if s.SymListFactory != nil {
		symlists = s.SymListFactory.m
	}
	session := milterSession{
		actions:  actions,
		protocol: protocol,
		sock:     conn,
		milter:   milter,
		logger:   s.Logger,
		symlists: symlists,
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
