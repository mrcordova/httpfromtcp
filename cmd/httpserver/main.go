package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mrcordova/httpfromtcp/internal/headers"
	"github.com/mrcordova/httpfromtcp/internal/request"
	"github.com/mrcordova/httpfromtcp/internal/response"
	"github.com/mrcordova/httpfromtcp/internal/server"
)
const port = 42069
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
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"){
		handlerProxy(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" && req.RequestLine.Method == "GET" {
		handlerVideo(w, req)
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

func handlerProxy(w *response.Writer, req *request.Request)  {
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")	

	h.Remove("Content-Length")
	// fmt.Println(h)
	w.WriteHeaders(h)
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 32)
	body := make([]byte, 0)
	// totalBytesRead := 0
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
		body = append(body, buf[:bytesRead]...)
		// totalBytesRead += bytesRead
	}
	hash := sha256.Sum256(body)
	
	w.WriteTrailers(headers.Headers{
		"X-Content-Sha256": hex.EncodeToString(hash[:]),
		"X-Content-Length": fmt.Sprint(len(body)),
	})
	w.WriteChunkedBodyDone()

	return
}
func handlerVideo(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Override("Content-Type", "video/mp4")
	h.Remove("Content-Length")
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Fatal(err)
	}
	w.WriteStatusLine(response.StatusCodeSuccess)
	w.WriteHeaders(h)
	w.WriteBody(data)
	// h.Set("Transfer-Encoding", "chunked")
	// h.Override("Trailer", "X-Content-SHA256, X-Content-Length")	

	h.Remove("Content-Length")
}