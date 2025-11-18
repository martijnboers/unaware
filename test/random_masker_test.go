package test

import (
	"testing"

	"unaware/pkg"
)

func TestRandomMasker(t *testing.T) {
	m := pkg.NewRandomMasker()

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
}
