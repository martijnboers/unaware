package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"unaware/pkg"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Anonymize data in JSON and XML files by replacing values with realistic-looking fakes.\n\n")
		fmt.Fprintf(os.Stderr, "Use the -method hashed option to preserve relationships by ensuring identical input values get the same masked output value. By default every run uses a random salt, use STATIC_SALT=test123 environment variable for consistent masking.\n\n")
		flag.PrintDefaults()
	}

	format := flag.String("format", "json", "The format of the input data (json or xml)")
	methodFlag := flag.String("method", "random", "Method of masking (random or hashed)")
	inputFile := flag.String("in", "", "Input file path (default: stdin)")
	outputFile := flag.String("out", "", "Output file path (default: stdout)")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create CPU profile: ", err)
			os.Exit(1)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing cpu profile file: %v\n", err)
			}
		}()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "could not start CPU profile: ", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	var strategy pkg.MaskingStrategy
	switch *methodFlag {
	case "hashed":
		var salt []byte
		if staticSalt := os.Getenv("STATIC_SALT"); staticSalt != "" {
			salt = []byte(staticSalt)
		} else {
			salt = make([]byte, 32)
			if _, err := rand.Read(salt); err != nil {
				fmt.Fprintln(os.Stderr, "failed to generate random salt:", err)
				os.Exit(1)
			}
		}
		strategy = pkg.Hashed(salt)
	case "random":
		strategy = pkg.Random()
	default:
		fmt.Println("No valid method found")
		os.Exit(1)
	}

	reader := os.Stdin
	var inputCloser io.Closer
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
			os.Exit(1)
		}
		inputCloser = f
		reader = f
	}
	if inputCloser != nil {
		defer inputCloser.Close()
	}

	writer := os.Stdout
	var outputCloser io.Closer
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		outputCloser = f
		writer = f
	}

	if err := pkg.Start(*format, reader, writer, strategy); err != nil {
		fmt.Fprintln(os.Stderr, err)
		// Clean up the potentially partially written file on error
		if outputCloser != nil {
			outputCloser.Close()
			os.Remove(*outputFile)
		}
		os.Exit(1)
	}

	if outputCloser != nil {
		if err := outputCloser.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing output file: %v\n", err)
			os.Exit(1)
		}
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create memory profile: ", err)
			os.Exit(1)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing memory profile file: %v\n", err)
			}
		}()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "could not write memory profile: ", err)
			os.Exit(1)
		}
	}

	if *outputFile != "" {
		fmt.Printf("Successfully masked input and saved to %s\n", *outputFile)
	}
}
