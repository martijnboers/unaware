package main

import (
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Masker interface {
	Mask(value interface{}) interface{}
}

var consistentSalt = []byte("unaware")

func NewConsistentMasker() Masker {
	return &consistentMasker{}
}

type consistentMasker struct{}

// We keep specific regexes for formats where govalidator is too strict or broad.
var (
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	ssnRegex  = regexp.MustCompile(`^\d{3}-\d{2,3}-\d{4}$`) // Catches XXX-XX-XXXX and XXX-XXX-XXXX
)

func (m *consistentMasker) Mask(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if govalidator.IsCreditCard(v) {
			return maskCreditCard(v)
		}
		if ssnRegex.MatchString(v) {
			return maskSSN(v)
		}
		if govalidator.IsRFC3339(v) {
			return maskTimestamp(v)
		}
		if govalidator.IsEmail(v) {
			return maskEmail(v)
		}
		if govalidator.IsIP(v) {
			return maskIP(v)
		}
		if dateRegex.MatchString(v) {
			return maskDate(v)
		}
		if govalidator.IsInt(v) {
			return maskIntString(v)
		}
		if govalidator.IsFloat(v) {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return fmt.Sprintf("%.2f", maskFloat64(f))
			}
		}
		return maskGenericString(v)
	case float64:
		return maskFloat64(v)
	case bool:
		return maskBool(v)
	default:
		return value
	}
}

func createSeededRand(s string) *rand.Rand {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	hasher.Write(consistentSalt)
	seed := hasher.Sum64()
	return rand.New(rand.NewSource(int64(seed)))
}

func maskCreditCard(s string) string {
	r := createSeededRand(s)
	return fmt.Sprintf("4%03d%04d%04d%04d", r.Intn(1000), r.Intn(10000), r.Intn(10000), r.Intn(10000))
}

func maskSSN(s string) string {
	r := createSeededRand(s)
	return fmt.Sprintf("%03d-%02d-%04d", r.Intn(900)+100, r.Intn(100), r.Intn(10000))
}

func maskTimestamp(s string) string {
	r := createSeededRand(s)
	min := time.Date(1990, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2023, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min
	sec := r.Int63n(delta) + min
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}

func maskDate(s string) string {
	r := createSeededRand(s)
	year := r.Intn(50) + 1970
	month := r.Intn(12) + 1
	day := r.Intn(28) + 1
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func maskEmail(s string) string {
	r := createSeededRand(s)
	user := fakeWords[r.Intn(len(fakeWords))]
	domain := fakeWords[r.Intn(len(fakeWords))]
	tld := fakeTLDs[r.Intn(len(fakeTLDs))]
	return fmt.Sprintf("%s@%s.%s", user, domain, tld)
}

func maskIP(s string) string {
	r := createSeededRand(s)
	return fmt.Sprintf("%d.%d.%d.%d", r.Intn(256), r.Intn(256), r.Intn(256), r.Intn(256))
}

func maskIntString(s string) string {
	r := createSeededRand(s)
	digits := "0123456789"
	var builder strings.Builder
	for i := 0; i < len(s); i++ {
		builder.WriteByte(digits[r.Intn(len(digits))])
	}
	return builder.String()
}

func maskGenericString(s string) string {
	r := createSeededRand(s)
	return fakeWords[r.Intn(len(fakeWords))]
}

func maskFloat64(f float64) float64 {
	s := fmt.Sprintf("%f", f)
	hash := sha256.Sum256(append([]byte(s), consistentSalt...))
	r := rand.New(rand.NewSource(int64(hash[0])<<56 | int64(hash[1])<<48 | int64(hash[2])<<40 | int64(hash[3])<<32 | int64(hash[4])<<24 | int64(hash[5])<<16 | int64(hash[6])<<8 | int64(hash[7])))
	val := r.Float64() * 1000
	return math.Round(val*100) / 100
}

func maskBool(b bool) bool {
	s := strconv.FormatBool(b)
	hash := sha256.Sum256(append([]byte(s), consistentSalt...))
	return hash[0]%2 == 0
}

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

var fakeWords = []string{
	"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf",
	"hotel", "india", "juliett", "kilo", "lima", "mike", "november",
	"oscar", "papa", "quebec", "romeo", "sierra", "tango", "uniform",
	"victor", "whiskey", "xray", "yankee", "zulu", "red", "green", "blue",
	"yellow", "purple", "orange", "silver", "gold", "mercury", "venus",
	"earth", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto",
}

var fakeTLDs = []string{"com", "net", "org", "io", "dev", "co", "xyz"}