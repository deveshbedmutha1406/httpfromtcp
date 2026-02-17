package server

import (
	"net"
	"fmt"
	"io"
	"bytes"
	"boot.theprimeagen.tv/internal/response"
	"boot.theprimeagen.tv/internal/request"

)

type HandlerError struct {
	StatusCode response.StatusCode
	Messsage string
}
type Handler func(w io.Writer, req *request.Request) *HandlerError 

type Server struct {
	closed bool
	handler Handler
}


func runConnection(s *Server, conn io.ReadWriteCloser){
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	
	r ,err := request.RequestFromReader(conn)
	if err != nil {	
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return 
		
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, r)
	var body []byte = nil
	var status response.StatusCode = response.StatusOk

	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Messsage)	
	}else{
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func runServer(s *Server, listener net.Listener) error {

	for {
		conn, err := listener.Accept()
		if s.closed {
			return nil
		}
		if err != nil {
			return err
		}
		go runConnection(s, conn)
	}	 
	
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}


	server := &Server {closed: false, handler: handler,}
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}