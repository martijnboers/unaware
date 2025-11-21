package pkg

import (
	"fmt"
	"io"
)

type MaskingStrategy func(*masker)

type processor interface {
	Process(r io.Reader, w io.Writer) error
}

// Hashed is a public "recipe". It returns a MaskingStrategy function that
// configures the processor to use deterministic, hash-based masking.
func Hashed(salt []byte) MaskingStrategy {
	return func(m *masker) {
		m.seeder = &hashedSeeder{salt: salt}
	}
}

// Random is a public "recipe". It returns a MaskingStrategy function that
// configures the processor to use non-deterministic, random masking.
func Random() MaskingStrategy {
	return func(m *masker) {
		m.seeder = &randomSeeder{}
	}
}

// Process is the single, top-level entry point for the entire package.
// It is the "chef" that takes the order and delegates to the correct station.
func Process(format string, r io.Reader, w io.Writer, strategy MaskingStrategy) error {
	var p processor
	switch format {
	case "json":
		p = newJSONProcessor(strategy)
	case "xml":
		p = newXMLProcessor(strategy)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return p.Process(r, w)
}
