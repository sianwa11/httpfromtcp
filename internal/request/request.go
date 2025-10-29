package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

const bufferSize = 8

type ParserState int

const (
	initalized ParserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		State:   initalized,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}

	for {
		if readToIndex >= len(buf) {
			newBuff := make([]byte, len(buf)*2)
			copy(newBuff, buf[:readToIndex])
			buf = newBuff
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.State != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.State, n)
				}
			}
			return nil, err
		}
		readToIndex += n

		parsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
			readToIndex -= parsed
		}

		if req.State == requestStateDone {
			break
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		totalBytesParsed += n
		if n == 0 {
			break
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case initalized:
		n, req, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *req
		r.State = requestStateParsingHeaders
		return n, nil

	case requestStateParsingHeaders:
		n, doneParsing, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if doneParsing {
			r.State = requestStateParsingBody
		}

		return n, nil

	case requestStateParsingBody:
		contentLengthStr := r.Headers.Get("Content-Length")
		if contentLengthStr == "" {
			r.State = requestStateDone
			return len(data), nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %v", err)
		}

		if len(data) < contentLength {
			return 0, nil
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) >= contentLength {
			r.State = requestStateDone
			return contentLength, nil
		}

		return contentLength, nil

	case requestStateDone:
		return 0, fmt.Errorf("error trying to read data in a requestStateDone state")

	default:
		return 0, fmt.Errorf("error trying to read data in unknown state")

	}
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, nil, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return idx + 2, nil, err
	}

	return idx + 2, requestLine, nil
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
