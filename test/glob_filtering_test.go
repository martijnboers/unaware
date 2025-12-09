package test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"unaware/pkg"
)

// TestGlobFilteringScenarios provides more robust tests for include/exclude logic.
// It checks for the presence or absence of original values rather than specific masked values,
// making it more resilient to changes in the faking library. It also adds CSV support.
func TestGlobFilteringScenarios(t *testing.T) {
	deterministicMaskerConfig := pkg.MaskerConfig{
		Method: pkg.MethodDeterministic,
		Salt:   []byte("static-salt-for-glob-test"),
	}

	jsonInput := map[string]any{
		"user": map[string]any{
			"id": "user-123",
			"personal": map[string]any{
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
		},
		"transaction_id": "txn-abc-456",
	}
	jsonInputBytes, err := json.Marshal(jsonInput)
	require.NoError(t, err)

	deeplyNestedInput := `{
		"company": {
			"departments": [
				{
					"name": "Engineering",
					"teams": [
						{
							"name": "Backend",
							"lead": { "details": { "name": "Charlie", "phone": "111-222-3333" } }
						}
					]
				}
			]
		},
		"metadata": { "audit": { "user": { "name": "AuditBot" } } }
	}`

	csvInput := `id,name,email,transaction_id
user-1,Alice,alice@example.com,txn-1
user-2,Bob,bob@example.com,txn-2`

	testCases := []struct {
		name           string
		format         string
		input          string
		include        []string
		exclude        []string
		shouldBeMasked map[string]string // map of key to original value that SHOULD be masked
		shouldBeKept   map[string]string // map of key to original value that SHOULD remain
	}{
		// --- JSON ---
		{
			name:    "JSON - Include with deep glob",
			format:  "json",
			input:   string(jsonInputBytes),
			include: []string{"user.personal.*"},
			shouldBeMasked: map[string]string{
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
			shouldBeKept: map[string]string{
				"id":             "user-123",
				"transaction_id": "txn-abc-456",
			},
		},
		{
			name:    "JSON - Exclude with glob takes precedence",
			format:  "json",
			input:   string(jsonInputBytes),
			include: []string{"user.**"},
			exclude: []string{"user.id"},
			shouldBeMasked: map[string]string{
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
			shouldBeKept: map[string]string{
				"id":             "user-123",
				"transaction_id": "txn-abc-456",
			},
		},
		{
			name:           "JSON - Wildcard in the middle for nested arrays",
			format:         "json",
			input:          deeplyNestedInput,
			include:        []string{"**.lead.details.phone"},
			shouldBeMasked: map[string]string{"phone": "111-222-3333"},
			shouldBeKept: map[string]string{
				"name":     "Charlie",
				"audit":    "AuditBot",
				"engineer": "Engineering",
			},
		},
		// --- CSV ---
		{
			name:   "CSV - No flags (mask all)",
			format: "csv",
			input:  csvInput,
			shouldBeMasked: map[string]string{
				"id":             "user-1",
				"name":           "Alice",
				"email":          "alice@example.com",
				"transaction_id": "txn-1",
			},
		},
		{
			name:    "CSV - Include single column",
			format:  "csv",
			input:   csvInput,
			include: []string{"email"},
			shouldBeMasked: map[string]string{
				"email": "alice@example.com",
			},
			shouldBeKept: map[string]string{
				"id":             "user-1",
				"name":           "Alice",
				"transaction_id": "txn-1",
			},
		},
		{
			name:    "CSV - Exclude with glob",
			format:  "csv",
			input:   csvInput,
			exclude: []string{"*id"},
			shouldBeMasked: map[string]string{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			shouldBeKept: map[string]string{
				"id":             "user-1",
				"transaction_id": "txn-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appConfig := pkg.AppConfig{
				Format:   tc.format,
				CPUCount: 1,
				Include:  tc.include,
				Exclude:  tc.exclude,
				Masker:   deterministicMaskerConfig,
			}

			var buf bytes.Buffer
			err := pkg.Start(strings.NewReader(tc.input), &buf, appConfig)
			require.NoError(t, err)

			output := buf.String()

			// Check that original values that should be masked are GONE
			for _, val := range tc.shouldBeMasked {
				require.NotContains(t, output, val, "value should have been masked but was found")
			}

			// Check that original values that should be kept are PRESENT
			for _, val := range tc.shouldBeKept {
				require.Contains(t, output, val, "value should have been kept but was not found")
			}

			// Sanity check: ensure output is valid for the format
			switch tc.format {
			case "json":
				require.True(t, json.Valid(buf.Bytes()), "output should be valid JSON")
			case "csv":
				_, err := csv.NewReader(&buf).ReadAll()
				require.NoError(t, err, "output should be valid CSV")
			}
		})
	}
}
