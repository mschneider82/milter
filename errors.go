package milter

import (
	"errors"
)

// pre-defined errors
var (
	ErrCloseSession = errors.New("Stop current milter processing")
	ErrMacroNoData  = errors.New("Macro definition with no data")
	ErrNoListenAddr = errors.New("no listen addr specified")
)
