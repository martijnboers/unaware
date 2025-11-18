package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type JSONMasker struct {
	method Method
}

func NewJSONMasker(method Method) *JSONMasker {
	return &JSONMasker{
		method: method,
	}
}

func (jm *JSONMasker) Mask(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	for {
		var data any
		if err := decoder.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		maskedData := jm.maskValue(data)

		if err := encoder.Encode(&maskedData); err != nil {
			return err
		}
	}

	return nil
}

func (jm *JSONMasker) maskValue(data any) any {
	switch v := data.(type) {
	case json.Number:
		maskedValue := jm.method.Mask(v)
		if v.String() == maskedValue.(json.Number).String() && len(v.String()) > 3 {
			fmt.Fprintf(os.Stderr, "Error: json.Number was not masked: %v\n", v)
			os.Exit(1)
		}
		return maskedValue
	case map[string]any:
		return jm.maskMap(v)
	case []any:
		return jm.maskSlice(v)
	default:
		maskedValue := jm.method.Mask(v)
		if data != nil && data == maskedValue {
			if _, isBool := data.(bool); isBool {
				return maskedValue // Don't check booleans
			}
			if s, isString := data.(string); isString && strings.TrimSpace(s) == "" {
				return maskedValue // Don't check whitespace-only strings
			}
			fmt.Fprintf(os.Stderr, "Error: value was not masked: %v\n", data)
			os.Exit(1)
		}
		return maskedValue
	}
}

func (jm *JSONMasker) maskMap(data map[string]any) map[string]any {
	maskedMap := make(map[string]any, len(data))
	for key, value := range data {
		maskedMap[key] = jm.maskValue(value)
	}
	return maskedMap
}

func (jm *JSONMasker) maskSlice(data []any) []any {
	maskedSlice := make([]any, len(data))
	for i, value := range data {
		maskedSlice[i] = jm.maskValue(value)
	}
	return maskedSlice
}
