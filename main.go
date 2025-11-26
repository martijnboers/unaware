package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
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
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Anonymize data in JSON, XML, CSV, and text files.\n\n")
		fmt.Fprintf(out, "USAGE:\n")
		fmt.Fprintf(out, "  unaware -format <type> -in <infile> [flags]\n\n")
		fmt.Fprintf(out, "EXAMPLES:\n")
		fmt.Fprintf(out, "  # Mask a JSON file using random values\n")
		fmt.Fprintf(out, "  unaware -format json -in input.json -out masked.json\n\n")
		fmt.Fprintf(out, "  # Mask a CSV file, keeping the output consistent between runs\n")
		fmt.Fprintf(out, "  STATIC_SALT=secret-key unaware -format csv -method deterministic -in data.csv > data_masked.csv\n\n")
		fmt.Fprintf(out, "  # Mask only email fields in a large JSON file\n")
		fmt.Fprintf(out, "  unaware -format json -in users.json -include \"*.email\"\n\n")
		fmt.Fprintf(out, "FLAGS:\n")
		flag.PrintDefaults()
	}

	format := flag.String("format", "json", "Format of the input data (json, xml, csv, text)")
	methodFlag := flag.String("method", "random", "Masking method (random or deterministic)")
	inputFile := flag.String("in", "", "Input file path (default: stdin)")
	outputFile := flag.String("out", "", "Output file path (default: stdout)")
	cpuCount := flag.Int("cpu", 4, "Number of CPU cores to use")
	firstN := flag.Int("first", 0, "Process only the first n records/lines (0 means all)")

	var includePatterns, excludePatterns stringSlice
	flag.Var(&includePatterns, "include", "Glob pattern to include keys for masking (can be specified multiple times)")
	flag.Var(&excludePatterns, "exclude", "Glob pattern to exclude keys from masking (can be specified multiple times)")

	flag.Parse()

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
		fmt.Fprintf(os.Stderr, "Error: Invalid method '%s'. Please use 'random' or 'deterministic'.\n", *methodFlag)
		os.Exit(1)
	}

	var reader io.Reader = os.Stdin
	var inputCloser io.Closer
	var fileInfo os.FileInfo

	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
			os.Exit(1)
		}
		fileInfo, err = f.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting file info: %v\n", err)
			os.Exit(1)
		}
		inputCloser = f
		reader = f
	}

	// Show progress bar only when writing to a file (not stdout) and input is a file
	if *outputFile != "" && fileInfo != nil && !fileInfo.IsDir() {
		bar := progressbar.NewOptions64(
			fileInfo.Size(),
			progressbar.OptionSetDescription("Masking..."),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionThrottle(65*1000000), // 65ms
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		)
		progressBarReader := progressbar.NewReader(reader, bar)
		reader = &progressBarReader
	}

	if inputCloser != nil {
		defer inputCloser.Close()
	}

	var writer io.Writer = os.Stdout
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

	if err := pkg.Start(*format, *cpuCount, reader, writer, strategy, includePatterns, excludePatterns, *firstN); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
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

	if *outputFile != "" {
		fmt.Printf("Successfully masked input and saved to %s\n", *outputFile)
	}
}