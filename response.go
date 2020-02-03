package milter

// Response represents a response structure returned by callback
// handlers to indicate how the milter server should proceed
type Response interface {
	Response() *Message
	Continue() bool
}

// SimpleResponse type to define list of pre-defined responses
type SimpleResponse byte

// Response returns a Message object reference
func (r SimpleResponse) Response() *Message {
	return &Message{byte(r), nil}
}

// Continue to process milter messages only if current code is Continue
func (r SimpleResponse) Continue() bool {
	return byte(r) == SMFIR_CONTINUE
}

// Define standard responses with no data
const (
	RespAccept   = SimpleResponse(SMFIR_ACCEPT)
	RespContinue = SimpleResponse(SMFIR_CONTINUE)
	RespDiscard  = SimpleResponse(SMFIR_DISCARD)
	RespReject   = SimpleResponse(SMFIR_REJECT)
	RespTempFail = SimpleResponse(SMFIR_TEMPFAIL)
)

// CustomResponse is a response instance used by callback handlers to indicate
// how the milter should continue processing of current message
type CustomResponse struct {
	code byte
	data []byte
}

// Response returns message instance with data
func (c *CustomResponse) Response() *Message {
	return &Message{c.code, c.data}
}

// Continue returns false if milter chain should be stopped, true otherwise
func (c *CustomResponse) Continue() bool {
	for _, q := range []byte{SMFIR_ACCEPT, SMFIR_DISCARD, SMFIR_REJECT, SMFIR_TEMPFAIL} {
		if c.code == q {
			return false
		}
	}
	return true
}

// NewResponse generates a new CustomResponse suitable for WritePacket
func NewResponse(code byte, data []byte) *CustomResponse {
	return &CustomResponse{code, data}
}

// NewResponseStr generates a new CustomResponse with string payload
// code should be SMFIR_REPLYCODE == 'y'
// data can be "550 5.2.0 mailbox unavailable."
func NewResponseStr(code byte, data string) *CustomResponse {
	return NewResponse(code, []byte(data+null))
}
