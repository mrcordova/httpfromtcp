package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/mrcordova/httpfromtcp/internal/request"
	"github.com/mrcordova/httpfromtcp/internal/response"
)
type HandlerError struct {
	StatusCode response.StatusCode
	Message string
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}
type Handler func(w io.Writer, req *request.Request) *HandlerError
// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: listener,
		handler: handler,
	}
	// handlerErr := handler(s.listener)
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}
	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}
	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusCodeSuccess)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
	return
}



