package main

import (
	"encoding/json"
	"io"
)

type JSONMasker struct {
	masker Masker
}

func NewJSONMasker(masker Masker) *JSONMasker {
	return &JSONMasker{
		masker: masker,
	}
}

func (jm *JSONMasker) Mask(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	for {
		var data interface{}
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

func (jm *JSONMasker) maskValue(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return jm.maskMap(v)
	case []interface{}:
		return jm.maskSlice(v)
	default:
		return jm.masker.Mask(v)
	}
}

func (jm *JSONMasker) maskMap(data map[string]interface{}) map[string]interface{} {
	maskedMap := make(map[string]interface{}, len(data))
	for key, value := range data {
		maskedMap[key] = jm.maskValue(value)
	}
	return maskedMap
}

func (jm *JSONMasker) maskSlice(data []interface{}) []interface{} {
	maskedSlice := make([]interface{}, len(data))
	for i, value := range data {
		maskedSlice[i] = jm.maskValue(value)
	}
	return maskedSlice
}
