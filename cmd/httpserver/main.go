package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
			h := response.GetDefaultHeaders(len(body))
			h.Override("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body := []byte("<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>")
			w.WriteStatusLine(response.StatusInternalServerError)
			h := response.GetDefaultHeaders(len(body))
			h.Override("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)

		} else if req.RequestLine.RequestTarget == "/httpbin/html" {
			handleHttpbinHTML(w, req)

		} else if req.RequestLine.RequestTarget == "/video" {
			handleVideoResponse(w, req)

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			numOfResp := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

			//call the httpbin to get stream
			resp, err := http.Get("https://httpbin.org/" + numOfResp)
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				h := response.GetDefaultHeaders(len(err.Error()))
				w.WriteHeaders(h)
				w.WriteBody([]byte(err.Error()))
				return
			}
			defer resp.Body.Close()

			w.WriteStatusLine(response.StatusOk)
			h := response.GetDefaultHeaders(0)
			h.Delete("content-length")
			h.Override("Transfer-Encoding", "chunked")
			w.WriteHeaders(h)

			buff := make([]byte, 1024)
			//buff := bytes.NewBuffer(arr)

			for {
				n, err := resp.Body.Read(buff)
				if err != nil {
					fmt.Println("Error reading response body:", err)
					break
				}
				fmt.Printf("read %d bytes from httpBin.org", n)
				if n > 0 {
					_, err := w.WriteChunkedBody(buff[:n])
					if err != nil {
						fmt.Println("Error writing chunked body:", err)
						break
					}
				}
			}
			w.WriteChunkedBodyDone()

		} else {
			body := []byte("<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>")
			w.WriteStatusLine(response.StatusOk)
			h := response.GetDefaultHeaders(len(body))
			h.Override("Content-Type", "text/html")
			w.WriteHeaders(h)
			w.WriteBody(body)
		}

	}

	//Server stuff
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

func handleVideoResponse(w *response.Writer, req *request.Request) {
	filePtr, err := os.Open("assets/vim.mp4")
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		h := response.GetDefaultHeaders(len(err.Error()))
		w.WriteHeaders(h)
		w.WriteBody([]byte(err.Error()))
		return
	}
	defer filePtr.Close()

	w.WriteStatusLine(response.StatusOk)
	h := response.GetDefaultHeaders(0)
	h.Delete("content-length")
	h.Override("Transfer-Encoding", "chunked")
	h.Override("Content-Type", "video/mp4")
	h.Override("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	buff := make([]byte, 1024)
	totalBytes := 0
	for {
		n, err := filePtr.Read(buff)
		totalBytes += n
		fmt.Printf("read %d bytes from file", n)
		if n > 0 {
			_, err := w.WriteChunkedBody(buff[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("Done reading response body with no errors")
				break
			} else {
				fmt.Println("Error reading contents of file", err)
				break
			}
		}
	}
	w.WriteChunkedBodyDone()

	trailers := headers.NewHeaders()
	trailers["X-Content-Length"] = fmt.Sprintf("%d", totalBytes)

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing chunked body:", err)
		return
	}
}

func handleHttpbinHTML(w *response.Writer, req *request.Request) {

	resp, err := http.Get("https://httpbin.org/html")
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		h := response.GetDefaultHeaders(len(err.Error()))
		w.WriteHeaders(h)
		w.WriteBody([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOk)
	h := response.GetDefaultHeaders(0)
	h.Delete("content-length")
	h.Override("Transfer-Encoding", "chunked")
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	buff := make([]byte, 1024)
	totalBytes := 0
	hasher := sha256.New()

	for {
		n, err := resp.Body.Read(buff)
		if n > 0 {
			fmt.Printf("read %d bytes from httpBin.org\n", n)
			_, werr := hasher.Write(buff[:n])
			if werr != nil {
				fmt.Println("Error writing chunk into sha256 hasher", werr)
				break
			}
			_, werr = w.WriteChunkedBody(buff[:n])
			if werr != nil {
				fmt.Println("Error writing chunked body:", werr)
				break
			}
			totalBytes += n
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("Done reading response body with no errors")
				break
			} else {
				fmt.Println("Done reading response body with errors", err)
				break
			}
		}
	}
	w.WriteChunkedBodyDone()

	//compute hash
	shaSum := fmt.Sprintf("%x", hasher.Sum(nil))
	trailers := headers.NewHeaders()

	trailers["X-Content-SHA256"] = shaSum
	trailers["X-Content-Length"] = fmt.Sprintf("%d", totalBytes)

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing chunked body:", err)
		return
	}

}
