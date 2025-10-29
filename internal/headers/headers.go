package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

const CRLF = "\r\n"

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	key = strings.TrimSpace(key)
	value := bytes.TrimSpace(parts[1])
	if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}

	h.Set(key, string(value))

	return idx + len(CRLF), false, nil
}

func (h Headers) Set(key, value string) {
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{v, value}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	v, ok := h[key]
	if !ok {
		return ""
	}
	return v
}

var fieldNamePattern = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`)

func isValidFieldName(s string) bool {
	return fieldNamePattern.MatchString(s)
}
