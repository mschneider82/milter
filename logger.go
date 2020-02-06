package milter

import "log"

// CustomLogger is a interface to inject a custom logger
type CustomLogger interface {
	Printf(format string, v ...interface{})
}

type nopLogger struct{}

func (n nopLogger) Printf(format string, v ...interface{}) {}

// NopLogger can be used to discard all logs caused by milter library
var NopLogger = CustomLogger(nopLogger{})

// StdOutLogger is the default logger used if no Logger was supplied
var StdOutLogger = CustomLogger(stdoutLogger{})

type stdoutLogger struct{}

func (s stdoutLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
