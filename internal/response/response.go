package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jafferhussain11/http-parse/internal/headers"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusUnauthorized        StatusCode = 401
	StatusNotFound            StatusCode = 404
	StatusInternalServerError StatusCode = 500
)

type WriterState int

const (
	writerStateStatusLine WriterState = iota // 0 - waiting for status line
	writerStateHeaders                       // 1 - status line done, waiting for headers
	writerStateBody
	writerStateBodyDone
	// 2 - headers done, waiting for body
)

type Writer struct {
	ioWriter    io.Writer
	writerState WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		ioWriter:    w,
		writerState: writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("You are not allowed to write statusLine while on current writer state")
	}

	err := writeStatusLine(w.ioWriter, statusCode)
	if err != nil {
		return err
	}
	w.writerState = writerStateHeaders
	return nil
}

// if user changes existing key overwrite exisitng value
// if user adds extra keys append those
// check non empty and
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("You are not allowed to write statusLine while on current writer state")

	}

	err := writeHeaders(w.ioWriter, headers)
	if err != nil {
		return err
	}
	fmt.Fprintf(w.ioWriter, "\r\n")
	w.writerState = writerStateBody
	return nil

}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in current state")
	}
	return w.ioWriter.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	lenInHex := fmt.Sprintf("%x", len(p))
	n, err := w.ioWriter.Write([]byte(lenInHex))
	if err != nil {
		return 0, err
	}
	n, err = w.ioWriter.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	n, err = w.WriteBody(p)
	if err != nil {
		return 0, err
	}
	w.ioWriter.Write([]byte("\r\n"))

	return n, nil

}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	//write 0 /r/n
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in current state")
	}

	n, err := w.ioWriter.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	w.writerState = writerStateBodyDone
	return n, nil

}

func (w *Writer) WriteTrailers(h headers.Headers) error {

	if w.writerState != writerStateBodyDone {
		return fmt.Errorf("cannot write Trailers in the current state")
	}

	err := writeHeaders(w.ioWriter, h)
	if err != nil {
		return fmt.Errorf("error writing trailers : %s", err)
	}

	_, err = w.ioWriter.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing trailer ending : %s", err)
	}
	return nil

}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-type"] = "text/plain"
	h["connection"] = "close"
	h["content-length"] = strconv.Itoa(contentLen)
	return h
}

func writeStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string

	switch statusCode {
	case StatusOk:
		statusLine = "HTTP/1.1 200 OK"
	case StatusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case StatusUnauthorized:
		statusLine = "HTTP/1.1 401 Unauthorized"
	case StatusNotFound:
		statusLine = "HTTP/1.1 404 Not Found"
	case StatusInternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error"
	default:
		return fmt.Errorf("unsupported status code: %d", statusCode)
	}

	//writes to Writer
	_, err := fmt.Fprintf(w, "%s\r\n", statusLine)
	return err
}

func writeHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
