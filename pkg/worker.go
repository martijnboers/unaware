package pkg

import (
	"encoding/json"
	"io"
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
}

func newConcurrentRunner(factory func() *masker, cpuCount int) *concurrentRunner {
	return &concurrentRunner{methodFactory: factory, cpuCount: cpuCount}
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
		results <- result{index: j.index, data: recursiveMask(workerMasker, j.data)}
	}
}

func recursiveMask(m *masker, data any) any {
	switch v := data.(type) {
	case json.Number, string, bool, nil:
		return m.mask(v)
	case map[string]any:
		maskedMap := make(map[string]any, len(v))
		for key, value := range v {
			maskedMap[key] = recursiveMask(m, value)
		}
		return maskedMap
	case []any:
		maskedSlice := make([]any, len(v))
		for i, value := range v {
			maskedSlice[i] = recursiveMask(m, value)
		}
		return maskedSlice
	default:
		return m.mask(v)
	}
}
