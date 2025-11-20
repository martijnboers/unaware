package pkg

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Processor interface {
	Mask(r io.Reader, w io.Writer) error
}

type App struct {
	Processor
	In        io.Reader
	Out       io.Writer
}

func NewApp(format string, masker Method) (*App, error) {
	var processor Processor
	switch format {
	case "json":
		processor = NewJSONProcessor(masker)
	case "xml":
		processor = NewXMLProcessor(masker)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return &App{
		Processor: processor,
		In:        os.Stdin,
		Out:       os.Stdout,
	}, nil
}

func (a *App) Run(inputFile, outputFile, format string) error {
	var reader io.Reader = a.In
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer f.Close()
		reader = f
	}

	inputBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	if err := validateInput(inputBytes, format); err != nil {
		return err
	}

	var writer io.Writer = a.Out
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	inputReader := bytes.NewReader(inputBytes)

	return a.Processor.Mask(inputReader, writer)
}

func validateInput(data []byte, format string) error {
	switch format {
	case "json":
		dec := json.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&json.RawMessage{}); err != nil {
			return fmt.Errorf("error: input is not valid JSON: %w", err)
		}
		if dec.More() {
			return fmt.Errorf("error: input contains multiple JSON documents or trailing data")
		}
	case "xml":
		dec := xml.NewDecoder(bytes.NewReader(data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break // Valid end of document
			}
			if err != nil {
				return fmt.Errorf("error: input is not valid XML: %w", err)
			}
		}
	}
	return nil
}
