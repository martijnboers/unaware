package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type seeder interface {
	SeedFaker(f *gofakeit.Faker, input any)
	SeedFakerForWord(f *gofakeit.Faker, word string)
}

type hashedSeeder struct {
	salt []byte
}

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

type randomSeeder struct{}

func (rs *randomSeeder) SeedFaker(f *gofakeit.Faker, input any)          { /* No-op */ }
func (rs *randomSeeder) SeedFakerForWord(f *gofakeit.Faker, word string) { /* No-op */ }

type masker struct {
	faker        *gofakeit.Faker
	seeder       seeder
	dateLayouts  []string
	emailRegex   *regexp.Regexp
	numLikeRegex *regexp.Regexp
}

type Method interface {
	Mask(value any) any
}

func newMasker(f *gofakeit.Faker, s seeder) *masker {
	return &masker{
		faker:  f,
		seeder: s,
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
}

func NewHashedMethod(salt []byte) Method {
	return newMasker(gofakeit.NewUnlocked(1), &hashedSeeder{salt: salt})
}

func NewRandomMethod() Method {
	return newMasker(gofakeit.New(0), &randomSeeder{})
}

func (m *masker) Mask(value any) any {
	if value == nil {
		return nil
	}

	m.seeder.SeedFaker(m.faker, value)

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return v
		}

		// 1. Unambiguous, non-numeric formats first.
		if _, err := url.ParseRequestURI(v); err == nil {
			return m.faker.URL()
		}
		if m.emailRegex.MatchString(v) {
			return m.faker.Email()
		}
		if _, err := net.ParseMAC(v); err == nil {
			return m.faker.MacAddress()
		}
		if ip := net.ParseIP(v); ip != nil {
			if ip.To4() != nil {
				return m.faker.IPv4Address()
			}
			return m.faker.IPv6Address()
		}

		// 2. Pure numeric strings.
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			return m.faker.Numerify(strings.Repeat("#", len(v)))
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			parts := strings.Split(v, ".")
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

		// 3. Specific date layouts.
		for _, layout := range m.dateLayouts {
			if _, err := time.Parse(layout, v); err == nil {
				return m.faker.Date().Format(layout)
			}
		}

		// 4. Mixed-character numeric strings.
		if m.numLikeRegex.MatchString(v) {
			var result strings.Builder
			for _, char := range v {
				if char >= '0' && char <= '9' {
					result.WriteString(strconv.Itoa(m.faker.Rand.Intn(10)))
				} else {
					result.WriteRune(char)
				}
			}
			return result.String()
		}

		// 5. Greedy date parser.
		if _, err := dateparse.ParseAny(v); err == nil {
			return m.faker.Date().Format(time.RFC3339)
		}

		// 6. No specific rule matched.
		words := strings.Split(v, " ")
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
