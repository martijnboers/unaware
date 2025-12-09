package test

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unaware/pkg"
)

func TestIDMaskingAndFormatValidation(t *testing.T) {
	salt := []byte("id-masking-salt")

	// Regexes for validation, mirroring the ones in the engine but for testing purposes.
	ulidRegex := regexp.MustCompile(`(?i)^[0-7][0-9a-hjkmnp-tv-z]{25}$`)
	ksuidRegex := regexp.MustCompile(`^[a-zA-Z0-9]{27}$`)

	testCases := []struct {
		name        string
		input       map[string]string
		validator   func(t *testing.T, outputValue string)
		shouldMatch bool
	}{
		{
			name:  "UUIDv1",
			input: map[string]string{"id": "d9428888-122b-11e1-b85c-619706ab3aae"},
			validator: func(t *testing.T, outputValue string) {
				_, err := uuid.Parse(outputValue)
				assert.NoError(t, err, "Masked UUIDv1 should be a valid UUID")
			},
			shouldMatch: true,
		},
		{
			name:  "UUIDv4",
			input: map[string]string{"id": "a1b2c3d4-e5f6-4890-8234-567890abcdef"},
			validator: func(t *testing.T, outputValue string) {
				_, err := uuid.Parse(outputValue)
				assert.NoError(t, err, "Masked UUIDv4 should be a valid UUID")
			},
			shouldMatch: true,
		},
		{
			name:  "UUIDv5",
			input: map[string]string{"id": "98765432-1234-5678-9012-abcdef123456"},
			validator: func(t *testing.T, outputValue string) {
				_, err := uuid.Parse(outputValue)
				assert.NoError(t, err, "Masked UUIDv5 should be a valid UUID")
			},
			shouldMatch: true,
		},
		{
			name:  "ULID",
			input: map[string]string{"id": "01F8X0J5Z5J5Z5J5Z5J5Z5J5Z5"},
			validator: func(t *testing.T, outputValue string) {
				assert.Regexp(t, ulidRegex, outputValue, "Masked ULID should have a valid ULID format")
			},
			shouldMatch: true,
		},
		{
			name:  "KSUID",
			input: map[string]string{"id": "0o1g2h3j4k5l6m7n8p9q0r1s2t3"},
			validator: func(t *testing.T, outputValue string) {
				assert.Regexp(t, ksuidRegex, outputValue, "Masked KSUID should have a valid KSUID format")
			},
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tc.input)
			require.NoError(t, err)

			appConfig := pkg.AppConfig{
				Format:   "json",
				CPUCount: 1,
				Masker: pkg.MaskerConfig{
					Method: pkg.MethodDeterministic,
					Salt:   salt,
				},
			}

			var buf bytes.Buffer
			err = pkg.Start(bytes.NewReader(inputBytes), &buf, appConfig)
			require.NoError(t, err)

			var output map[string]string
			err = json.Unmarshal(buf.Bytes(), &output)
			require.NoError(t, err, "Output should be valid JSON. Got: %s", buf.String())

			originalValue := tc.input["id"]
			maskedValue := output["id"]

			assert.NotEqual(t, originalValue, maskedValue, "Value should be masked")
			assert.NotContains(t, buf.String(), originalValue, "Original value should not be in the output")

			if tc.validator != nil {
				tc.validator(t, maskedValue)
			}
		})
	}
}
