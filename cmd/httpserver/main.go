package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handlerProxy(w, req)
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

func handlerProxy(w *response.Writer, req *request.Request) {
	resource := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := fmt.Sprintf("https://httpbin.org/%s", resource)

	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("failed request to %s, %s", url, err)
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusCode(res.StatusCode))
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	responseTracker := []byte{}
	for {
		n, err := res.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				_, err = w.WriteChunkedBodyDone()
				if err != nil {
					fmt.Println("error writing chunked body done:", err)
				}
				break
			}
			log.Fatalf("error reading into buffer: %s", err)
		}

		_, err = w.WriteChunkedBody(buf[:n])
		if err != nil {
			fmt.Println("error writing chunked body:", err)
		}
		responseTracker = append(responseTracker, buf[:n]...)
	}

	sum := sha256.Sum256(responseTracker)
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sum))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(responseTracker)))
	w.WriteTrailers(trailers)
}
