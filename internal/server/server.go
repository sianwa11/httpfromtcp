package server

import (
	"fmt"
	"httpfromtcp/internal/response"
	"net"
	"sync/atomic"
)

type Server struct {
	Address  string
	Listener net.Listener
	closed   atomic.Bool
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.Listener.Close()
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

		go func() {
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func Serve(port int) (*Server, error) {
	portStr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", portStr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Address:  portStr,
		Listener: listener,
	}

	go func() {
		err := server.listen()
		if err != nil {
			fmt.Printf("error in server listen: %v", err)
		}
	}()

	return server, nil
}
