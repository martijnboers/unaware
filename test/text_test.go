package test

import (
	"bytes"
	"strings"
	"testing"
	"unaware/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextProcessor(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Single line with email",
			input: "this line has an email: test@example.com",
		},
		{
			name:  "Multi-line input",
			input: "hello world\n127.0.0.1\n2024-01-01",
		},
		{
			name:  "Empty input",
			input: "",
		},
		{
			name: "Line with only whitespace",
			input: "   \t   \n ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := pkg.Start("text", 1, strings.NewReader(tc.input), &buf, pkg.Random(), nil, nil, 0)
			require.NoError(t, err)

			output := buf.String()

			if tc.input == "" {
				assert.Equal(t, "", output)
				return
			}

			// The output should not be the same as the input (unless it's only whitespace)
			if strings.TrimSpace(tc.input) != "" {
				assert.NotEqual(t, tc.input, output, "Masked output should not be the same as the input")
			}

			// The number of lines should be preserved
			inputLines := strings.Split(tc.input, "\n")
			outputLines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
			assert.Len(t, outputLines, len(inputLines), "Number of lines should be preserved")

		})
	}
}
