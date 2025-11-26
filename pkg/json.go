package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type jsonProcessor struct {
	methodFactory func() *masker
	include       []string
	exclude       []string
}

func newJSONProcessor(strategy MaskingStrategy, include, exclude []string) *jsonProcessor {
	return &jsonProcessor{
		methodFactory: func() *masker {
			return newMasker(strategy)
		},
		include: include,
		exclude: exclude,
	}
}

func (jp *jsonProcessor) Process(r io.Reader, w io.Writer, cpuCount int, firstN int) error {
	br := newPeekingReader(r)
	firstChar, err := br.PeekFirstChar()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	if firstChar == '[' {
		return jp.processRootArray(br, w, cpuCount, firstN)
	}

	// Note: --first is not applied for single root object JSON as there is only one "record".
	return jp.processConcurrentObject(br, w, cpuCount)
}

func (jp *jsonProcessor) processRootArray(r io.Reader, w io.Writer, cpuCount int, firstN int) error {
	runner := newConcurrentRunner(jp.methodFactory, cpuCount, jp.include, jp.exclude)
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	_, _ = decoder.Token() // consume '['
	recordCount := 0
	chunkReader := func() (any, error) {
		if firstN > 0 && recordCount >= firstN {
			return nil, io.EOF
		}
		if !decoder.More() {
			_, err := decoder.Token()
			if err != nil && err != io.EOF {
				return nil, err
			}
			return nil, io.EOF
		}
		var chunk any
		err := decoder.Decode(&chunk)
		recordCount++
		return chunk, err
	}
	return runner.Run(w, chunkReader, &jsonAssembler{isRootArray: true})
}

func (jp *jsonProcessor) processConcurrentObject(r io.Reader, w io.Writer, cpuCount int) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	if _, err := decoder.Token(); err != nil { // consume '{'
		return err
	}
	if _, err := w.Write([]byte("{\n")); err != nil {
		return err
	}

	isFirst := true
	for decoder.More() {
		if !isFirst {
			if _, err := w.Write([]byte(",\n")); err != nil {
				return err
			}
		}
		isFirst = false

		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", keyToken)
		}

		keyBytes, _ := json.Marshal(key)
		if _, err := w.Write(keyBytes); err != nil {
			return err
		}
		if _, err := w.Write([]byte(": ")); err != nil {
			return err
		}

		// We decode the next value into json.RawMessage instead of a fully-parsed interface{}. 
		// This buffers just this single value, allowing us to inspect its type 
		// (e.g., to see if it's an array) without having to buffer the entire io.Reader. 
		// This is a compromise that enables opportunistic concurrency for top-level arrays 
		// in an object while maintaining a streaming approach for the overall structure.
		var rawVal json.RawMessage
		if err := decoder.Decode(&rawVal); err != nil {
			return err
		}

		trimmedVal := bytes.TrimSpace(rawVal)
		if len(trimmedVal) > 0 && trimmedVal[0] == '[' {
			if _, err := w.Write([]byte("[\n")); err != nil {
				return err
			}
			arrDecoder := json.NewDecoder(bytes.NewReader(trimmedVal))
			arrDecoder.UseNumber()
			_, _ = arrDecoder.Token()

			runner := newConcurrentRunner(jp.methodFactory, cpuCount, jp.include, jp.exclude)
			runner.Root = key
			chunkReader := func() (any, error) {
				if !arrDecoder.More() {
					return nil, io.EOF
				}
				var chunk any
				err := arrDecoder.Decode(&chunk)
				return chunk, err
			}
			assembler := &jsonAssembler{}
			if err := runner.Run(w, chunkReader, assembler); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n]")); err != nil {
				return err
			}
		} else {
			valDecoder := json.NewDecoder(bytes.NewReader(rawVal))
			valDecoder.UseNumber()
			var val any
			if err := valDecoder.Decode(&val); err != nil {
				return err
			}
			maskedVal := jp.recursiveMask(jp.methodFactory(), key, val)
			maskedBytes, err := json.MarshalIndent(maskedVal, "  ", "  ")
			if err != nil {
				return err
			}
			if _, err := w.Write(maskedBytes); err != nil {
				return err
			}
		}
	}

	if _, err := w.Write([]byte("\n}")); err != nil {
		return err
	}
	return nil
}

type jsonAssembler struct {
	isRootArray bool
}

func (a *jsonAssembler) WriteStart(w io.Writer) error {
	if a.isRootArray {
		_, err := w.Write([]byte("[\n"))
		return err
	}
	return nil
}

func (a *jsonAssembler) WriteItem(w io.Writer, item any, isFirst bool) error {
	if !isFirst {
		if _, err := w.Write([]byte(",\n")); err != nil {
			return err
		}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("  ", "  ")
	return encoder.Encode(item)
}

func (a *jsonAssembler) WriteEnd(w io.Writer) error {
	if a.isRootArray {
		_, err := w.Write([]byte("\n]\n"))
		return err
	}
	return nil
}

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

func (jp *jsonProcessor) recursiveMask(m *masker, key string, data any) any {
	switch v := data.(type) {
	case json.Number:
		if shouldMask(key, jp.include, jp.exclude) {
			s := v.String()
			if strings.Contains(s, ".") {
				parts := strings.Split(s, ".")
				template := strings.Repeat("#", len(parts[0])) + "." + strings.Repeat("#", len(parts[1]))
				return json.Number(m.faker.Numerify(template))
			}
			return json.Number(m.faker.Numerify(strings.Repeat("#", len(s))))
		}
		return v
	case string, bool, nil:
		if shouldMask(key, jp.include, jp.exclude) {
			return m.mask(v)
		}
		return v
	case map[string]any:
		maskedMap := make(map[string]any, len(v))
		for k, value := range v {
			fullKey := k
			if key != "" {
				fullKey = key + "." + k
			}
			maskedMap[k] = jp.recursiveMask(m, fullKey, value)
		}
		return maskedMap
	case []any:
		maskedSlice := make([]any, len(v))
		for i, value := range v {
			maskedSlice[i] = jp.recursiveMask(m, key, value)
		}
		return maskedSlice
	default:
		return v // Return unsupported types as is
	}
}
