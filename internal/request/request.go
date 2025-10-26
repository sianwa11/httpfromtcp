package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	req, err := parseRequestLine(data)
	if err != nil {
		return &Request{}, err
	}
	return &Request{
		RequestLine: *req,
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, err
	}

	return requestLine, nil
}

func requestLineFromString(str string) (*RequestLine, error) {

	parts := strings.Split(str, " ")

	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}

	method := parts[0]
	requestTarget := parts[1]
	httpVersion := parts[2]

	if strings.ToUpper(method) != method {
		return nil, fmt.Errorf("invalid method, must be uppercase")
	}

	httpParts := strings.Split(httpVersion, "/")
	if httpParts[0] != "HTTP" {
		return nil, fmt.Errorf("unrecognized http version")
	}

	if len(httpParts) < 2 || httpParts[1] != "1.1" {
		return nil, fmt.Errorf("invalid http version")
	}

	if len(strings.Split(requestTarget, " ")) > 1 {
		return nil, fmt.Errorf("invalid request target")
	}

	return &RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: requestTarget,
		Method:        method,
	}, nil

}
