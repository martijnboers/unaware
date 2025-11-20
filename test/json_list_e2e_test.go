package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"unaware/pkg"
)

const jsonListTestData = `[
  {
    "name": "France",
    "capital": "Paris",
    "population": 67364357,
    "area": 551695,
    "currency": "Euro",
    "languages": [
      "French"
    ],
    "region": "Europe",
    "subregion": "Western Europe",
    "flag": "https://upload.wikimedia.org/wikipedia/commons/c/c3/Flag_of_France.svg"
  },
  {
    "name": "Germany",
    "capital": "Berlin",
    "population": 83240525,
    "area": 357022,
    "currency": "Euro",
    "languages": [
      "German"
    ],
    "region": "Europe",
    "subregion": "Western Europe",
    "flag": "https://upload.wikimedia.org/wikipedia/commons/b/ba/Flag_of_Germany.svg"
  }
]`

func TestJSONMasker_ListEndToEnd(t *testing.T) {
	masker := pkg.NewJSONProcessor(pkg.NewSaltedMethod(testSalt))
	var buf bytes.Buffer
	err := masker.Mask(strings.NewReader(jsonListTestData), &buf)
	if err != nil {
		t.Fatalf("Failed to mask JSON list: %v", err)
	}
	maskedJSON := buf.Bytes()

	var result []map[string]any
	if err := json.Unmarshal(maskedJSON, &result); err != nil {
		t.Fatalf("Failed to unmarshal masked JSON list: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 items in the list, but got %d", len(result))
	}

	for _, item := range result {
		if item["name"] == "France" || item["name"] == "Germany" {
			t.Errorf("Name was not masked: %s", item["name"])
		}
		if item["capital"] == "Paris" || item["capital"] == "Berlin" {
			t.Errorf("Capital was not masked: %s", item["capital"])
		}
	}
}
