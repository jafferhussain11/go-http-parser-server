package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/jafferhussain11/http-parse/internal/response"
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

	h := response.GetDefaultHeaders(0)

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Fatalf("error sending response: %s\n", err.Error())
	}

	err = response.WriteHeaders(conn, h)
	if err != nil {
		log.Fatalf("error sending headers: %s\n", err.Error())
	}
	fmt.Fprintf(conn, "\r\n")
	conn.Close()
}
