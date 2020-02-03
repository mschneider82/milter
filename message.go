package milter

// Message represents a command sent from milter client
type Message struct {
	Code byte
	Data []byte
}
