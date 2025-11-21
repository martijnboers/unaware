package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type JSONProcessor struct {
	Method
}

func NewJSONProcessor(method Method) *JSONProcessor {
	return &JSONProcessor{
		Method: method,
	}
}

type state struct {
	wr      io.Writer
	level   int
	indent  string
	newline string
}

func (s *state) Write(p []byte) (n int, err error) {
	return s.wr.Write(p)
}

func (s *state) Indent() {
	s.level++
}

func (s *state) Unindent() {
	s.level--
}

func (s *state) WriteNewline() {
	s.WriteString(s.newline)
	s.WriteString(strings.Repeat(s.indent, s.level))
}

func (s *state) WriteString(str string) {
	_, _ = io.WriteString(s.wr, str)
}

func (jp *JSONProcessor) Mask(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	// Control indentation manually 
	state := &state{
		wr:      w,
		level:   0,
		indent:  "  ",
		newline: "\n",
	}

	return jp.processStream(decoder, state)
}

func (jp *JSONProcessor) processStream(dec *json.Decoder, s *state) error {
	t, err := dec.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	switch token := t.(type) {
	case json.Delim:
		switch token {
		case '{':
			s.WriteString("{")
			s.Indent()
			return jp.processObject(dec, s)
		case '[':
			s.WriteString("[")
			s.Indent()
			return jp.processArray(dec, s)
		default:
			return fmt.Errorf("unexpected delimiter: %v", token)
		}
	default:
		// Handle top-level primitive values
		maskedValue := jp.Method.Mask(t)
		// Use json.Marshal to get correct JSON formatting (e.g., quotes on strings)
		jsonBytes, err := json.Marshal(maskedValue)
		if err != nil {
			return err
		}
		_, err = s.Write(jsonBytes)
		return err
	}
}

// processObject handles the key-value pairs inside a JSON object.
func (jp *JSONProcessor) processObject(dec *json.Decoder, s *state) error {
	isFirst := true
	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}

		if delim, ok := t.(json.Delim); ok && delim == '}' {
			s.Unindent()
			if !isFirst { // Add a newline before the closing brace if the object was not empty
				s.WriteNewline()
			}
			s.WriteString("}")
			return nil
		}

		if !isFirst {
			s.WriteString(",")
		}
		isFirst = false
		s.WriteNewline()

		// The token must be a string key.
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected string key in object, got %T", t)
		}

		// Marshal the key to get proper quotes and escaping.
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return err
		}
		_, _ = s.Write(keyBytes)
		s.WriteString(": ")

		// Recursively process the value.
		if err := jp.processStream(dec, s); err != nil {
			return err
		}
	}
}

func (jp *JSONProcessor) processArray(dec *json.Decoder, s *state) error {
	isFirst := true
	for {
		if !dec.More() {
			_, err := dec.Token() // Consume the closing ']'
			if err != nil {
				return err
			}
			s.Unindent()
			if !isFirst { // Add a newline before the closing bracket if the array was not empty
				s.WriteNewline()
			}
			s.WriteString("]")
			return nil
		}

		if !isFirst {
			s.WriteString(",")
		}
		isFirst = false
		s.WriteNewline()

		// Recursively process the array element.
		if err := jp.processStream(dec, s); err != nil {
			return err
		}
	}
}
