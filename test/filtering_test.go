package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"unaware/pkg"
)

func TestFilteringScenarios(t *testing.T) {
	strategy := pkg.Deterministic([]byte("static-salt"))

	jsonInput := `{
	"user": {
		"id": "user-123",
		"personal": {
			"name": "John Doe",
			"email": "john.doe@example.com"
		},
		"metadata": {
			"last_login": "2023-10-27T10:00:00Z",
			"ip_address": "203.0.113.195"
		}
	},
	"transaction_id": "txn-abc-456"
}`

	xmlInput := `<root>
	<user id="user-xyz">
		<personal>
			<name>Jane Doe</name>
			<email>jane.doe@example.com</email>
		</personal>
		<metadata>
			<last_login>2023-11-01T12:30:00Z</last_login>
			<ip_address>198.51.100.22</ip_address>
		</metadata>
	</user>
	<transaction_id>txn-def-789</transaction_id>
</root>`

	testCases := []struct {
		name        string
		format      string
		input       string
		include     []string
		exclude     []string
		expected    []string // Substrings to check for in the output
		notExpected []string // Substrings that should NOT be in the output
	}{
		// JSON Scenarios
		{
			name:   "JSON - No Flags (Mask All)",
			format: "json",
			input:  jsonInput,
			expected: []string{
				`"id": "chapter"`,
				`"name": "Cute His"`,
				`"email": "clevelandsenger@sawayn.io"`,
				`"last_login": "1945-09-30T19:45:32Z"`,
				`"ip_address": "113.186.161.194"`,
				`"transaction_id": "without"`,
			},
			notExpected: []string{
				`"id": "user-123"`,
				`"name": "John Doe"`,
			},
		},
		{
			name:    "JSON - Exclude Only (Blacklist)",
			format:  "json",
			input:   jsonInput,
			exclude: []string{"*.id", "*.ip_address"},
			expected: []string{
				`"id": "user-123"`,                 // Not masked
				`"ip_address": "203.0.113.195"`,     // Not masked
				`"name": "Cute His"`,               // Masked
				`"transaction_id": "without"`, // Masked
			},
		},
		{
			name:    "JSON - Include Only (Whitelist)",
			format:  "json",
			input:   jsonInput,
			include: []string{"user.personal.*"},
			expected: []string{
				`"id": "user-123"`,                 // Not masked (not in include)
				`"name": "Cute His"`,               // Masked
				`"email": "clevelandsenger@sawayn.io"`, // Masked
				`"ip_address": "203.0.113.195"`,     // Not masked
			},
		},
		{
			name:    "JSON - Combined Include and Exclude",
			format:  "json",
			input:   jsonInput,
			include: []string{"user.*", "user.personal.*", "user.metadata.*"},
			exclude: []string{"user.id", "user.metadata.last_login"},
			expected: []string{
				`"id": "user-123"`,                     // Not masked (excluded)
				`"name": "Cute His"`,                   // Masked (included)
				`"last_login": "2023-10-27T10:00:00Z"`, // Not masked (excluded)
				`"transaction_id": "txn-abc-456"`,     // Not masked (not included)
			},
		},
		// XML Scenarios
		{
			name:   "XML - No Flags (Mask All)",
			format: "xml",
			input:  xmlInput,
			expected: []string{
				`<user id="constantly">`,
				`<name>Finally His</name>`,
				`<email>alfredofritsch@dickinson.net</email>`,
				`<last_login>2001-03-07T14:59:55Z</last_login>`,
				`<ip_address>238.108.102.226</ip_address>`,
				`<transaction_id>her</transaction_id>`,
			},
		},
		{
			name:    "XML - Exclude Attributes and Elements",
			format:  "xml",
			input:   xmlInput,
			exclude: []string{"root.user.id", "root.user.personal.name"},
			expected: []string{
				`<user id="user-xyz">`, // Not masked
				`<name>Jane Doe</name>`,   // Not masked
				`<email>alfredofritsch@dickinson.net</email>`, // Masked
			},
		},
		{
			name:    "XML - Include with Wildcard",
			format:  "xml",
			input:   xmlInput,
			include: []string{"root.user.metadata.*"},
			expected: []string{
				`<name>Jane Doe</name>`,                         // Not masked (not included)
				`<last_login>2001-03-07T14:59:55Z</last_login>`, // Masked
				`<ip_address>238.108.102.226</ip_address>`,     // Masked
			},
		},
		{
			name:    "XML - Combined Include and Exclude",
			format:  "xml",
			input:   xmlInput,
			include: []string{"root.user.*", "root.user.personal.*", "root.user.metadata.*"},
			exclude: []string{"root.user.id", "root.user.metadata.ip_address"},
			expected: []string{
				`<user id="user-xyz">`,                     // Not masked (excluded)
				`<name>Finally His</name>`,                  // Masked (included)
				`<ip_address>198.51.100.22</ip_address>`,   // Not masked (excluded)
				`<transaction_id>txn-def-789</transaction_id>`, // Not masked (not included)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := pkg.Start(tc.format, 1, strings.NewReader(tc.input), &buf, strategy, tc.include, tc.exclude)
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tc.expected {
				require.Contains(t, output, expected, "Output should contain expected substring")
			}
			for _, notExpected := range tc.notExpected {
				require.NotContains(t, output, notExpected, "Output should not contain unexpected substring")
			}
		})
	}
}
