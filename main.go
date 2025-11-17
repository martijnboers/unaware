package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Define command-line flags
	format := flag.String("format", "json", "The format of the input data (json or xml)")
	consistent := flag.Bool("consistent", true, "Use consistent masking")
	inputFile := flag.String("in", "", "Input file path (default: stdin)")
	outputFile := flag.String("out", "", "Output file path (default: stdout)")
	flag.Parse()

	// Determine the masker to use
	var m Masker
	if *consistent {
		m = NewConsistentMasker()
	} else {
		m = NewRandomMasker()
	}

	// Get the input reader
	var reader io.Reader = os.Stdin
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		reader = f
	}

	// Get the output writer
	var writer io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		writer = f
	}

	// Process the data based on the format
	switch *format {
	case "json":
		jsonMasker := NewJSONMasker(m)
		if err := jsonMasker.Mask(reader, writer); err != nil {
			fmt.Fprintf(os.Stderr, "Error masking JSON: %v\n", err)
			os.Exit(1)
		}
	case "xml":
		xmlMasker := NewXMLMasker(m)
		if err := xmlMasker.Mask(reader, writer); err != nil {
			fmt.Fprintf(os.Stderr, "Error masking XML: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported format: %s\n", *format)
		flag.Usage()
		os.Exit(1)
	}

	if *outputFile != "" {
		fmt.Printf("Successfully masked %s and saved to %s\n", getFileName(*inputFile), *outputFile)
	}
}

func getFileName(path string) string {
	if path == "" {
		return "stdin"
	}
	return filepath.Base(path)
}
