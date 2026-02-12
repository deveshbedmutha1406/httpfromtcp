package request

import (
	"bytes"
	"fmt"
	"io"
	"boot.theprimeagen.tv/internal/headers"
	"strconv"
)

// GET /coffee HTTP/1.1
type RequestLine struct {
	HttpVersion   string	// 1.1
	RequestTarget string	// /coffee
	Method        string	// GET POST etc...
}


type parserState string
const (
	StateInit parserState = "init"
	StateDone parserState = "done"
	StateError parserState = "error"
	StateHeaders parserState = "headers"
	StateBody parserState = "body"
)

// request contain multiple thing like request line, headers, body etc...
type Request struct {
	RequestLine RequestLine
	Headers *headers.Headers
	Body string
	state parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: headers.NewHeaders(), 
		Body: "",
	}
}

//parse request line

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request-line ")

var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")

var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("request in error state")

var SEPARATOR = []byte("\r\n")

func getInt(headers *headers.Headers, name string, defaultValue int) int {

	contentLength := headers.Get(name)
	if contentLength == "" {	// content length header not found
		// if there if no cl then we assume there is no body
		return defaultValue
	}
	clength, err := strconv.Atoi(contentLength)
	if err != nil {
		return defaultValue
	}
	return clength
}

func (r *Request) hasBody() bool {
	bodyLen := getInt(r.Headers, "content-length", 0)
	return bodyLen != 0
}

func parseRequestLine(b []byte) (*RequestLine, int, error){
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {	// if we dont find seprator so we need few more bytes return 0 denotes continue.
		return nil, 0, nil
	}

	startLine := b[:idx]	// we found the sep we need info till sep not including it.
	read := idx + len(SEPARATOR)	// this denotes how many bytes we have read / consumed.

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
   
	if !validMethods[string(parts[0])] || string(parts[2]) != "HTTP/1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data[] byte) (int, error) {
	read := 0
	outer:
	for{
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {	// when it have not yet recv rn clrf
				break outer
			}
			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				}else{
					r.state = StateDone
				}
			}
		case StateBody:
			contentLength := getInt(r.Headers,"content-length", 0)
			if contentLength == 0{
				r.state = StateDone
				break outer
			}

			remaining := min((contentLength - len(r.Body)), len(data[read:]))
			
			r.Body += string(data[read:])
		
			read += remaining

			if len(r.Body) == contentLength {
				r.state = StateDone
			}

			if len(data[read:]) == 0 {
				break outer
			}

		case StateDone:
			break outer
		}
	}
	return read, nil
}	

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}
 

func RequestFromReader(reader io.Reader) (*Request, error){

	request := newRequest()

	// NOTE: buf could get overrun... header or body exceeds 1k
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done(){ 
		n, err := reader.Read(buf[bufLen:])
		// TODO: what to do here 

		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])	// parse 
		// out 

		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN: bufLen])
		bufLen -= readN	// this is to delete the stuff which we have read example request line finished then empty the buffer or replace that place.

	}

	return request, nil

}


