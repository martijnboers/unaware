package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type XMLMasker struct {
	masker Masker
}

func NewXMLMasker(masker Masker) *XMLMasker {
	return &XMLMasker{
		masker: masker,
	}
}

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
		case xml.StartElement:
			for i := range se.Attr {
				attr := &se.Attr[i]
				originalValue := attr.Value
				maskedValue := xm.masker.Mask(originalValue)
				maskedString := fmt.Sprintf("%v", maskedValue)

				if strings.TrimSpace(originalValue) != "" && originalValue == maskedString {
					fmt.Fprintf(os.Stderr, "Error: attribute was not masked: %s\n", originalValue)
					os.Exit(1)
				}

				attr.Value = maskedString
			}
			if err := encoder.EncodeToken(se); err != nil {
				return err
			}
		case xml.CharData:
			trimmedData := strings.TrimSpace(string(se))
			if len(trimmedData) > 0 {
				maskedValue := xm.masker.Mask(trimmedData)
				maskedString := fmt.Sprintf("%v", maskedValue)

				if trimmedData == maskedString {
					fmt.Fprintf(os.Stderr, "Error: value was not masked: %s\n", trimmedData)
					os.Exit(1)
				}

				if err := encoder.EncodeToken(xml.CharData(maskedString)); err != nil {
					return err
				}
			}
		default:
			if err := encoder.EncodeToken(token); err != nil {
				return err
			}
		}
	}

	return encoder.Flush()
}
