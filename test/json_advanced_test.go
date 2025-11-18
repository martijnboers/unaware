package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"unaware/pkg"
)

func TestJSONMasker_Advanced(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		asserter   func(t *testing.T, output string)
		randomHash bool
	}{
		{
			name:  "integer should not become float",
			input: `{"key": 123}`,
			asserter: func(t *testing.T, output string) {
				maskedValue := gjson.Get(output, "key")
				if maskedValue.Type != gjson.Number {
					t.Errorf("Expected key to be a number, but got %s", maskedValue.Type)
				}
				if maskedValue.Int() == 123 {
					t.Errorf("Expected integer to be masked, but it was unchanged. Output: %s", output)
				}
			},
			randomHash: true,
		},
		{
			name:  "multi-word string should be masked to multi-word string",
			input: `{"key": "mask me"}`,
			asserter: func(t *testing.T, output string) {
				maskedValue := gjson.Get(output, "key").String()
				if maskedValue == "mask me" {
					t.Errorf("Expected string to be masked, but it was unchanged. Output: %s", output)
				}
				if len(strings.Split(maskedValue, " ")) != 2 {
					t.Errorf("Expected masked string to have 2 words, but got %d. Output: %s", len(strings.Split(maskedValue, " ")), output)
				}
				if gjson.Get(output, "key").Type != gjson.String {
					t.Errorf("Expected key to be a string, but got %s", gjson.Get(output, "key").Type)
				}
			},
			randomHash: true,
		},
		{
			name:  "URL should be masked to a URL",
			input: `{"key": "https://example.com/path"}`,
			asserter: func(t *testing.T, output string) {
				maskedValue := gjson.Get(output, "key").String()
				if maskedValue == "https://example.com/path" {
					t.Errorf("Expected URL to be masked, but it was unchanged. Output: %s", output)
				}
				if !strings.Contains(maskedValue, "://") {
					t.Errorf("Expected masked value to be a URL, but it was not. Output: %s", output)
				}
			},
			randomHash: true,
		},
		{
			name:  "capitalization should be mimicked",
			input: `{"key": "Mask Me"}`,
			asserter: func(t *testing.T, output string) {
				maskedValue := gjson.Get(output, "key").String()
				if maskedValue == "Mask Me" {
					t.Errorf("Expected string to be masked, but it was unchanged. Output: %s", output)
				}
				words := strings.Split(maskedValue, " ")
				if len(words) != 2 {
					t.Errorf("Expected masked string to have 2 words, but got %d. Output: %s", len(words), output)
				}
				for _, word := range words {
					if !strings.HasPrefix(word, strings.ToUpper(string(word[0]))) {
						t.Errorf("Expected word to be capitalized, but it wasn't. Word: %s", word)
					}
				}
			},
			randomHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := pkg.NewSaltedMasker(testSalt)

			jsonMasker := pkg.NewJSONMasker(masker)
			input := bytes.NewReader([]byte(tt.input))
			var output bytes.Buffer

			if err := jsonMasker.Mask(input, &output); err != nil {
				t.Fatalf("Mask() error = %v", err)
			}

			tt.asserter(t, output.String())
		})
	}
}
