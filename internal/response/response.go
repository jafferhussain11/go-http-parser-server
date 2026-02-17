package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jafferhussain11/http-parse/internal/headers"
)

type StatusCode int

const (
	statusOk                  StatusCode = 200
	statusBadRequest          StatusCode = 400
	statusUnauthorized        StatusCode = 401
	statusNotFound            StatusCode = 404
	statusInternalServerError StatusCode = 500
)

//type Response struct {
//	statusLine string
//	headers    map[string]string
//}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Type"] = "text/plain"
	h["Connection"] = "close"
	h["Content-Length"] = strconv.Itoa(contentLen)
	return h
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string

	switch statusCode {
	case statusOk:
		statusLine = "HTTP/1.1 200 OK"
	case statusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case statusUnauthorized:
		statusLine = "HTTP/1.1 401 Unauthorized"
	case statusNotFound:
		statusLine = "HTTP/1.1 404 Not Found"
	case statusInternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error"
	default:
		return fmt.Errorf("unsupported status code: %d", statusCode)
	}

	//writes to Writer
	_, err := fmt.Fprintf(w, "%s\r\n", statusLine)
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
