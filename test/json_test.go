package test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unaware/pkg"
)

// ComplexJSON defines a nested structure for robust JSON processing tests.
type ComplexJSON struct {
	User struct {
		Details struct {
			Name    string `json:"name"`
			Contact struct {
				Email string `json:"email"`
				Phone string `json:"phone"`
			} `json:"contact"`
		} `json:"details"`
		Roles    []string `json:"roles"`
		Metadata struct {
			LastLogin string `json:"last_login"`
			IPAddress string `json:"ip_address"`
		} `json:"metadata"`
	} `json:"user"`
	Data []struct {
		Value any `json:"value"`
	} `json:"data"`
	IsActive bool        `json:"is_active"`
	Score    json.Number `json:"score"`
}

// TestJSONProcessing_Structured validates the masking logic by marshalling and unmarshalling
// a struct, ensuring data is masked while structure and types are preserved.
func TestJSONProcessing_Structured(t *testing.T) {
	salt := []byte("structured-json-salt")

	input := ComplexJSON{
		User: struct {
			Details struct {
				Name    string `json:"name"`
				Contact struct {
					Email string `json:"email"`
					Phone string `json:"phone"`
				} `json:"contact"`
			} `json:"details"`
			Roles    []string `json:"roles"`
			Metadata struct {
				LastLogin string `json:"last_login"`
				IPAddress string `json:"ip_address"`
			} `json:"metadata"`
		}{
			Details: struct {
				Name    string `json:"name"`
				Contact struct {
					Email string `json:"email"`
					Phone string `json:"phone"`
				} `json:"contact"`
			}{
				Name: "Alice Smith",
				Contact: struct {
					Email string `json:"email"`
					Phone string `json:"phone"`
				}{
					Email: "alice.smith@example.com",
					Phone: "123-456-7890",
				},
			},
			Roles: []string{"admin", "editor"},
			Metadata: struct {
				LastLogin string `json:"last_login"`
				IPAddress string `json:"ip_address"`
			}{
				LastLogin: "2023-10-27T10:00:00Z",
				IPAddress: "203.0.113.195",
			},
		},
		Data: []struct {
			Value any `json:"value"`
		}{
			{Value: json.Number("12345")},
			{Value: "sensitive-data-string"},
		},
		IsActive: true,
		Score:    json.Number("98.7"),
	}

	inputBytes, err := json.MarshalIndent(input, "", "  ")
	require.NoError(t, err)

	var buf bytes.Buffer
	err = pkg.Start("json", 1, bytes.NewReader(inputBytes), &buf, pkg.Hashed(salt))
	require.NoError(t, err)

	var output ComplexJSON
	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	decoder.UseNumber()
	err = decoder.Decode(&output)
	require.NoError(t, err, "Output should be valid JSON parsable into the struct. Got: %s", buf.String())

	// Assert that sensitive string values have been changed
	assert.NotEqual(t, input.User.Details.Name, output.User.Details.Name, "Name should be masked")
	assert.NotEqual(t, input.User.Details.Contact.Email, output.User.Details.Contact.Email, "Email should be masked")
	assert.NotEqual(t, input.User.Details.Contact.Phone, output.User.Details.Contact.Phone, "Phone should be masked")
	assert.NotEqual(t, input.User.Metadata.IPAddress, output.User.Metadata.IPAddress, "IPAddress should be masked")
	assert.NotEqual(t, input.User.Metadata.LastLogin, output.User.Metadata.LastLogin, "LastLogin should be masked")

	// Assert array values
	require.Len(t, output.User.Roles, len(input.User.Roles), "Roles array length should be preserved")
	assert.NotEqual(t, input.User.Roles[0], output.User.Roles[0], "Role[0] should be masked")
	assert.NotEqual(t, input.User.Roles[1], output.User.Roles[1], "Role[1] should be masked")

	// Assert mixed-type array values
	require.Len(t, output.Data, len(input.Data), "Data array length should be preserved")
	assert.NotEqual(t, input.Data[0].Value, output.Data[0].Value, "Data[0].Value (number) should be masked")
	assert.NotEqual(t, input.Data[1].Value, output.Data[1].Value, "Data[1].Value (string) should be masked")

	// Assert that types are preserved
	assert.IsType(t, json.Number(""), output.Data[0].Value, "Type of numeric value should be preserved as json.Number")
	assert.IsType(t, "", output.Data[1].Value, "Type of string value should be preserved as string")
	assert.IsType(t, json.Number(""), output.Score, "Type of json.Number should be preserved")
	assert.NotEqual(t, input.Score, output.Score, "Score should be masked")

	// Assert that the original values are not in the output string
	outputString := buf.String()
	assert.NotContains(t, outputString, "Alice Smith")
	assert.NotContains(t, outputString, "alice.smith@example.com")
	assert.NotContains(t, outputString, "123-456-7890")
	assert.NotContains(t, outputString, "sensitive-data-string")
}

func TestJSONStreamingArray(t *testing.T) {
	salt := []byte("streaming-salt")
	input := `[
		{"id": 1, "data": "first"},
		{"id": 2, "data": "second"},
		{"id": 3, "data": "third"}
	]`

	var buf bytes.Buffer
	err := pkg.Start("json", 1, strings.NewReader(input), &buf, pkg.Hashed(salt))
	require.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.HasPrefix(output, "[\n"))
	assert.True(t, strings.HasSuffix(output, "\n]\n"))
	assert.Equal(t, 3, strings.Count(output, `"id":`))
	assert.NotContains(t, output, "first")
	assert.NotContains(t, output, "second")
	assert.NotContains(t, output, "third")
}

func TestEmptyReader(t *testing.T) {
	var buf bytes.Buffer
	err := pkg.Start("json", 1, strings.NewReader(""), &buf, pkg.Random())
	require.NoError(t, err)
	assert.Equal(t, "", buf.String())
}

func TestReaderError(t *testing.T) {
	errorReader := &errorReader{}
	var buf bytes.Buffer
	err := pkg.Start("json", 1, errorReader, &buf, pkg.Random())
	require.Error(t, err)
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
