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
	method := flag.String("method", "random", "Method of masking (random or hashed)")
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

	var masker pkg.Method
	switch *method {
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

		masker = pkg.NewHashedMethod(salt)
	case "random":
		masker = pkg.NewRandomMethod()
	default:
		fmt.Println("No valid method found")
		os.Exit(1)
	}

	app, err := NewApp(*format, masker)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := app.Run(*inputFile, *outputFile, *format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

type Processor interface {
	Mask(r io.Reader, w io.Writer) error
}

type App struct {
	Processor
	In  io.Reader
	Out io.Writer
}

func NewApp(format string, masker pkg.Method) (*App, error) {
	var processor Processor
	switch format {
	case "json":
		processor = pkg.NewJSONProcessor(masker)
	case "xml":
		processor = pkg.NewXMLProcessor(masker)
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
	reader := a.In

	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing memory profile file: %v\n", err)
			}
		}()
		reader = f
	}

	writer := a.Out

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing memory profile file: %v\n", err)
			}
		}()
		writer = f
	}

	return a.Mask(reader, writer)
}
