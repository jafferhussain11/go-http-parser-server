package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/jafferhussain11/http-parse/internal/request"
	"github.com/jafferhussain11/http-parse/internal/response"
)

type Server struct {
	isClosed atomic.Bool
	Listener net.Listener
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func Serve(port int, handler Handler) (*Server, error) {
	portString := ":" + strconv.Itoa(port)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Could not create tcp Listener", err)
	}

	server := &Server{
		Listener: l,
		handler:  handler,
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		w := response.NewWriter(conn)
		w.WriteStatusLine(response.StatusBadRequest)
		return
	}

	w := response.NewWriter(conn)
	s.handler(w, req)
}
