package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

	// Read all input into a buffer so we can validate it before masking.
	inputBytes, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Validate the input strictly, ensuring no trailing data exists.
	switch *format {
	case "json":
		// Use a decoder to ensure the input is a single, valid JSON entity.
		dec := json.NewDecoder(bytes.NewReader(inputBytes))
		if err := dec.Decode(&struct{}{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Input is not valid JSON. (%v)\n", err)
			os.Exit(1)
		}
		if dec.More() {
			fmt.Fprintf(os.Stderr, "Error: Input contains multiple JSON objects or trailing data.\n")
			os.Exit(1)
		}
	case "xml":
		// The XML decoder is inherently strict and will find syntax errors.
		// We also need to check for trailing data manually.
		dec := xml.NewDecoder(bytes.NewReader(inputBytes))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break // Valid end of document
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Input is not valid XML. (%v)\n", err)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported format: %s\n", *format)
		flag.Usage()
		os.Exit(1)
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

	// Create a new reader from our validated byte slice for the masker.
	inputReader := bytes.NewReader(inputBytes)

	// Process the data based on the format
	switch *format {
	case "json":
		jsonMasker := NewJSONMasker(m)
		if err := jsonMasker.Mask(inputReader, writer); err != nil {
			// This error should ideally not happen now, but we keep it for safety.
			fmt.Fprintf(os.Stderr, "Error masking JSON: %v\n", err)
			os.Exit(1)
		}
	case "xml":
		xmlMasker := NewXMLMasker(m)
		if err := xmlMasker.Mask(inputReader, writer); err != nil {
			fmt.Fprintf(os.Stderr, "Error masking XML: %v\n", err)
			os.Exit(1)
		}
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
