package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/mrcordova/httpfromtcp/internal/request"
	"github.com/mrcordova/httpfromtcp/internal/response"
	"github.com/mrcordova/httpfromtcp/internal/server"
)
const port = 42069
const httpbinUrl = "https://httpbin.org/stream"
func main() {
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

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	} 
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/stream/"){
		n, err := strconv.Atoi(strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/stream/"))
		if err != nil {
			// return
			log.Fatalf("no number found in httpbin route")
		}
		handlerProxy(w, req, n)
		return
	}	
	handler200(w, req)
	return
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handlerProxy(w *response.Writer, _ *request.Request, n int)  {
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	w.WriteHeaders(h)
	// fmt.Print(h)
	resp, err := http.Get(fmt.Sprintf("%s/%v", httpbinUrl, n))
	if err != nil {
		// fmt.Println("here")
		log.Fatalf("get response failed: %s", err)
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 32)
	for {
		bytesRead, err := io.ReadFull(resp.Body, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				log.Fatalf("%s", err)
			}
		}
		w.WriteChunkedBody(buf[:bytesRead])
	}
	w.WriteChunkedBodyDone()

	return
}
