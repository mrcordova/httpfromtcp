package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

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

func handler(w io.Writer, req *request.Request) *server.HandlerError  {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.StatusCode(400),
			Message: "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.StatusCode(500),
			Message: "Woopsie, my bad\n",
		}
	default:
		w.Write([]byte("All good, frfr\n"))
		return nil
	}
}