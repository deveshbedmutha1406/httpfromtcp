package server

import (
	"net"
	"fmt"
	"io"
	"boot.theprimeagen.tv/internal/response"
	"boot.theprimeagen.tv/internal/request"

)

type HandlerError struct {
	StatusCode response.StatusCode
	Messsage string
}
type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed bool
	handler Handler
}


func runConnection(s *Server, conn io.ReadWriteCloser){
	defer conn.Close()

	responseWriter := response.NewWriter(conn) 

	r ,err := request.RequestFromReader(conn)
	if err != nil {	
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return 
	}

	s.handler(responseWriter, r)
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