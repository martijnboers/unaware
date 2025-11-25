package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type processor interface {
	Process(r io.Reader, w io.Writer, cpuCount int) error
}

func Start(format string, cpuCount int, r io.Reader, w io.Writer, strategy MaskingStrategy, include, exclude []string) error {
	var p processor
	switch format {
	case "json":
		p = newJSONProcessor(strategy, include, exclude)
	case "xml":
		p = newXMLProcessor(strategy, include, exclude)
	case "csv":
		p = newCSVProcessor(strategy, include, exclude)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return p.Process(r, w, cpuCount)
}

func shouldMask(key string, include, exclude []string) bool {
	if len(exclude) > 0 {
		for _, pattern := range exclude {
			if matched, _ := filepath.Match(pattern, key); matched {
				return false
			}
		}
	}
	if len(include) > 0 {
		for _, pattern := range include {
			if matched, _ := filepath.Match(pattern, key); matched {
				return true
			}
		}
		return false
	}
	return true
}

type seeder interface {
	SeedFaker(f *gofakeit.Faker, input any)
	SeedFakerForWord(f *gofakeit.Faker, word string)
}

type hashedSeeder struct{ salt []byte }

func (hs *hashedSeeder) SeedFaker(f *gofakeit.Faker, input any) {
	var seedInput string
	switch v := input.(type) {
	case string:
		seedInput = v
	case json.Number:
		seedInput = v.String()
	case bool:
		seedInput = strconv.FormatBool(v)
	default:
		seedInput = "[UNSUPPORTED]"
	}
	f.Rand.Seed(hs.createSeed(seedInput))
}
func (hs *hashedSeeder) SeedFakerForWord(f *gofakeit.Faker, word string) {
	f.Rand.Seed(hs.createSeed(word))
}
func (hs *hashedSeeder) createSeed(s string) int64 {
	mac := hmac.New(sha256.New, hs.salt)
	mac.Write([]byte(s))
	seedBytes := mac.Sum(nil)
	return int64(binary.BigEndian.Uint64(seedBytes))
}

type MaskingStrategy func(*masker)

func Hashed(salt []byte) MaskingStrategy {
	return func(m *masker) {
		m.seeder = &hashedSeeder{salt: salt}
	}
}

type randomSeeder struct{}

func (rs *randomSeeder) SeedFaker(f *gofakeit.Faker, input any)          { /* No-op */ }
func (rs *randomSeeder) SeedFakerForWord(f *gofakeit.Faker, word string) { /* No-op */ }

func Random() MaskingStrategy {
	return func(m *masker) {
		m.seeder = &randomSeeder{}
	}
}

type masker struct {
	faker        *gofakeit.Faker
	seeder       seeder
	dateLayouts  []string
	emailRegex   *regexp.Regexp
	numLikeRegex *regexp.Regexp
}

func newMasker(strategy MaskingStrategy) *masker {
	m := &masker{
		dateLayouts: []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"2006-01",
			"01/02/2006",
			time.RFC1123,
		},
		emailRegex:   regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`),
		numLikeRegex: regexp.MustCompile(`^[\d\s-]+$`),
	}
	strategy(m)
	if _, ok := m.seeder.(*hashedSeeder); ok {
		m.faker = gofakeit.NewUnlocked(1)
	} else {
		m.faker = gofakeit.New(0)
	}
	return m
}

func (m *masker) mask(value any) any {
	if value == nil {
		return nil
	}
	m.seeder.SeedFaker(m.faker, value)
	switch v := value.(type) {
	case string:
		s := v
		if strings.TrimSpace(s) == "" {
			return s
		}
		if _, err := url.ParseRequestURI(s); err == nil {
			return m.faker.URL()
		}
		if m.emailRegex.MatchString(s) {
			return m.faker.Email()
		}
		if _, err := net.ParseMAC(s); err == nil {
			return m.faker.MacAddress()
		}
		if ip := net.ParseIP(s); ip != nil {
			if ip.To4() != nil {
				return m.faker.IPv4Address()
			}
			return m.faker.IPv6Address()
		}
		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			return m.faker.Numerify(strings.Repeat("#", len(s)))
		}
		if _, err := strconv.ParseFloat(s, 64); err == nil {
			parts := strings.Split(s, ".")
			integerPart := parts[0]
			fractionalPart := ""
			if len(parts) > 1 {
				fractionalPart = parts[1]
			}
			template := strings.Repeat("#", len(integerPart))
			if fractionalPart != "" {
				template += "." + strings.Repeat("#", len(fractionalPart))
			}
			return m.faker.Numerify(template)
		}
		for _, layout := range m.dateLayouts {
			if _, err := time.Parse(layout, s); err == nil {
				return m.faker.Date().Format(layout)
			}
		}
		if m.numLikeRegex.MatchString(s) {
			var result strings.Builder
			for _, char := range s {
				if char >= '0' && char <= '9' {
					result.WriteString(strconv.Itoa(m.faker.Rand.Intn(10)))
				} else {
					result.WriteRune(char)
				}
			}
			return result.String()
		}
		if _, err := dateparse.ParseAny(s); err == nil {
			return m.faker.Date().Format(time.RFC3339)
		}
		words := strings.Split(s, " ")
		maskedWords := make([]string, len(words))
		caser := cases.Title(language.English)
		for i, word := range words {
			m.seeder.SeedFakerForWord(m.faker, word)
			maskedWord := m.faker.Word()
			if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
				maskedWords[i] = caser.String(maskedWord)
			} else {
				maskedWords[i] = maskedWord
			}
		}
		return strings.Join(maskedWords, " ")
	case json.Number:
		s := v.String()
		if strings.Contains(s, ".") {
			parts := strings.Split(s, ".")
			template := strings.Repeat("#", len(parts[0])) + "." + strings.Repeat("#", len(parts[1]))
			return json.Number(m.faker.Numerify(template))
		}
		return json.Number(m.faker.Numerify(strings.Repeat("#", len(s))))
	case bool:
		return m.faker.Bool()
	}
	return "[MASKED UNSUPPORTED TYPE]"
}
