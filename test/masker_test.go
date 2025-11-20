package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"unaware/pkg"
)

func TestConsistentMasker(t *testing.T) {
	m := pkg.NewSaltedMethod(testSalt)

	s1 := "hello"
	s2 := "world"
	maskedS1 := m.Mask(s1)
	maskedS2 := m.Mask(s2)

	if maskedS1 == s1 || maskedS2 == s2 {
		t.Error("String masking should change the value")
	}
	if maskedS1 != m.Mask(s1) {
		t.Error("Consistent string masking should produce the same result for the same input")
	}
	if maskedS1 == maskedS2 {
		t.Error("Consistent string masking should produce different results for different inputs")
	}

	f1 := json.Number("123.456")
	f2 := json.Number("789.012")
	maskedF1 := m.Mask(f1)
	maskedF2 := m.Mask(f2)

	if maskedF1 == f1 || maskedF2 == f2 {
		t.Error("Float64 masking should change the value")
	}
	if maskedF1 != m.Mask(f1) {
		t.Error("Consistent float64 masking should produce the same result for the same input")
	}
	if maskedF1 == maskedF2 {
		t.Error("Consistent float64 masking should produce different results for different inputs")
	}
}

func TestMaskingAvalancheEffect(t *testing.T) {
	m := pkg.NewSaltedMethod(testSalt)

	input1 := "hello world"
	input2 := "hello worle" // Single character difference
	masked1 := m.Mask(input1)
	masked2 := m.Mask(input2)

	if masked1 == masked2 {
		t.Error("A small change in input should produce a different masked output, but it did not.")
	}
}

func TestUniquenessForDifferentInputs(t *testing.T) {
	m := pkg.NewSaltedMethod(testSalt)

	inputs := []string{"ACTIVE", "INACTIVE", "PENDING", "DELETED", "ARCHIVED"}
	outputs := make(map[string]bool)
	for _, input := range inputs {
		masked := m.Mask(input).(string)
		if outputs[masked] {
			t.Errorf("Collision detected: Masked value '%s' is not unique for input '%s'", masked, input)
		}
		outputs[masked] = true
	}
}

func TestJSONMasker(t *testing.T) {
	inputJSON := `{"name": "John Doe", "age": 30, "isStudent": false, "courses": ["Math", "Science"]}`

	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer
	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() error = %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal masked JSON: %v", err)
	}

	if result["name"] == "John Doe" {
		t.Error("Expected 'name' to be masked")
	}
	if result["age"] == 30.0 {
		t.Error("Expected 'age' to be masked")
	}
}

func TestXMLMasker(t *testing.T) {
	inputXML := `<person><name>John Doe</name><age>30</age></person>`
	m := pkg.NewSaltedMethod(testSalt)
	xm := pkg.NewXMLProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer
	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "<name>") {
		t.Error("Expected output to contain <name> tag")
	}
	if strings.Contains(output, ">John Doe<") {
		t.Error("Expected 'John Doe' to be masked")
	}
	if strings.Contains(output, ">30<") {
		t.Error("Expected '30' to be masked")
	}
}
