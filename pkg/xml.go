package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type XMLMasker struct {
	method Method
}

func NewXMLMasker(method Method) *XMLMasker {
	return &XMLMasker{
		method: method,
	}
}

func (mt *XMLMasker) Mask(r io.Reader, w io.Writer) error {
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
				maskedValue := mt.method.Mask(originalValue)
				maskedString := fmt.Sprintf("%v", maskedValue)

				if strings.TrimSpace(originalValue) != "" && len(originalValue) > 3 && originalValue == maskedString {
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
				maskedValue := mt.method.Mask(trimmedData)
				maskedString := fmt.Sprintf("%v", maskedValue)

				if trimmedData == maskedString && len(trimmedData) > 3 {
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
