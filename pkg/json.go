package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type jsonProcessor struct {
	methodFactory func() *masker
}

func newJSONProcessor(strategy MaskingStrategy) *jsonProcessor {
	return &jsonProcessor{
		methodFactory: func() *masker {
			return newMasker(strategy)
		},
	}
}

func (jp *jsonProcessor) Process(r io.Reader, w io.Writer, cpuCount int) error {
	br := newPeekingReader(r)
	firstChar, err := br.PeekFirstChar()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	if firstChar == '[' {
		runner := newConcurrentRunner(jp.methodFactory, cpuCount)
		decoder := json.NewDecoder(br)
		decoder.UseNumber()
		_, _ = decoder.Token()
		chunkReader := func() (any, error) {
			if !decoder.More() {
				_, err := decoder.Token()
				if err != nil && err != io.EOF {
					return nil, err
				}
				return nil, io.EOF
			}
			var chunk any
			err := decoder.Decode(&chunk)
			return chunk, err
		}
		return runner.Run(w, chunkReader, &jsonAssembler{})
	}
	return jp.processSerially(br, w)
}

type jsonAssembler struct{}

func (a *jsonAssembler) WriteStart(w io.Writer) error { _, err := w.Write([]byte("[\n")); return err }
func (a *jsonAssembler) WriteItem(w io.Writer, item any, isFirst bool) error {
	if !isFirst {
		if _, err := w.Write([]byte(",\n")); err != nil {
			return err
		}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(item)
}
func (a *jsonAssembler) WriteEnd(w io.Writer) error { _, err := w.Write([]byte("\n]\n")); return err }

type peekingReader struct {
	r    io.Reader
	peek []byte
	err  error
}

func newPeekingReader(r io.Reader) *peekingReader { return &peekingReader{r: r} }

func (pr *peekingReader) Read(p []byte) (n int, err error) {
	if len(pr.peek) > 0 {
		n = copy(p, pr.peek)
		pr.peek = pr.peek[n:]
		return n, nil
	}
	if pr.err != nil {
		return 0, pr.err
	}
	return pr.r.Read(p)
}
func (pr *peekingReader) PeekFirstChar() (byte, error) {
	if pr.err != nil {
		return 0, pr.err
	}
	if len(pr.peek) > 0 {
		for i, c := range pr.peek {
			if !isWhitespace(c) {
				pr.peek = pr.peek[i:]
				return c, nil
			}
		}
	}
	buf := make([]byte, 128)
	for {
		n, err := pr.r.Read(buf)
		if err != nil {
			pr.err = err
			return 0, err
		}
		for i := range n {
			c := buf[i]
			if !isWhitespace(c) {
				pr.peek = buf[i:n]
				return c, nil
			}
		}
	}
}
func isWhitespace(c byte) bool { return c == ' ' || c == '\n' || c == '\r' || c == '\t' }

func (jp *jsonProcessor) processSerially(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	state := &state{wr: w, level: 0, indent: "  ", newline: "\n"}
	serialMasker := jp.methodFactory()
	return jp.processStream(decoder, state, serialMasker)
}

type state struct {
	wr              io.Writer
	level           int
	indent, newline string
}

func (s *state) Indent()   { s.level++ }
func (s *state) Unindent() { s.level-- }

func (jp *jsonProcessor) processStream(dec *json.Decoder, s *state, m *masker) error {
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
			if _, err := io.WriteString(s.wr, "{"); err != nil {
				return err
			}
			s.Indent()
			return jp.processObject(dec, s, m)
		case '[':
			if _, err := io.WriteString(s.wr, "["); err != nil {
				return err
			}
			s.Indent()
			return jp.processArray(dec, s, m)
		default:
			return fmt.Errorf("unexpected delimiter: %v", token)
		}
	default:
		maskedValue := m.mask(t)
		jsonBytes, err := json.Marshal(maskedValue)
		if err != nil {
			return err
		}
		_, err = s.wr.Write(jsonBytes)
		return err
	}
}
func (jp *jsonProcessor) processObject(dec *json.Decoder, s *state, m *masker) error {
	isFirst := true
	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := t.(json.Delim); ok && delim == '}' {
			s.Unindent()
			if !isFirst {
				if _, err := io.WriteString(s.wr, s.newline+strings.Repeat(s.indent, s.level)); err != nil {
					return err
				}
			}
			_, err = io.WriteString(s.wr, "}")
			return err
		}
		if !isFirst {
			if _, err := io.WriteString(s.wr, ","); err != nil {
				return err
			}
		}
		isFirst = false
		if _, err := io.WriteString(s.wr, s.newline+strings.Repeat(s.indent, s.level)); err != nil {
			return err
		}
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected string key in object, got %T", t)
		}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return err
		}
		if _, err := s.wr.Write(keyBytes); err != nil {
			return err
		}
		if _, err := io.WriteString(s.wr, ": "); err != nil {
			return err
		}
		if err := jp.processStream(dec, s, m); err != nil {
			return err
		}
	}
}
func (jp *jsonProcessor) processArray(dec *json.Decoder, s *state, m *masker) error {
	isFirst := true
	for {
		if !dec.More() {
			_, err := dec.Token()
			if err != nil {
				return err
			}
			s.Unindent()
			if !isFirst {
				if _, err := io.WriteString(s.wr, s.newline+strings.Repeat(s.indent, s.level)); err != nil {
					return err
				}
			}
			_, err = io.WriteString(s.wr, "]")
			return err
		}
		if !isFirst {
			if _, err := io.WriteString(s.wr, ","); err != nil {
				return err
			}
		}
		isFirst = false
		if _, err := io.WriteString(s.wr, s.newline+strings.Repeat(s.indent, s.level)); err != nil {
			return err
		}
		if err := jp.processStream(dec, s, m); err != nil {
			return err
		}
	}
}
