package test

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strings"
	"testing"

	"unaware/pkg"
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

func TestXMLMasker_WithAttributes(t *testing.T) {
	inputXML := `<data><item id="a1b2-c3d4-e5f6" secret="secret-value">test</item></data>`
	m := pkg.NewHashedMethod(testSalt)
	xm := pkg.NewXMLProcessor(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with attributes error = %v", err)
	}

	output := out.String()
	if strings.Contains(output, "secret-value") {
		t.Errorf("Expected sensitive attribute 'secret-value' to be masked, but it was not.")
	}
	if strings.Contains(output, "a1b2-c3d4-e5f6") {
		t.Errorf("Expected id attribute 'a1b2-c3d4-e5f6' to be masked, but it was not.")
	}
}

func TestXMLMasker_ComprehensiveFormatValidation(t *testing.T) {
	m := pkg.NewHashedMethod(testSalt)
	xm := pkg.NewXMLProcessor(m)

	var in bytes.Buffer
	in.WriteString(comprehensiveXMLInput)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with comprehensive XML error = %v", err)
	}

	expectedFormats := map[string]*regexp.Regexp{
		"date_of_birth":         regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
		"ssn":                   regexp.MustCompile(`^\d{3}-\d{2,3}-\d{4}$`),
		"email":                 regexp.MustCompile(`^.+@.+\..+$`),
		"zip_code":              regexp.MustCompile(`^\d{5}$`),
		"card_number":           regexp.MustCompile(`^\d{16}$`),
		"expiry_date":           regexp.MustCompile(`^\d{4}-\d{2}$`),
		"cvv":                   regexp.MustCompile(`^\d{3}$`),
		"last_transaction_date": regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),
		"last_visit_date":       regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
		"last_login_ip":         regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`),
		"last_login_timestamp":  regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),
	}

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
}
