package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jafferhussain11/http-parse/internal/headers"
	"github.com/jafferhussain11/http-parse/internal/request"
	"github.com/jafferhussain11/http-parse/internal/response"
	"github.com/jafferhussain11/http-parse/internal/server"
)

const port = 42069

func main() {

	handler := func(w *response.Writer, req *request.Request) {

		if req.RequestLine.RequestTarget == "/yourproblem" {
			body := []byte("<html>\n  <head>\n    <title>400 Bad Request</title>\n  </head>\n  <body>\n    <h1>Bad Request</h1>\n    <p>Your request honestly kinda sucked.</p>\n  </body>\n</html>")
			w.WriteStatusLine(response.StatusBadRequest)
			h := headers.NewHeaders()
			h["Content-Type"] = "text/html"
			h["Content-Length"] = strconv.Itoa(len(body))
			w.WriteHeaders(h)
			w.WriteBody(body)

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body := []byte("<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>")
			w.WriteStatusLine(response.StatusInternalServerError)
			h := headers.NewHeaders()
			h["Content-Type"] = "text/html"
			h["Content-Length"] = strconv.Itoa(len(body))
			w.WriteHeaders(h)
			w.WriteBody(body)

		} else {
			body := []byte("<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>")
			w.WriteStatusLine(response.StatusOk)
			h := headers.NewHeaders()
			h["Content-Type"] = "text/html"
			h["Content-Length"] = strconv.Itoa(len(body))
			w.WriteHeaders(h)
			w.WriteBody(body)
		}

	}
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
