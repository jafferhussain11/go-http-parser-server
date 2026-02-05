package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	isClosed atomic.Bool
	Listener net.Listener
}

func Serve(port int) (*Server, error) {
	portString := ":" + strconv.Itoa(port)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Could not create tcp Listener", err)
	}

	server := &Server{
		Listener: l,
	}
	server.isClosed.Store(false)

	//this frees the main goroutine
	go server.listen()

	return server, nil

}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		log.Fatal("error closing server", err)
	}
	return nil
}

func (s *Server) listen() {

	for !s.isClosed.Load() {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		//this frees the listening goroutine
		go func(c net.Conn) {
			s.handle(c)
			c.Close()
		}(conn)
	}

}

func (s *Server) handle(conn net.Conn) {
	//write to stream, not console ! like fmt.Printf
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nHello World!")
	conn.Close()
}
