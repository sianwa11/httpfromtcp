package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
	"sync/atomic"
)

type Handler func(w *response.Writer, req *request.Request)

// Server is a HTTP 1.1 server
type Server struct {
	handler  Handler
	Listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		handler:  handler,
		Listener: listener,
	}

	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return nil
			}
			return fmt.Errorf("error accepting connection: %s", err)
		}

		fmt.Println("connection has been accepted")

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	s.handler(w, req)
	return
}

// type HandlerError struct {
// 	StatusCode response.StatusCode
// 	Message    string
// }

// func (he *HandlerError) Write(w io.Writer) {
// 	writer := response.NewWriter(w)
// 	writer.WriteStatusLine(he.StatusCode)
// 	messageBytes := []byte(he.Message)
// 	headers := response.GetDefaultHeaders(len(messageBytes))
// 	writer.WriteHeaders(headers)
// 	writer.WriteBody(messageBytes)
// }
