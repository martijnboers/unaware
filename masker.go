package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// Masker is the interface that defines the masking behavior.
type Masker interface {
	Mask(value interface{}) interface{}
}

// consistentSalt is used to make the hashing consistent.
var consistentSalt = []byte("unaware")

// NewConsistentMasker creates a new masker that consistently masks values.
func NewConsistentMasker() Masker {
	return &consistentMasker{}
}

type consistentMasker struct{}

func (m *consistentMasker) Mask(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		// To ensure consistency between formats (e.g. XML vs JSON),
		// we try to parse the string as a more specific type.
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return maskFloat64(f)
		}
		if b, err := strconv.ParseBool(v); err == nil {
			return maskBool(b)
		}
		return maskString(v)
	case float64:
		return maskFloat64(v)
	case bool:
		return maskBool(v)
	default:
		return value
	}
}

func maskString(s string) string {
	hash := sha256.Sum256(append([]byte(s), consistentSalt...))
	return hex.EncodeToString(hash[:])
}

func maskFloat64(f float64) float64 {
	s := fmt.Sprintf("%f", f)
	hash := sha256.Sum256(append([]byte(s), consistentSalt...))
	// Use the first 8 bytes of the hash to create a new float.
	// This is not guaranteed to be in the same range, but it's a simple way to get a consistent random float.
	// A more sophisticated approach might be needed for specific requirements.
	r := rand.New(rand.NewSource(int64(hash[0])<<56 | int64(hash[1])<<48 | int64(hash[2])<<40 | int64(hash[3])<<32 | int64(hash[4])<<24 | int64(hash[5])<<16 | int64(hash[6])<<8 | int64(hash[7])))
	return r.Float64() * 1000 // Scale it to a reasonable number
}

func maskBool(b bool) bool {
	// Simple approach: flip the boolean based on a hash.
	s := strconv.FormatBool(b)
	hash := sha256.Sum256(append([]byte(s), consistentSalt...))
	return hash[0]%2 == 0
}

// NewRandomMasker creates a new masker that randomly masks values.
func NewRandomMasker() Masker {
	return &randomMasker{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type randomMasker struct {
	rand *rand.Rand
}

func (m *randomMasker) Mask(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch value.(type) {
	case string:
		return m.randomString(10)
	case float64:
		return m.rand.Float64() * 1000
	case bool:
		return m.rand.Intn(2) == 0
	default:
		return value
	}
}

func (m *randomMasker) randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01256789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[m.rand.Intn(len(charset))]
	}
	return string(b)
}
