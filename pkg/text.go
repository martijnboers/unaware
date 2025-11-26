package pkg

import (
	"bufio"
	"io"
	"runtime"
	"sync"
)

type textProcessor struct {
	config AppConfig
}

func newTextProcessor(config AppConfig) *textProcessor {
	return &textProcessor{
		config: config,
	}
}

// Process reads newline-delimited text from r, masks each line concurrently, and writes to w.
func (p *textProcessor) Process(r io.Reader, w io.Writer) error {
	cpuCount := p.config.CPUCount
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
			if p.config.FirstN > 0 && lineCount >= p.config.FirstN {
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
	masker := newMasker(p.config.Masker)
	for line := range jobs {
		maskedLine := masker.mask(line)
		results <- maskedLine.(string)
	}
}
