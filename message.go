package gosock

import "bytes"

var DELIM = []byte{32}

type responseMessage struct {
	room       string
	event      string
	to         string
	data       []byte
	connection *Connection
}

func newResponseMessage(room, event, to string, data []byte, c *Connection) *responseMessage {
	return &responseMessage{room, event, to, data, c}
}

func (r *responseMessage) Data() []byte {
	return bytes.Join([][]byte{[]byte(r.event), r.data}, DELIM)
}

func getRequestMessage(msg []byte) (string, []byte) {
	result := bytes.SplitN(msg, DELIM, 2)
	var event string
	var data []byte
	switch len(result) {
	case 1:
		event = string(result[0])
	case 2:
		event = string(result[0])
		data = result[1]
	}

	return event, data
}
