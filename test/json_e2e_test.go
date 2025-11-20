package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"unaware/pkg"
)

const jsonTestData = `{
  "customers": [
    {
      "id": "CUST001",
      "personal_info": {
        "first_name": "John",
        "last_name": "Doe",
        "date_of_birth": "1985-03-15",
        "ssn": "123-456-7890"
      },
      "contact_details": {
        "email": "john.doe@example.com",
        "phone": "555-0101",
        "address": {
          "street": "123 Maple Street",
          "city": "Anytown",
          "state": "CA",
          "zip_code": "90210"
        }
      },
      "financial_data": {
        "account_balance": 15230.55,
        "credit_card": {
          "card_number": "4111222233334444",
          "expiry_date": "2026-12",
          "cvv": "123"
        },
        "last_transaction_amount": 250.75,
        "last_transaction_date": "2025-11-15T14:30:00Z"
      },
      "medical_records": {
        "record_id": "MRN-JD-1123",
        "last_visit_date": "2025-10-20",
        "cholesterol_level": 205.7
      },
      "login_info": {
        "username": "johndoe85",
        "password": "S3cureP@ssw0rd!",
        "last_login_ip": "192.168.1.101",
        "last_login_timestamp": "2025-11-17T09:15:22Z"
      }
    }
  ]
}`

func TestJSONMasker_EndToEnd(t *testing.T) {
	masker := pkg.NewJSONProcessor(pkg.NewSaltedMethod(testSalt))
	var buf bytes.Buffer
	err := masker.Mask(strings.NewReader(jsonTestData), &buf)
	if err != nil {
		t.Fatalf("Failed to mask JSON: %v", err)
	}
	maskedJSON := buf.Bytes()

	var result map[string]any
	if err := json.Unmarshal(maskedJSON, &result); err != nil {
		t.Fatalf("Failed to unmarshal masked JSON: %v", err)
	}

	customers := result["customers"].([]any)
	for _, c := range customers {
		customer := c.(map[string]any)
		personalInfo := customer["personal_info"].(map[string]any)
		contactDetails := customer["contact_details"].(map[string]any)
		financialData := customer["financial_data"].(map[string]any)
		loginInfo := customer["login_info"].(map[string]any)

		if personalInfo["ssn"].(string) == "123-456-7890" {
			t.Error("ssn was not masked")
		}
		if contactDetails["email"].(string) == "john.doe@example.com" {
			t.Error("email was not masked")
		}
		if financialData["credit_card"].(map[string]interface{})["card_number"].(string) == "4111222233334444" {
			t.Error("card_number was not masked")
		}
		if loginInfo["last_login_ip"].(string) == "192.168.1.101" {
			t.Error("last_login_ip was not masked")
		}
	}
}
