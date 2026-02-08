package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// GET /coffee HTTP/1.1
type RequestLine struct {
	HttpVersion   string	// 1.1
	RequestTarget string	// /coffee
	Method        string	// GET POST etc...
}


// request contain multiple thing like request line, headers, body etc...
type Request struct {
	RequestLine RequestLine
}

//parse request line

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request-line ")

var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")

var SEPARATOR = "\r\n"



func parseRequestLine(b string) (*RequestLine, string, error){
	idx := strings.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, b, nil
	}

	startLine := b[:idx]
	restOfMsg := b[idx + len(SEPARATOR): ]

	parts := strings.Split(startLine, " ")

	if len(parts) != 3 {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
   
	if !validMethods[parts[0]] || parts[2] != "HTTP/1.1" {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method: parts[0],
		RequestTarget: parts[1],
		HttpVersion: httpParts[1],
	}

	return rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error){

	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to io.ReadAll"),
			err,
		)
	}

	str := string(data)

	rl, _, err := parseRequestLine(str)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Error in request parsing"), err)
	}
	return &Request{
		RequestLine: *rl,
	}, err

}


