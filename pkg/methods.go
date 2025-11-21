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

	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Method interface {
	Mask(value any) any
}

func NewHashedMethod(salt []byte) Method {
	return &HashMethod{
		salt:  salt,
		faker: gofakeit.NewUnlocked(1),
	}
}

type HashMethod struct {
	salt  []byte
	faker *gofakeit.Faker
}

func NewRandomMethod() Method {
	return &RandomMethod{
		faker: gofakeit.New(0),
	}
}

type RandomMethod struct {
	faker *gofakeit.Faker
}

var (
	emailRegex      = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	numberLikeRegex = regexp.MustCompile(`^[\d\s-]+$`)
)

func (m *HashMethod) Mask(value any) any {
	if value == nil {
		return nil
	}

	var seedInput string
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return v
		}
		seedInput = v
	case json.Number:
		seedInput = v.String()
	case bool:
		seedInput = strconv.FormatBool(v)
	default:
		return "[MASKED UNSUPPORTED TYPE]"
	}

	// Re-seed the main faker for deterministic logic.
	seed := m.createSeed(seedInput)
	m.faker.Rand.Seed(seed)

	switch v := value.(type) {
	case string:
		// Date and Time parsing should be before number-like checks
		layouts := []string{time.RFC3339, "2006-01-02", "2006-01", "01/02/2006"}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				year := m.faker.IntRange(2000, 2039)
				month := time.Month(m.faker.IntRange(1, 12))
				day := m.faker.IntRange(1, 28) // Day between 1-28 to be safe for all months

				hour, min, sec := t.Clock()
				nsec := t.Nanosecond()
				loc := t.Location()

				newDate := time.Date(year, month, day, hour, min, sec, nsec, loc)
				return newDate.Format(layout)
			}
		}

		if numberLikeRegex.MatchString(v) {
			return m.maskStructuredString(v)
		}
		if _, err := url.ParseRequestURI(v); err == nil {
			return m.maskURL()
		}
		if emailRegex.MatchString(v) {
			return m.maskEmail()
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

		// Generic number checking should be last before word masking
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

		words := strings.Split(v, " ")
		maskedWords := make([]string, len(words))
		caser := cases.Title(language.English)
		for i, word := range words {
			// Re-seed the main faker for each word for deterministic word-level masking.
			m.faker.Rand.Seed(m.createSeed(word))
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
			integerPart := parts[0]
			fractionalPart := parts[1]
			template := strings.Repeat("#", len(integerPart)) + "." + strings.Repeat("#", len(fractionalPart))
			return json.Number(m.faker.Numerify(template))
		}
		return json.Number(m.faker.Numerify(strings.Repeat("#", len(s))))

	case bool:
		return m.faker.Bool()
	}

	return "[MASKED]"
}

func (m *HashMethod) maskURL() string {
	domain := m.randomString(10)
	path1 := m.randomString(4)
	path2 := m.randomString(4)
	return "https://www." + domain + ".local/" + path1 + "/" + path2
}

func (m *HashMethod) maskEmail() string {
	user := m.randomString(10)
	domain := m.randomString(10)
	return user + "@" + domain + ".local"
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func (m *HashMethod) randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[m.faker.Rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (m *HashMethod) maskStructuredString(s string) string {
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

func (m *HashMethod) createSeed(s string) int64 {
	mac := hmac.New(sha256.New, m.salt)
	mac.Write([]byte(s))
	seedBytes := mac.Sum(nil)
	return int64(binary.BigEndian.Uint64(seedBytes))
}

func (r *RandomMethod) Mask(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return v // Preserve empty and whitespace-only strings.
		}
		// Date and Time parsing should be before number-like checks
		layouts := []string{time.RFC3339, "2006-01-02", "2006-01", "01/02/2006"}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				year := r.faker.IntRange(2000, 2039)
				month := time.Month(r.faker.IntRange(1, 12))
				day := r.faker.IntRange(1, 28)

				hour, min, sec := t.Clock()
				nsec := t.Nanosecond()
				loc := t.Location()

				newDate := time.Date(year, month, day, hour, min, sec, nsec, loc)
				return newDate.Format(layout)
			}
		}
		if numberLikeRegex.MatchString(v) {
			var result strings.Builder
			for _, char := range v {
				if char >= '0' && char <= '9' {
					result.WriteString(strconv.Itoa(r.faker.IntRange(0, 9)))
				} else {
					result.WriteRune(char)
				}
			}
			return result.String()
		}
		if _, err := url.ParseRequestURI(v); err == nil {
			return r.faker.URL()
		}
		if emailRegex.MatchString(v) {
			return r.faker.Email()
		}
		if _, err := net.ParseMAC(v); err == nil {
			return r.faker.MacAddress()
		}
		if ip := net.ParseIP(v); ip != nil {
			if ip.To4() != nil {
				return r.faker.IPv4Address()
			}
			return r.faker.IPv6Address()
		}

		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			return r.faker.Numerify(strings.Repeat("#", len(v)))
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
			return r.faker.Numerify(template)
		}

		words := strings.Split(v, " ")
		maskedWords := make([]string, len(words))
		caser := cases.Title(language.English)
		for i, word := range words {
			maskedWord := r.faker.Word()
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
			integerPart := parts[0]
			fractionalPart := parts[1]
			template := strings.Repeat("#", len(integerPart)) + "." + strings.Repeat("#", len(fractionalPart))
			return json.Number(r.faker.Numerify(template))
		}
		return json.Number(r.faker.Numerify(strings.Repeat("#", len(s))))

	case bool:
		return r.faker.Bool()
	}

	return "[MASKED]"
}
