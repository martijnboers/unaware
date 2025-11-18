package test

import (
	"net"
	"strconv"
	"testing"
	"time"
	"unaware/pkg"
)

func TestExtendedMasking(t *testing.T) {
	masker := pkg.NewSaltedMethod([]byte("test-salt"))

	tests := []struct {
		name         string
		input        string
		assertOutput func(t *testing.T, original, masked string)
	}{
		{
			name:  "Different Date Format (MM/DD/YYYY)",
			input: "11/18/2025",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("Date was not masked")
				}
				if _, err := time.Parse("01/02/2006", masked); err != nil {
					t.Errorf("Masked value is not a valid date in MM/DD/YYYY format: %s", masked)
				}
			},
		},
		{
			name:  "Another Date Format (YYYY-MM)",
			input: "2025-11",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("Date was not masked")
				}
				if _, err := time.Parse("2006-01", masked); err != nil {
					t.Errorf("Masked value is not a valid date in YYYY-MM format: %s", masked)
				}
			},
		},
		{
			name:  "MAC Address",
			input: "00:00:5e:00:53:01",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("MAC address was not masked")
				}
				if _, err := net.ParseMAC(masked); err != nil {
					t.Errorf("Masked value is not a valid MAC address: %s", masked)
				}
			},
		},
		{
			name:  "Generic Integer String",
			input: "1234567890",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("Generic integer string was not masked")
				}
				if _, err := strconv.ParseInt(masked, 10, 64); err != nil {
					t.Errorf("Masked value is not a valid integer string: %s", masked)
				}
				if len(masked) != len(original) {
					t.Errorf("Masked integer string has different length. Original: %d, Masked: %d", len(original), len(masked))
				}
			},
		},
		{
			name:  "Generic Float String",
			input: "12345.67890",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("Generic float string was not masked")
				}
				if _, err := strconv.ParseFloat(masked, 64); err != nil {
					t.Errorf("Masked value is not a valid float string: %s", masked)
				}
				if len(masked) != len(original) {
					t.Errorf("Masked float string has different length. Original: %d, Masked: %d", len(original), len(masked))
				}
			},
		},
		{
			name:  "Fallback Word Masking",
			input: "ThisIsJustSomeText",
			assertOutput: func(t *testing.T, original, masked string) {
				if masked == original {
					t.Errorf("Fallback string was not masked")
				}
				// It should not be a number
				if _, err := strconv.ParseFloat(masked, 64); err == nil {
					t.Errorf("Fallback string should not be a number, but it was: %s", masked)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maskedValue := masker.Mask(tt.input).(string)
			tt.assertOutput(t, tt.input, maskedValue)
		})
	}
}
