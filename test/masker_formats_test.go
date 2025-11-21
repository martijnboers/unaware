package test

import (
	"testing"

	"unaware/pkg"
)

func TestEmptyAndWhitespaceMasking(t *testing.T) {
	m := pkg.NewHashedMethod(testSalt)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Whitespace string",
			input: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked := m.Mask(tt.input).(string)
			if masked != tt.input {
				t.Errorf("Expected string to be preserved, but it was changed. Got: '%s'", masked)
			}
		})
	}
}
