package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestConsistentMasker(t *testing.T) {
	m := NewConsistentMasker()

	// Test string masking
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

	// Test float64 masking
	f1 := 123.456
	f2 := 789.012
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

	// Test bool masking
	b1 := true
	b2 := false
	maskedB1 := m.Mask(b1)
	maskedB2 := m.Mask(b2)

	// This is a weak test, as it could randomly pass. A better test would be to check the distribution.
	if maskedB1 == b1 && maskedB2 == b2 {
		t.Log("Bool masking might not have changed the values, which is possible but unlikely for both")
	}
}

func TestRandomMasker(t *testing.T) {
	m := NewRandomMasker()

	// Test string masking
	s1 := "hello"
	maskedS1a := m.Mask(s1)
	maskedS1b := m.Mask(s1)

	if maskedS1a == s1 {
		t.Error("String masking should change the value")
	}
	if maskedS1a == maskedS1b {
		t.Error("Random string masking should produce different results for the same input")
	}

	// Test float64 masking
	f1 := 123.456
	maskedF1a := m.Mask(f1)
	maskedF1b := m.Mask(f1)

	if maskedF1a == f1 {
		t.Error("Float64 masking should change the value")
	}
	if maskedF1a == maskedF1b {
		t.Error("Random float64 masking should produce different results for the same input")
	}

	// Test bool masking
	// This is a weak test, as it could randomly pass.
	b1 := true
	maskedB1a := m.Mask(b1)
	maskedB1b := m.Mask(b1)
	if maskedB1a == b1 && maskedB1b == b1 {
		t.Log("Random bool masking might not have changed the value, which is possible")
	}
}

func TestJSONMasker(t *testing.T) {
	inputJSON := `{"name": "John Doe", "age": 30, "isStudent": false, "courses": ["Math", "Science"]}`

	m := NewConsistentMasker()
	jm := NewJSONMasker(m)

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

	// Check if the structure is preserved
	if _, ok := result["name"].(string); !ok {
		t.Error("Expected 'name' to be a string")
	}
	if _, ok := result["age"].(float64); !ok {
		t.Error("Expected 'age' to be a number")
	}
	if _, ok := result["isStudent"].(bool); !ok {
		t.Error("Expected 'isStudent' to be a boolean")
	}
	if courses, ok := result["courses"].([]interface{}); !ok || len(courses) != 2 {
		t.Error("Expected 'courses' to be an array of two items")
	}

	// Check if values are masked
	if result["name"] == "John Doe" {
		t.Error("Expected 'name' to be masked")
	}
	if result["age"] == 30 {
		t.Error("Expected 'age' to be masked")
	}
}

func TestXMLMasker(t *testing.T) {
	inputXML := `<person><name>John Doe</name><age>30</age></person>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

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
