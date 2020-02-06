package milter

import (
	"log"
	"net"
)

// An Option configures a Server using the functional options paradigm
// popularized by Rob Pike.
type Option interface {
	apply(*Server)
}

// An ListenerOption configures a Server using the functional options paradigm
// popularized by Rob Pike.
type ListenerOption interface {
	lapply(*Server)
}

type optionFunc func(*Server)

func (f optionFunc) apply(Server *Server) { f(Server) }

type loptionFunc func(*Server)

func (f loptionFunc) lapply(Server *Server) { f(Server) }

// WithLogger adds an Logger
func WithLogger(l CustomLogger) Option {
	return optionFunc(func(server *Server) {
		server.logger = l
	})
}

// WithListener adds an Listener
func WithListener(listener net.Listener) ListenerOption {
	return loptionFunc(func(server *Server) {
		server.listener = listener
	})
}

// WithTCPListener e.g. "127.0.0.1:12349"
func WithTCPListener(address string) ListenerOption {
	protocol := "tcp"
	// bind to listening address
	socket, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatal(err)
	}

	return WithListener(socket)
}

// WithUnixSocket e.g. "/var/spool/postfix/var/run/milter/milter.sock"
// make sure that the file does not exist!
func WithUnixSocket(file string) ListenerOption {
	protocol := "unix"
	// bind to listening address
	socket, err := net.Listen(protocol, file)
	if err != nil {
		log.Fatal(err)
	}

	return WithListener(socket)
}

// WithPanicHandler Adds the error panic handler
// Multiple panic handlers are supported
func WithPanicHandler(handler func(error)) Option {
	return optionFunc(func(server *Server) {
		server.errHandlers = append(server.errHandlers, handler)
	})
}
