package pkg

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
)

type chunkReader func() (any, error)
type assembler interface {
	WriteStart(w io.Writer) error
	WriteItem(w io.Writer, item any, isFirst bool) error
	WriteEnd(w io.Writer) error
}
type job struct {
	index int
	data  any
}
type result struct {
	index int
	data  any
}

type concurrentRunner struct {
	methodFactory func() *masker
	cpuCount      int
	include       []string
	exclude       []string
	Root          string
}

func newConcurrentRunner(factory func() *masker, cpuCount int, include, exclude []string) *concurrentRunner {
	return &concurrentRunner{
		methodFactory: factory,
		cpuCount:      cpuCount,
		include:       include,
		exclude:       exclude,
	}
}

func (cr *concurrentRunner) Run(w io.Writer, crr chunkReader, a assembler) error {
	jobs := make(chan job)
	results := make(chan result)
	var wg sync.WaitGroup
	for range cr.cpuCount {
		wg.Add(1)
		go cr.worker(&wg, jobs, results)
	}
	var dispatchErr error
	go func() {
		jobIndex := 0
		for {
			dataChunk, err := crr()
			if err == io.EOF {
				break
			}
			if err != nil {
				dispatchErr = err
				break
			}
			jobs <- job{index: jobIndex, data: dataChunk}
			jobIndex++
		}
		close(jobs)
	}()
	go func() { wg.Wait(); close(results) }()

	if err := a.WriteStart(w); err != nil {
		return err
	}
	resultsBuffer := make(map[int]any)
	nextIndexToWrite := 0
	isFirst := true
	for res := range results {
		resultsBuffer[res.index] = res.data
		for {
			maskedData, ok := resultsBuffer[nextIndexToWrite]
			if !ok {
				break
			}
			if err := a.WriteItem(w, maskedData, isFirst); err != nil {
				return err
			}
			isFirst = false
			delete(resultsBuffer, nextIndexToWrite)
			nextIndexToWrite++
		}
	}
	if dispatchErr != nil {
		return dispatchErr
	}
	return a.WriteEnd(w)
}

func (cr *concurrentRunner) worker(wg *sync.WaitGroup, jobs <-chan job, results chan<- result) {
	defer wg.Done()
	workerMasker := cr.methodFactory()
	for j := range jobs {
		results <- result{index: j.index, data: cr.recursiveMask(workerMasker, cr.Root, j.data)}
	}
}

func (cr *concurrentRunner) recursiveMask(m *masker, key string, data any) any {
	switch v := data.(type) {
	case json.Number, string, bool, nil:
		if shouldMask(key, cr.include, cr.exclude) {
			return m.mask(v)
		}
		return v
	case map[string]any:
		maskedMap := make(map[string]any, len(v))
		for k, value := range v {
			if k == "#text" {
				// This is the text content of the parent element (e.g., the "2002" in <year>2002</year>).
				// The key for filtering is the parent's key, which is already in the 'key' variable.
				if shouldMask(key, cr.include, cr.exclude) {
					maskedMap[k] = m.mask(value)
				} else {
					maskedMap[k] = value
				}
			} else {
				// This is a nested element or an attribute.
				// Attributes from the XML decoder are prefixed with '-'.
				nestedKey := strings.TrimPrefix(k, "-")
				fullKey := nestedKey
				if key != "" {
					fullKey = key + "." + nestedKey
				}
				maskedMap[k] = cr.recursiveMask(m, fullKey, value)
			}
		}
		return maskedMap
	case []any:
		maskedSlice := make([]any, len(v))
		for i, value := range v {
			maskedSlice[i] = cr.recursiveMask(m, key, value)
		}
		return maskedSlice
	default:
		if shouldMask(key, cr.include, cr.exclude) {
			return m.mask(v)
		}
		return v
	}
}
