package response

import (
	"fmt"
	"io"

	"github.com/mrcordova/httpfromtcp/internal/headers"
)


type ResponseLine struct {
	HttpVersion string
	StatusCode StatusCode
	ReasonPhrase string
}
type responseState int

const (
ResponseStatuslineWrite responseState = iota
ResponseHeadersWrite
ResponseBodyWrite
ResponseStartWrite
)


func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := w.Write(getStatusLine(statusCode))
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}



