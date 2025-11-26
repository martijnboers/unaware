package test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unaware/pkg"
)

const (
	jsonInputSubset = `[
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
		{"id": 4, "name": "David"}
	]`
	xmlInputSubset = `<?xml version="1.0" encoding="UTF-8"?>
	<users>
		<user><id>1</id><name>Alice</name></user>
		<user><id>2</id><name>Bob</name></user>
		<user><id>3</id><name>Charlie</name></user>
		<user><id>4</id><name>David</name></user>
	</users>`
	csvInputSubset = `id,name
1,Alice
2,Bob
3,Charlie
4,David`
	textInputSubset = `Alice
Bob
Charlie
David`
)

func TestSubsettingFirstN(t *testing.T) {
	testCases := []struct {
		name              string
		format            string
		input             string
		firstN            int
		expectedCount     int
		assertContains    string
		assertNotContains string
	}{
		{
			name:              "JSON First 2",
			format:            "json",
			input:             jsonInputSubset,
			firstN:            2,
			expectedCount:     2,
			assertContains:    `"id":`,
			assertNotContains: `"name": "Charlie"`,
		},
		{
			name:              "JSON First 0 (All)",
			format:            "json",
			input:             jsonInputSubset,
			firstN:            0,
			expectedCount:     4,
			assertContains:    `"id":`,
			assertNotContains: `"name": "Alice"`,
		},
		{
			name:              "XML First 2",
			format:            "xml",
			input:             xmlInputSubset,
			firstN:            2,
			expectedCount:     2,
			assertContains:    `<id>`,
			assertNotContains: `<name>Charlie</name>`,
		},
		{
			name:              "CSV First 2",
			format:            "csv",
			input:             csvInputSubset,
			firstN:            2,
			expectedCount:     2, // Data rows
			assertContains:    "id,name",
			assertNotContains: "3,Charlie",
		},
		{
			name:              "Text First 2",
			format:            "text",
			input:             textInputSubset,
			firstN:            2,
			expectedCount:     2,
			assertContains:    "\n",
			assertNotContains: "Charlie",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader(tc.input)
			var out bytes.Buffer

			appConfig := pkg.AppConfig{
				Format:   tc.format,
				CPUCount: 1,
				FirstN:   tc.firstN,
				Masker: pkg.MaskerConfig{
					Method: pkg.MethodRandom,
				},
			}

			err := pkg.Start(in, &out, appConfig)
			require.NoError(t, err)

			outputStr := out.String()

			if tc.assertContains != "" {
				assert.Contains(t, outputStr, tc.assertContains, "Output should contain expected substring")
			}
			if tc.assertNotContains != "" {
				assert.NotContains(t, outputStr, tc.assertNotContains, "Output should not contain unexpected substring")
			}

			switch tc.format {
			case "json":
				var result []map[string]any
				err := json.Unmarshal(out.Bytes(), &result)
				require.NoError(t, err, "Output should be valid JSON")
				assert.Len(t, result, tc.expectedCount, "JSON array length should match expected count")
			case "xml":
				var xmlParsed any
				require.NoError(t, xml.Unmarshal(out.Bytes(), &xmlParsed), "Output should be valid XML")
				count := strings.Count(outputStr, "<user>")
				assert.Equal(t, tc.expectedCount, count, "XML <user> count should match expected count")
			case "csv":
				r := csv.NewReader(&out)
				records, err := r.ReadAll()
				require.NoError(t, err, "Output should be valid CSV")
				assert.Len(t, records, tc.expectedCount+1, "CSV record count (including header) should match expected")
			case "text":
				lines := strings.Split(strings.TrimSpace(outputStr), "\n")
				assert.Len(t, lines, tc.expectedCount, "Text line count should match expected count")
			}
		})
	}
}
