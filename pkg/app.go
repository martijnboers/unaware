package pkg

import (
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

	var writer io.Writer = a.Out

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	return a.Processor.Mask(reader, writer)
}
