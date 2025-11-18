package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"

	"unaware/pkg"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Anonymize data in JSON and XML files by replacing values with realistic-looking fakes.\n\n")
		fmt.Fprintf(os.Stderr, "Use the -random-hash flag to preserve relationships by ensuring identical input values get the same masked output.\n\n")
		flag.PrintDefaults()
	}

	format := flag.String("format", "json", "The format of the input data (json or xml)")
	inputFile := flag.String("in", "", "Input file path (default: stdin)")
	outputFile := flag.String("out", "", "Output file path (default: stdout)")
	randomHash := flag.Bool("random-hash", false, "Hash data using random salt")
	flag.Parse()

	var masker pkg.Masker
	if *randomHash {
		salt := make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			fmt.Fprintln(os.Stderr, "failed to generate random salt:", err)
			os.Exit(1)
		}
		masker = pkg.NewSaltedMasker(salt)
	} else {
		masker = pkg.NewRandomMasker()
	}

	app, err := pkg.NewApp(*format, masker)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := app.Run(*inputFile, *outputFile, *format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *outputFile != "" {
		fmt.Printf("Successfully masked input and saved to %s\n", *outputFile)
	}
}
