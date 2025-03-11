package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const crlf = "\r\n"
const specialCharacters = "!#$%&'*+-.^_`|~"
type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	idx := bytes.Index(data, []byte(crlf))
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

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)

	if len(key) < 1 {
		return 0, false, fmt.Errorf("Key is too short")
	}
	if !validTokens(key) {
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}
	h.Set(key, string(value))
	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	lowerKey := strings.ToLower(key)
	if _, ok := h[lowerKey]; ok {
		h[lowerKey] += fmt.Sprintf(", %s", value)
	} else {
		h[lowerKey] = value
	}
}

func validTokens(data string) bool {
	for _, char := range data {
		if unicode.IsDigit(char) == false && unicode.IsLetter(char) == false && strings.Contains(specialCharacters, string(char)) == false  {
			fmt.Println("here")
			return false
		}
		
	}
	return true
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}