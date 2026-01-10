package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		//return full map size
		return 2, true, nil
	}
	dataString := string(data[:idx])
	key, value, err := sanitizeDataString(dataString)
	if err != nil {
		return 0, false, err
	}

	h[key] = value
	return len(dataString) + 2, false, err

}

func sanitizeDataString(dataString string) (string, string, error) {

	trimmedString := strings.Trim(dataString, " ")
	//get first idx of colon
	colonIdx := strings.Index(trimmedString, ":")

	//remove ws b/w key and colon
	key := strings.TrimPrefix(trimmedString[:colonIdx], " ")
	if strings.Contains(key, " ") {
		return "", "", fmt.Errorf("malformed header key, no spacing b/w key and colon")
	}

	value := strings.Trim(trimmedString[colonIdx+1:], " ")
	return key, value, nil

}
