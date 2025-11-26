package pkg

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

type csvProcessor struct {
	config        AppConfig
	methodFactory func() *masker
}

func newCSVProcessor(config AppConfig) *csvProcessor {
	return &csvProcessor{
		config: config,
		methodFactory: func() *masker {
			return newMasker(config.Masker)
		},
	}
}

func (p *csvProcessor) Process(r io.Reader, w io.Writer) error {
	csvReader := csv.NewReader(r)

	header, err := csvReader.Read()
	if err == io.EOF {
		return nil // Handle empty file
	}
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	// chunkReader reads one CSV row at a time and converts it into a map.
	// This map is the "chunk" our concurrent runner will process, providing the
	// necessary key (column name) for filtering and masking.
	rowCount := 0
	chunkReader := func() (any, error) {
		if p.config.FirstN > 0 && rowCount >= p.config.FirstN {
			return nil, io.EOF
		}
		record, err := csvReader.Read()
		if err != nil {
			return nil, err // Let the runner handle io.EOF
		}
		rowCount++
		rowMap := make(map[string]any, len(header))
		for i, value := range record {
			if i < len(header) {
				rowMap[header[i]] = value
			}
		}
		return rowMap, nil
	}

	assembler := &csvAssembler{
		header: header,
		writer: csv.NewWriter(w),
	}
	runner := newConcurrentRunner(p.methodFactory, p.config)

	return runner.Run(w, chunkReader, assembler)
}

type csvAssembler struct {
	header []string
	writer *csv.Writer
	// A mutex is needed because multiple workers will call WriteItem concurrently.
	mu sync.Mutex
}

func (a *csvAssembler) WriteStart(w io.Writer) error {
	return a.writer.Write(a.header)
}

func (a *csvAssembler) WriteItem(w io.Writer, item any, isFirst bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	rowMap, ok := item.(map[string]any)
	if !ok {
		return fmt.Errorf("csv assembler expected map[string]any, but got %T", item)
	}

	// Convert the map back into a slice of strings in the correct order.
	record := make([]string, len(a.header))
	for i, key := range a.header {
		if val, ok := rowMap[key]; ok {
			record[i] = fmt.Sprintf("%v", val)
		}
	}

	return a.writer.Write(record)
}

func (a *csvAssembler) WriteEnd(w io.Writer) error {
	// The csv.Writer must be flushed to ensure all buffered data is written.
	a.writer.Flush()
	return a.writer.Error()
}
