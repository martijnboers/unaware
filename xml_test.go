package main

import (
	"bytes"
	"encoding/xml"
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
  <customer id="CUST002">
    <personal_info>
      <first_name>Jane</first_name>
      <last_name>Smith</last_name>
      <date_of_birth>1992-07-22</date_of_birth>
      <ssn>987-654-3210</ssn>
    </personal_info>
    <contact_details>
      <email>jane.smith@example.com</email>
      <phone>555-0102</phone>
      <address>
        <street>456 Oak Avenue</street>
        <city>Someville</city>
        <state>NY</state>
        <zip_code>10001</zip_code>
      </address>
    </contact_details>
    <financial_data>
      <account_balance>5430.12</account_balance>
      <credit_card>
        <card_number>5555666677778888</card_number>
        <expiry_date>2027-08</expiry_date>
        <cvv>456</cvv>
      </credit_card>
      <last_transaction_amount>-50.25</last_transaction_amount>
      <last_transaction_date>2025-11-16T18:45:10Z</last_transaction_date>
    </financial_data>
    <medical_records>
        <record_id>MRN-JS-4567</record_id>
        <last_visit_date>2025-09-05</last_visit_date>
        <cholesterol_level>180.2</cholesterol_level>
    </medical_records>
    <login_info>
        <username>janesmith92</username>
        <password>MyP@ssw0rd123</password>
        <last_login_ip>10.0.0.5</last_login_ip>
        <last_login_timestamp>2025-11-17T08:30:55Z</last_login_timestamp>
    </login_info>
  </customer>
    <customer id="CUST003">
    <personal_info>
      <first_name>Robert</first_name>
      <last_name>Jones</last_name>
      <date_of_birth>1978-11-30</date_of_birth>
      <ssn>555-123-4567</ssn>
    </personal_info>
    <contact_details>
      <email>robert.jones@example.com</email>
      <phone>555-0103</phone>
      <address>
        <street>789 Pine Lane</street>
        <city>Metrocity</city>
        <state>TX</state>
        <zip_code>75001</zip_code>
      </address>
    </contact_details>
    <financial_data>
      <account_balance>250000.00</account_balance>
      <credit_card>
        <card_number>371234567890123</card_number>
        <expiry_date>2025-10</expiry_date>
        <cvv>7890</cvv>
      </credit_card>
      <last_transaction_amount>1234.56</last_transaction_amount>
      <last_transaction_date>2025-11-12T11:05:00Z</last_transaction_date>
    </financial_data>
    <medical_records>
        <record_id>MRN-RJ-8901</record_id>
        <last_visit_date>2025-11-01</last_visit_date>
        <cholesterol_level>220.5</cholesterol_level>
    </medical_records>
    <login_info>
        <username>robjones78</username>
        <password>Ch@ngeMe!78</password>
        <last_login_ip>203.0.113.42</last_login_ip>
        <last_login_timestamp>2025-11-16T22:10:18Z</last_login_timestamp>
    </login_info>
  </customer>
</customers>
`

// A struct that mirrors the XML structure for robust testing.
type Customers struct {
	Customers []Customer `xml:"customer"`
}
type Customer struct {
	ID              string           `xml:"id,attr"`
	PersonalInfo    PersonalInfo     `xml:"personal_info"`
	ContactDetails  ContactDetails   `xml:"contact_details"`
	FinancialData   FinancialData    `xml:"financial_data"`
	MedicalRecords  MedicalRecords   `xml:"medical_records"`
	LoginInfo       LoginInfo        `xml:"login_info"`
}
type PersonalInfo struct {
	FirstName   string `xml:"first_name"`
	LastName    string `xml:"last_name"`
	DateOfBirth string `xml:"date_of_birth"`
	SSN         string `xml:"ssn"`
}
type ContactDetails struct {
	Email   string  `xml:"email"`
	Phone   string  `xml:"phone"`
	Address Address `xml:"address"`
}
type Address struct {
	Street  string `xml:"street"`
	City    string `xml:"city"`
	State   string `xml:"state"`
	ZipCode string `xml:"zip_code"`
}
type FinancialData struct {
	AccountBalance        string      `xml:"account_balance"`
	CreditCard            CreditCard  `xml:"credit_card"`
	LastTransactionAmount string      `xml:"last_transaction_amount"`
	LastTransactionDate   string      `xml:"last_transaction_date"`
}
type CreditCard struct {
	CardNumber string `xml:"card_number"`
	ExpiryDate string `xml:"expiry_date"`
	CVV        string `xml:"cvv"`
}
type MedicalRecords struct {
	RecordID         string `xml:"record_id"`
	LastVisitDate    string `xml:"last_visit_date"`
	CholesterolLevel string `xml:"cholesterol_level"`
}
type LoginInfo struct {
	Username           string `xml:"username"`
	Password           string `xml:"password"`
	LastLoginIP        string `xml:"last_login_ip"`
	LastLoginTimestamp string `xml:"last_login_timestamp"`
}

func TestXMLMasker_ComprehensivePII(t *testing.T) {
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(comprehensiveXMLInput)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with comprehensive XML error = %v", err)
	}

	// 1. Unmarshal original and masked XML to compare structures
	var originalData, maskedData Customers
	if err := xml.Unmarshal([]byte(comprehensiveXMLInput), &originalData); err != nil {
		t.Fatalf("Failed to unmarshal original XML: %v", err)
	}
	if err := xml.Unmarshal(out.Bytes(), &maskedData); err != nil {
		t.Fatalf("Failed to unmarshal masked XML: %v", err)
	}

	// 2. Check structural integrity
	if len(originalData.Customers) != len(maskedData.Customers) {
		t.Fatalf("Number of customers changed. Original: %d, Masked: %d", len(originalData.Customers), len(maskedData.Customers))
	}

	// 3. Check a sample of fields to ensure attributes are preserved and values are masked
	for i := 0; i < len(originalData.Customers); i++ {
		originalCust := originalData.Customers[i]
		maskedCust := maskedData.Customers[i]

		// Attribute should be preserved
		if originalCust.ID != maskedCust.ID {
			t.Errorf("Customer ID attribute was modified. Original: %s, Masked: %s", originalCust.ID, maskedCust.ID)
		}

		// Personal info should be masked
		if originalCust.PersonalInfo.FirstName == maskedCust.PersonalInfo.FirstName {
			t.Errorf("First name was not masked for customer %s", originalCust.ID)
		}
		if originalCust.PersonalInfo.SSN == maskedCust.PersonalInfo.SSN {
			t.Errorf("SSN was not masked for customer %s", originalCust.ID)
		}

		// Contact details should be masked
		if originalCust.ContactDetails.Email == maskedCust.ContactDetails.Email {
			t.Errorf("Email was not masked for customer %s", originalCust.ID)
		}
		if originalCust.ContactDetails.Address.Street == maskedCust.ContactDetails.Address.Street {
			t.Errorf("Street was not masked for customer %s", originalCust.ID)
		}

		// Financial data should be masked
		if originalCust.FinancialData.CreditCard.CardNumber == maskedCust.FinancialData.CreditCard.CardNumber {
			t.Errorf("Card number was not masked for customer %s", originalCust.ID)
		}
		if originalCust.FinancialData.AccountBalance == maskedCust.FinancialData.AccountBalance {
			t.Errorf("Account balance was not masked for customer %s", originalCust.ID)
		}

		// Login info should be masked
		if originalCust.LoginInfo.Password == maskedCust.LoginInfo.Password {
			t.Errorf("Password was not masked for customer %s", originalCust.ID)
		}
		if originalCust.LoginInfo.LastLoginIP == maskedCust.LoginInfo.LastLoginIP {
			t.Errorf("Last login IP was not masked for customer %s", originalCust.ID)
		}
	}

	// 4. Final check: ensure no sensitive strings from the original appear in the output
	sensitiveStrings := []string{
		"John", "Doe", "123-456-7890", "john.doe@example.com", "S3cureP@ssw0rd!",
		"Jane", "Smith", "987-654-3210", "jane.smith@example.com", "MyP@ssw0rd123",
		"Robert", "Jones", "555-123-4567", "robert.jones@example.com", "Ch@ngeMe!78",
		"4111222233334444", "192.168.1.101",
	}

	maskedOutputString := out.String()
	for _, s := range sensitiveStrings {
		if strings.Contains(maskedOutputString, s) {
			t.Errorf("Masked output contains sensitive string: %s", s)
		}
	}
}