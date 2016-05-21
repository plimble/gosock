package gosocklite

import "bytes"

var DELIM = []byte{32}

func mergeData(event string, data []byte) []byte {
	return bytes.Join([][]byte{[]byte(event), data}, DELIM)
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
