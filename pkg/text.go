package pkg

import (
	"bufio"
	"io"
	"runtime"
	"sync"
)

// textProcessor processes unstructured text data line by line.
type textProcessor struct {
	strategy MaskingStrategy
}

// newTextProcessor creates a new processor for plain text data.
func newTextProcessor(strategy MaskingStrategy, _, _ []string) *textProcessor {
	return &textProcessor{
		strategy: strategy,
	}
}

// Process reads newline-delimited text from r, masks each line concurrently, and writes to w.
func (p *textProcessor) Process(r io.Reader, w io.Writer, cpuCount int, firstN int) error {
	if cpuCount <= 0 {
		cpuCount = runtime.NumCPU()
	}

	jobs := make(chan string, cpuCount)
	results := make(chan string, cpuCount)

	// Start worker pool
	wg := &sync.WaitGroup{}
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go p.worker(wg, jobs, results)
	}

	// Start a goroutine to read the file and send lines to the jobs channel
	go func() {
		scanner := bufio.NewScanner(r)
		const maxCapacity = 1024 * 1024 // 1MB
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		lineCount := 0
		for scanner.Scan() {
			if firstN > 0 && lineCount >= firstN {
				break
			}
			jobs <- scanner.Text()
			lineCount++
		}
		close(jobs)
	}()

	// Start a goroutine to close the results channel once all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Write results to the output
	writer := bufio.NewWriter(w)
	defer writer.Flush()
	for result := range results {
		if _, err := writer.WriteString(result + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func (p *textProcessor) worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- string) {
	defer wg.Done()
	masker := newMasker(p.strategy)
	for line := range jobs {
		maskedLine := masker.mask(line)
		results <- maskedLine.(string)
	}
}
