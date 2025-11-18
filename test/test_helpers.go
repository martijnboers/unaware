package test

import (
	"regexp"
	"testing"
)

var testSalt = []byte("test-salt")

func validateField(t *testing.T, value, pattern, fieldName string) {
	t.Helper()
	re := regexp.MustCompile(pattern)
	if !re.MatchString(value) {
		t.Errorf("Validation failed for %s. Got: %s, Expected pattern: %s", fieldName, value, pattern)
	}
}

func validateFloatField(t *testing.T, value any, fieldName string) {
	t.Helper()
	if _, ok := value.(float64); !ok {
		t.Errorf("Validation failed for %s. Expected a float64, but got %T", fieldName, value)
	}
}
