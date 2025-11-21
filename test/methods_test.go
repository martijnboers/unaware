package test

import (
	"encoding/json"
	"net"
	"net/url"
	"testing"
	"time"

	"unaware/pkg" // Make sure this import path is correct for your project
)

func TestUnifiedMasker(t *testing.T) {
	testCases := []struct {
		name             string
		mode             string // "hashed" or "random"
		input            any
		assertProperties func(t *testing.T, masker pkg.Method, original any)
	}{
		{
			name:  "URL Hashed",
			mode:  "hashed",
			input: "https://random.org/something",
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 == original {
					t.Error("Hashed URL should not equal original")
				}
				if masked1 != masked2 {
					t.Error("Hashed URL should be deterministic, but got two different values")
				}
				if _, err := url.ParseRequestURI(masked1.(string)); err != nil {
					t.Errorf("Expected a valid URL, but got error: %v", err)
				}
			},
		},
		{
			name:  "URL Random",
			mode:  "random",
			input: "https://random.org/something",
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 == original {
					t.Error("Random URL should not equal original")
				}
				if masked1 == masked2 {
					t.Error("Random URL should not be deterministic, but got the same value twice")
				}
				if _, err := url.ParseRequestURI(masked1.(string)); err != nil {
					t.Errorf("Expected a valid URL, but got error: %v", err)
				}
			},
		},

		// --- Date Tests (Known Format) ---
		{
			name:  "Date Hashed (YYYY-MM-DD)",
			mode:  "hashed",
			input: "2024-10-26",
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 != masked2 {
					t.Error("Hashed Date should be deterministic")
				}
				if _, err := time.Parse("2006-01-02", masked1.(string)); err != nil {
					t.Errorf("Expected format YYYY-MM-DD, but failed to parse: %v", err)
				}
			},
		},
		{
			name:  "Date Random (YYYY-MM-DD)",
			mode:  "random",
			input: "2024-10-26",
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 == masked2 {
					t.Error("Random Date should not be deterministic")
				}
				if _, err := time.Parse("2006-01-02", masked1.(string)); err != nil {
					t.Errorf("Expected format YYYY-MM-DD, but failed to parse: %v", err)
				}
			},
		},

		// --- MAC Address Test ---
		{
			name:  "MAC Address Hashed",
			mode:  "hashed",
			input: "00:00:5e:00:53:01",
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 != masked2 {
					t.Error("Hashed MAC address should be deterministic")
				}
				if _, err := net.ParseMAC(masked1.(string)); err != nil {
					t.Errorf("Expected a valid MAC address, but got error: %v", err)
				}
			},
		},

		// --- Non-String Type Tests ---
		{
			name:  "json.Number Hashed",
			mode:  "hashed",
			input: json.Number("123.45"),
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				masked1 := masker.Mask(original)
				masked2 := masker.Mask(original)

				if masked1 != masked2 {
					t.Error("Hashed json.Number should be deterministic")
				}
				if len(original.(json.Number).String()) != len(masked1.(json.Number).String()) {
					t.Error("Masked json.Number should have same length")
				}
			},
		},
		{
			name:  "Boolean Random",
			mode:  "random",
			input: true,
			assertProperties: func(t *testing.T, masker pkg.Method, original any) {
				// For booleans, we just check that the output is a boolean.
				// Testing for non-determinism is flaky as it could randomly be the same.
				masked := masker.Mask(original)
				if _, ok := masked.(bool); !ok {
					t.Error("Expected masked value to be a boolean")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var masker pkg.Method
			if tc.mode == "hashed" {
				masker = pkg.NewHashedMethod([]byte(testSalt))
			} else {
				masker = pkg.NewRandomMethod()
			}

			if tc.assertProperties != nil {
				tc.assertProperties(t, masker, tc.input)
			}
		})
	}
}

func TestRandomMasker(t *testing.T) {
	m := pkg.NewRandomMethod()

	// Test string masking
	s1 := "hello"
	maskedS1a := m.Mask(s1)
	maskedS1b := m.Mask(s1)

	if maskedS1a == s1 {
		t.Error("String masking should change the value")
	}
	if maskedS1a == maskedS1b {
		t.Error("Random string masking should produce different results for the same input")
	}
}
