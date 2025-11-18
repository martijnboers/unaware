package test

import (
	"encoding/json"
	"net"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
	"unaware/pkg"
)

func TestTypePreservation(t *testing.T) {
	masker := pkg.NewSaltedMethod([]byte("test-salt"))

	tests := []struct {
		name         string
		input        any
		assertOutput func(t *testing.T, original, masked any)
	}{
		{
			name:  "String (Generic Word)",
			input: "Some random text",
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(string) == original.(string) {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "String (URL)",
			input: "https://example.com",
			assertOutput: func(t *testing.T, original, masked any) {
				u, err := url.Parse(masked.(string))
				if err != nil {
					t.Fatalf("Masked URL is not a valid URL: %s", masked)
				}
				if !strings.HasSuffix(u.Hostname(), ".local") {
					t.Errorf("Masked URL hostname should end with .local, but got: %s", u.Hostname())
				}
			},
		},
		{
			name:  "String (Email)",
			input: "test@example.com",
			assertOutput: func(t *testing.T, original, masked any) {
				if !strings.HasSuffix(masked.(string), ".local") {
					t.Errorf("Masked email should end with .local, but got: %s", masked)
				}
			},
		},
		{
			name:  "String (IPv4)",
			input: "192.168.1.1",
			assertOutput: func(t *testing.T, original, masked any) {
				if net.ParseIP(masked.(string)) == nil {
					t.Errorf("Masked value is not a valid IP address: %s", masked)
				}
			},
		},
		{
			name:  "String (IPv6)",
			input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			assertOutput: func(t *testing.T, original, masked any) {
				if net.ParseIP(masked.(string)) == nil {
					t.Errorf("Masked value is not a valid IP address: %s", masked)
				}
			},
		},
		{
			name:  "String (MAC Address)",
			input: "00:00:5e:00:53:01",
			assertOutput: func(t *testing.T, original, masked any) {
				if _, err := net.ParseMAC(masked.(string)); err != nil {
					t.Errorf("Masked value is not a valid MAC address: %s", masked)
				}
			},
		},
		{
			name:  "String (Date)",
			input: "2025-11-18",
			assertOutput: func(t *testing.T, original, masked any) {
				if _, err := time.Parse("2006-01-02", masked.(string)); err != nil {
					t.Errorf("Masked value is not a valid date: %s", masked)
				}
			},
		},
		{
			name:  "String (Generic Integer)",
			input: "987654321",
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(string) == original.(string) {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "String (Generic Float)",
			input: "987.654",
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(string) == original.(string) {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "JSON Number (Integer)",
			input: json.Number("12345"),
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(json.Number).String() == original.(json.Number).String() {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "JSON Number (Float)",
			input: json.Number("123.45"),
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(json.Number).String() == original.(json.Number).String() {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "JSON Number (Large Integer)",
			input: json.Number("67027460809240096712491686925493061126"),
			assertOutput: func(t *testing.T, original, masked any) {
				if masked.(json.Number).String() == original.(json.Number).String() {
					t.Errorf("Value was not masked")
				}
			},
		},
		{
			name:  "Boolean",
			input: true,
			assertOutput: func(t *testing.T, original, masked any) {
				// Booleans are random, so they might be the same.
				// We just check the type.
			},
		},
		{
			name:  "Nil",
			input: nil,
			assertOutput: func(t *testing.T, original, masked any) {
				if masked != nil {
					t.Errorf("Nil should be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maskedValue := masker.Mask(tt.input)

			originalType := reflect.TypeOf(tt.input)
			maskedType := reflect.TypeOf(maskedValue)

			if originalType != maskedType {
				t.Errorf("Type was not preserved. Original type: %v, Masked type: %v", originalType, maskedType)
			}

			tt.assertOutput(t, tt.input, maskedValue)
		})
	}
}
