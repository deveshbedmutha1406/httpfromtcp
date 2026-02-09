package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")	// registered nurse clrf 

func NewHeaders() *Headers{
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	key := strings.ToLower(name)
	if existing := h.headers[key]; existing != "" {
		h.headers[key] = existing + "," + value
	} else {
		h.headers[key] = value
	}
}

/*
Uppercase letters: A-Z
Lowercase letters: a-z
Digits: 0-9
Special characters: 
*/

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9'{
			found = true
		}
		switch ch{
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true

		}
		if !found {
			return false
		}
	}
	return true
}

// gives key value error if any
func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed header")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	// name should not contain any spaces
	if bytes.HasSuffix(name, []byte(" ")){
		return "", "", fmt.Errorf("malformed header")
	}


	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error){
	read := 0	// total bytes read.
	done := false
	// data := []byte("Host: localhost:42069\r\nContent-Type: text/html\r\n")

	for {
		idx := bytes.Index(data[read:], rn)	// start reading from prev read.
		if idx == -1 {
			break
		}

		if idx == 0 {	// empty headers.
			done = true
			read += len(rn)
			break
		}

		name, val, err := parseHeader(data[read:read + idx])
		
		if err != nil {
			return 0, false, err
		}
		if !isToken([]byte(name)){
			return 0, false, fmt.Errorf("malformed header name")
		}
		h.Set(name, val)
		read += idx + len(rn)	// we have taken first header lets move to next one. second header start after rn so idx + len(rn)

	}

	return read, done, nil
}

