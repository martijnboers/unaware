package pkg

import (
	"bufio"
	"io"
)

// textProcessor processes unstructured text data line by line.
type textProcessor struct {
	masker *masker
}

// newTextProcessor creates a new processor for plain text data.
func newTextProcessor(strategy MaskingStrategy, include, exclude []string) *textProcessor {
	return &textProcessor{
		masker: newMasker(strategy),
	}
}

// Process reads newline-delimited text from r, masks each line, and writes to w.
// It is designed for text-based data; binary or non-UTF-8 input will be
// processed on a best-effort basis, which may produce nonsensical output but
// will not cause a crash.
func (p *textProcessor) Process(r io.Reader, w io.Writer, _ int) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)

	// Increase the buffer size to handle long lines, preventing bufio.ErrTooLong.
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		maskedLine := p.masker.mask(line)
		if _, err := writer.WriteString(maskedLine.(string) + "\n"); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return writer.Flush()
}
