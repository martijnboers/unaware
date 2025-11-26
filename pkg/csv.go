package pkg

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

// csvProcessor handles the reading, concurrent masking, and writing of CSV data.
type csvProcessor struct {
	methodFactory func() *masker
	include       []string
	exclude       []string
}

// newCSVProcessor creates a new processor for CSV files.
func newCSVProcessor(strategy MaskingStrategy, include, exclude []string) *csvProcessor {
	return &csvProcessor{
		methodFactory: func() *masker {
			return newMasker(strategy)
		},
		include: include,
		exclude: exclude,
	}
}

// Process orchestrates the reading and assembling of the CSV data.
func (p *csvProcessor) Process(r io.Reader, w io.Writer, cpuCount int, firstN int) error {
	csvReader := csv.NewReader(r)

	// Read the header to get column names.
	header, err := csvReader.Read()
	if err == io.EOF {
		return nil // Handle empty file
	}
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	// The chunkReader function reads one row at a time and converts it into a map,
	// which is the "chunk" our concurrent runner will process.
	rowCount := 0
	chunkReader := func() (any, error) {
		if firstN > 0 && rowCount >= firstN {
			return nil, io.EOF
		}
		record, err := csvReader.Read()
		if err != nil {
			return nil, err // Let the runner handle io.EOF
		}
		rowCount++
		// Convert the row (a slice of strings) into a map using the header.
		// This provides the necessary key (column name) for masking.
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
	runner := newConcurrentRunner(p.methodFactory, cpuCount, p.include, p.exclude)

	return runner.Run(w, chunkReader, assembler)
}

// csvAssembler is responsible for writing the processed data back into a CSV format.
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
	// The concurrent runner might call this from multiple goroutines,
	// so we lock to ensure writes are not interleaved.
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
	// The csv.Writer needs to be flushed to ensure all buffered data is written.
	a.writer.Flush()
	return a.writer.Error()
}
