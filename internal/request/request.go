package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const bufferSize = 8

type ParserState int

const (
	initalized ParserState = iota
	done
)

type Request struct {
	RequestLine RequestLine
	State       ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
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
		r.State = done
		return n, nil

	case done:
		return 0, fmt.Errorf("error trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error trying to read data in a done state")

	}

}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		State: initalized,
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
				req.State = done
				break
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

		if req.State == done {
			break
		}
	}

	return req, nil
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
