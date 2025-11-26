package test

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"unaware/pkg"
)

// ComplexXML defines a nested structure for robust XML processing tests.
type ComplexXML struct {
	XMLName  xml.Name `xml:"root"`
	Metadata struct {
		Timestamp string `xml:"timestamp"`
	} `xml:"metadata"`
	DataItems []Item `xml:"dataitems>item"`
}

// Item represents a nested element within the ComplexXML structure.
type Item struct {
	Key     string `xml:"key,attr"`
	Value   string `xml:"value"`
	Details struct {
		IP  string `xml:"ip"`
		MAC string `xml:"mac"`
	} `xml:"details"`
}

// TestXMLProcessing_Structured validates the masking logic by marshalling and unmarshalling
// a struct, ensuring data is masked while structure and types are preserved.
func TestXMLProcessing_Structured(t *testing.T) {
	salt := []byte("structured-xml-salt")

	input := ComplexXML{
		Metadata: struct {
			Timestamp string `xml:"timestamp"`
		}{Timestamp: "2023-10-27T12:00:00Z"},
		DataItems: []Item{
			{
				Key:   "A1",
				Value: "Some sensitive data",
				Details: struct {
					IP  string `xml:"ip"`
					MAC string `xml:"mac"`
				}{IP: "203.0.113.42", MAC: "00:1B:44:11:3A:B7"},
			},
			{
				Key:   "B2",
				Value: "More private info",
				Details: struct {
					IP  string `xml:"ip"`
					MAC string `xml:"mac"`
				}{IP: "198.51.100.8", MAC: "00:1B:44:11:3A:B8"},
			},
		},
	}

	inputBytes, err := xml.MarshalIndent(input, "", "  ")
	require.NoError(t, err)
	inputWithHeader := append([]byte(xml.Header), inputBytes...)

	appConfig := pkg.AppConfig{
		Format:   "xml",
		CPUCount: 1,
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodDeterministic,
			Salt:   salt,
		},
	}

	var buf bytes.Buffer
	err = pkg.Start(bytes.NewReader(inputWithHeader), &buf, appConfig)
	require.NoError(t, err)

	var output ComplexXML
	err = xml.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err, "Output should be valid XML parsable into the struct. Got: %s", buf.String())

	assert.NotEqual(t, input.Metadata.Timestamp, output.Metadata.Timestamp, "Timestamp should be masked")
	require.Len(t, output.DataItems, len(input.DataItems), "DataItems length should be preserved")

	assert.NotEqual(t, input.DataItems[0].Key, output.DataItems[0].Key, "Item 1 Key should be masked")
	assert.NotEqual(t, input.DataItems[0].Value, output.DataItems[0].Value, "Item 1 Value should be masked")
	assert.NotEqual(t, input.DataItems[0].Details.IP, output.DataItems[0].Details.IP, "Item 1 IP should be masked")
	assert.NotEqual(t, input.DataItems[0].Details.MAC, output.DataItems[0].Details.MAC, "Item 1 MAC should be masked")

	assert.NotEqual(t, input.DataItems[1].Key, output.DataItems[1].Key, "Item 2 Key should be masked")
	assert.NotEqual(t, input.DataItems[1].Value, output.DataItems[1].Value, "Item 2 Value should be masked")
	assert.NotEqual(t, input.DataItems[1].Details.IP, output.DataItems[1].Details.IP, "Item 2 IP should be masked")
	assert.NotEqual(t, input.DataItems[1].Details.MAC, output.DataItems[1].Details.MAC, "Item 2 MAC should be masked")

	outputString := buf.String()
	assert.NotContains(t, outputString, "Some sensitive data")
	assert.NotContains(t, outputString, "More private info")
	assert.NotContains(t, outputString, "203.0.113.42")
}

func TestXMLStreaming(t *testing.T) {
	salt := []byte("xml-streaming-salt")
	input := `<items>
				<item id="1">data1</item>
				<item id="2">data2</item>
				<item id="3">data3</item>
			</items>`

	appConfig := pkg.AppConfig{
		Format:   "xml",
		CPUCount: 1,
		Masker: pkg.MaskerConfig{
			Method: pkg.MethodDeterministic,
			Salt:   salt,
		},
	}

	var buf bytes.Buffer
	err := pkg.Start(strings.NewReader(input), &buf, appConfig)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<items>")
	assert.Contains(t, output, "</items>")
	assert.Equal(t, 3, strings.Count(output, "<item id="))
	assert.NotContains(t, output, "data1")
	assert.NotContains(t, output, "data2")
	assert.NotContains(t, output, "data3")
}
