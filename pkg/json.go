package pkg

import (
	"encoding/json"
	"io"
)

type JSONProcessor struct {
	Method
}

func NewJSONProcessor(method Method) *JSONProcessor {
	return &JSONProcessor{
		Method: method,
	}
}

func (jm *JSONProcessor) Mask(r io.Reader, w io.Writer) error {
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

func (jm *JSONProcessor) maskValue(data any) any {
	switch v := data.(type) {
	case json.Number:
		return jm.Method.Mask(v)
	case map[string]any:
		return jm.maskMap(v)
	case []any:
		return jm.maskSlice(v)
	default:
		return jm.Method.Mask(v)
	}
}

func (jm *JSONProcessor) maskMap(data map[string]any) map[string]any {
	maskedMap := make(map[string]any, len(data))
	for key, value := range data {
		maskedMap[key] = jm.maskValue(value)
	}
	return maskedMap
}

func (jm *JSONProcessor) maskSlice(data []any) []any {
	maskedSlice := make([]any, len(data))
	for i, value := range data {
		maskedSlice[i] = jm.maskValue(value)
	}
	return maskedSlice
}
