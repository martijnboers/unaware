package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strings"
	"testing"
)

const comprehensiveXMLInput = `<?xml version="1.0" encoding="UTF-8"?>
<customers>
  <customer id="CUST001">
    <personal_info>
      <first_name>John</first_name>
      <last_name>Doe</last_name>
      <date_of_birth>1985-03-15</date_of_birth>
      <ssn>123-456-7890</ssn>
    </personal_info>
    <contact_details>
      <email>john.doe@example.com</email>
      <phone>555-0101</phone>
      <address>
        <street>123 Maple Street</street>
        <city>Anytown</city>
        <state>CA</state>
        <zip_code>90210</zip_code>
      </address>
    </contact_details>
    <financial_data>
      <account_balance>15230.55</account_balance>
      <credit_card>
        <card_number>4111222233334444</card_number>
        <expiry_date>2026-12</expiry_date>
        <cvv>123</cvv>
      </credit_card>
      <last_transaction_amount>250.75</last_transaction_amount>
      <last_transaction_date>2025-11-15T14:30:00Z</last_transaction_date>
    </financial_data>
    <medical_records>
        <record_id>MRN-JD-1123</record_id>
        <last_visit_date>2025-10-20</last_visit_date>
        <cholesterol_level>205.7</cholesterol_level>
    </medical_records>
    <login_info>
        <username>johndoe85</username>
        <password>S3cureP@ssw0rd!</password>
        <last_login_ip>192.168.1.101</last_login_ip>
        <last_login_timestamp>2025-11-17T09:15:22Z</last_login_timestamp>
    </login_info>
  </customer>
</customers>
`

func TestXMLMasker_ComprehensiveFormatValidation(t *testing.T) {
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(comprehensiveXMLInput)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with comprehensive XML error = %v", err)
	}

	// Define the expected format for each element's masked data.
	expectedFormats := map[string]*regexp.Regexp{
		"first_name":            regexp.MustCompile(`^[a-z]+$`),
		"last_name":             regexp.MustCompile(`^[a-z]+$`),
		"date_of_birth":         regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
		"ssn":                   regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`),
		"email":                 regexp.MustCompile(`^[a-z]+@[a-z]+\.[a-z]{2,3}$`),
		"phone":                 regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"street":                regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"city":                  regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"state":                 regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"zip_code":              regexp.MustCompile(`^\d{5}$`),
		"account_balance":       regexp.MustCompile(`^-?\d+\.\d{2}$`),
		"card_number":           regexp.MustCompile(`^\d{16}$`),
		"expiry_date":           regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"cvv":                   regexp.MustCompile(`^\d{3}$`),
		"last_transaction_amount": regexp.MustCompile(`^-?\d+\.\d{2}$`),
		"last_transaction_date": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),
		"record_id":             regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"last_visit_date":       regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
		"cholesterol_level":     regexp.MustCompile(`^-?\d+\.\d{2}$`),
		"username":              regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"password":              regexp.MustCompile(`^[a-z]+$`), // Falls back to generic
		"last_login_ip":         regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`),
		"last_login_timestamp":  regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),
	}

	// Manually decode the XML token by token to check element content.
	decoder := xml.NewDecoder(&out)
	var currentKey string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error decoding masked XML: %v", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			currentKey = se.Name.Local
		case xml.CharData:
			data := string(se)
			if pattern, ok := expectedFormats[currentKey]; ok {
				if !pattern.MatchString(data) {
					t.Errorf("Element <%s> has incorrectly formatted masked data. Got: '%s', Expected pattern: '%s'", currentKey, data, pattern.String())
				}
			}
			currentKey = "" // Reset key after checking content
		}
	}

	// Final check: ensure no sensitive strings from the original appear in the output
	sensitiveStrings := []string{
		">John<", ">Doe<", ">123-456-7890<", ">john.doe@example.com<",
		">S3cureP@ssw0rd!<", ">4111222233334444<", ">192.168.1.101<",
	}
	maskedOutputString := out.String()
	for _, s := range sensitiveStrings {
		if strings.Contains(maskedOutputString, s) {
			t.Errorf("Masked output contains sensitive string wrapped in tags: %s", s)
		}
	}
}
