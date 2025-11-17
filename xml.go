package main

import (
	"encoding/xml"
	"fmt"
	"io"
)

// XMLMasker masks an XML stream.
type XMLMasker struct {
	masker Masker
}

// NewXMLMasker creates a new XMLMasker.
func NewXMLMasker(masker Masker) *XMLMasker {
	return &XMLMasker{
		masker: masker,
	}
}

// Mask masks the XML stream from the reader and writes it to the writer.
func (xm *XMLMasker) Mask(r io.Reader, w io.Writer) error {
	decoder := xml.NewDecoder(r)
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch se := token.(type) {
		case xml.CharData:
			maskedValue := xm.masker.Mask(string(se))
			// Convert the masked value (which could be string, float, or bool) back to a string.
			maskedString := fmt.Sprintf("%v", maskedValue)
			if err := encoder.EncodeToken(xml.CharData(maskedString)); err != nil {
				return err
			}
		default:
			if err := encoder.EncodeToken(token); err != nil {
				return err
			}
		}
	}

	return encoder.Flush()
}
