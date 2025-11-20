package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"unaware/pkg"
)

func TestJSONMasker_EmptyObject(t *testing.T) {
	inputJSON := `{}`
	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer

	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() with empty object error = %v", err)
	}

	if strings.TrimSpace(out.String()) != "{}" {
		t.Errorf("Expected empty object, got %s", out.String())
	}
}

func TestJSONMasker_EmptyArray(t *testing.T) {
	inputJSON := `[]`
	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer

	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() with empty array error = %v", err)
	}

	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("Expected empty array, got %s", out.String())
	}
}

func TestJSONMasker_NullValues(t *testing.T) {
	inputJSON := `{"key": null}`
	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer

	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() with null values error = %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal masked JSON: %v", err)
	}

	if result["key"] != nil {
		t.Errorf("Expected null value to be preserved, got %v", result["key"])
	}
}

func TestJSONMasker_NestedStructure(t *testing.T) {
	inputJSON := `{"user": {"name": "Jane Doe", "details": {"age": 25, "city": "New York"}}}`
	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer

	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() with nested structure error = %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal masked JSON: %v", err)
	}

	user, ok := result["user"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'user' to be a map")
	}
	if user["name"] == "Jane Doe" {
		t.Error("Expected 'name' to be masked")
	}
	details, ok := user["details"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'details' to be a map")
	}
	if details["age"] == 25 {
		t.Error("Expected 'age' to be masked")
	}
	if details["city"] == "New York" {
		t.Error("Expected 'city' to be masked")
	}
}

func TestJSONMasker_MixedArray(t *testing.T) {
	inputJSON := `["string", 123, true, null]`
	m := pkg.NewSaltedMethod(testSalt)
	jm := pkg.NewJSONProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputJSON)
	var out bytes.Buffer

	err := jm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("JSONMasker.Mask() with mixed array error = %v", err)
	}

	var result []any
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal masked JSON: %v", err)
	}

	if result[0] == "string" {
		t.Error("Expected string in array to be masked")
	}
	if result[1] == 123 {
		t.Error("Expected number in array to be masked")
	}
	if result[3] != nil {
		t.Error("Expected null in array to be preserved")
	}
}
