package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"unaware/pkg"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Anonymize data in JSON, XML, and CSV files by replacing values with realistic-looking fakes.\n\n")
		fmt.Fprintf(os.Stderr, "Use the -method deterministic option to preserve relationships by ensuring identical input values get the same masked output value. \n\n")
		fmt.Fprintf(os.Stderr, "By default every run uses a random salt, use STATIC_SALT=test123 environment variable for consistent masking.")

		flag.PrintDefaults()
	}

	format := flag.String("format", "json", "The format of the input data (json, xml, csv, or text)")
	methodFlag := flag.String("method", "random", "Method of masking (random or deterministic)")
	inputFile := flag.String("in", "", "Input file path (default: stdin)")
	outputFile := flag.String("out", "", "Output file path (default: stdout)")
	cpuCount := flag.Int("cpu", 4, "Numbers of cpu cores used")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile := flag.String("memprofile", "", "write memory profile to `file`")

	var includePatterns, excludePatterns stringSlice
	flag.Var(&includePatterns, "include", "Glob pattern to include keys for masking (can be specified multiple times)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern to exclude keys from masking (can be specified multiple times)")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	var strategy pkg.MaskingStrategy
	switch *methodFlag {
	case "deterministic":
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
		strategy = pkg.Deterministic(salt)
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

	if err := pkg.Start(*format, *cpuCount, reader, writer, strategy, includePatterns, excludePatterns); err != nil {
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
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	if *outputFile != "" {
		fmt.Printf("Successfully masked input and saved to %s\n", *outputFile)
	}
}
