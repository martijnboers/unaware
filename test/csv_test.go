package test

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unaware/pkg"
)

func TestCSVProcessing_Deterministic(t *testing.T) {
	salt := []byte("csv-salt")
	input := `id,name,email,ip_address,notes
1,John Doe,john.doe@example.com,192.168.1.1,some notes
2,Jane Smith,jane.smith@example.net,10.0.0.2,more data`

	appConfig := pkg.AppConfig{
		Format:   "csv",
		CPUCount: 2,
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodDeterministic,
			Salt:   salt,
		},
	}

	var buf bytes.Buffer
	err := pkg.Start(strings.NewReader(input), &buf, appConfig)
	require.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)

	// Verify header is preserved
	assert.True(t, strings.HasPrefix(output, "id,name,email,ip_address,notes\n"))

	// Verify original sensitive data is gone
	assert.NotContains(t, output, "John Doe")
	assert.NotContains(t, output, "john.doe@example.com")
	assert.NotContains(t, output, "192.168.1.1")
	assert.NotContains(t, output, "Jane Smith")
	assert.NotContains(t, output, "jane.smith@example.net")
	assert.NotContains(t, output, "10.0.0.2")

	// Verify the structure is a valid CSV with the correct number of records
	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	require.NoError(t, err)
	assert.Len(t, records, 3) // Header + 2 rows
	assert.Len(t, records[1], 5)
	assert.Len(t, records[2], 5)
}

func TestCSVProcessing_WithInclude(t *testing.T) {
	salt := []byte("csv-include-salt")
	input := `id,name,email,ip_address,notes
1,John Doe,john.doe@example.com,192.168.1.1,some notes
2,Jane Smith,jane.smith@example.net,10.0.0.2,more data`

	appConfig := pkg.AppConfig{
		Format:   "csv",
		CPUCount: 2,
		Include:  []string{"email", "ip_address"},
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodDeterministic,
			Salt:   salt,
		},
	}

	var buf bytes.Buffer
	err := pkg.Start(strings.NewReader(input), &buf, appConfig)
	require.NoError(t, err)

	output := buf.String()

	// Verify sensitive data that should be masked is gone
	assert.NotContains(t, output, "john.doe@example.com")
	assert.NotContains(t, output, "192.168.1.1")

	// Verify data that should be preserved is still there
	assert.Contains(t, output, "John Doe")
	assert.Contains(t, output, "Jane Smith")
	assert.Contains(t, output, "some notes")

	// Verify structure and read the output to check specific fields
	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3)

	// Check first data row
	assert.Equal(t, "1", records[1][0])
	assert.Equal(t, "John Doe", records[1][1])
	assert.NotEqual(t, "john.doe@example.com", records[1][2]) // Masked
	assert.NotEqual(t, "192.168.1.1", records[1][3])          // Masked
	assert.Equal(t, "some notes", records[1][4])
}

func TestCSVProcessing_WithExclude(t *testing.T) {
	salt := []byte("csv-exclude-salt")
	input := `id,name,email,ip_address,notes
1,John Doe,john.doe@example.com,192.168.1.1,some notes`

	appConfig := pkg.AppConfig{
		Format:   "csv",
		CPUCount: 2,
		Exclude:  []string{"id", "notes"},
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodDeterministic,
			Salt:   salt,
		},
	}

	var buf bytes.Buffer
	err := pkg.Start(strings.NewReader(input), &buf, appConfig)
	require.NoError(t, err)

	output := buf.String()

	// Verify sensitive data that should be masked is gone
	assert.NotContains(t, output, "John Doe")
	assert.NotContains(t, output, "john.doe@example.com")

	// Verify data that should be preserved is still there
	assert.Contains(t, output, "1,")          // id is preserved
	assert.Contains(t, output, ",some notes") // notes is preserved

	// Verify structure
	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "1", records[1][0])                       // Preserved
	assert.NotEqual(t, "John Doe", records[1][1])             // Masked
	assert.NotEqual(t, "john.doe@example.com", records[1][2]) // Masked
	assert.NotEqual(t, "192.168.1.1", records[1][3])          // Masked
	assert.Equal(t, "some notes", records[1][4])              // Preserved
}

func TestEmptyReader_CSV(t *testing.T) {
	appConfig := pkg.AppConfig{
		Format:   "csv",
		CPUCount: 1,
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodRandom,
		},
	}

	var buf bytes.Buffer
	err := pkg.Start(strings.NewReader(""), &buf, appConfig)
	require.NoError(t, err)
	assert.Equal(t, "", buf.String())
}

func TestReaderError_CSV(t *testing.T) {
	errorReader := &errorReader{}
	appConfig := pkg.AppConfig{
		Format:   "csv",
		CPUCount: 1,
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodRandom,
		},
	}
	var buf bytes.Buffer
	err := pkg.Start(errorReader, &buf, appConfig)
	require.Error(t, err)
}
