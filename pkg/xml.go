package pkg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type XMLProcessor struct {
	Method
}

func NewXMLProcessor(method Method) *XMLProcessor {
	return &XMLProcessor{
		Method: method,
	}
}

func (mt *XMLProcessor) Mask(r io.Reader, w io.Writer) error {
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
				maskedValue := mt.Method.Mask(originalValue)
				maskedString := fmt.Sprintf("%v", maskedValue)
				attr.Value = maskedString
			}
			if err := encoder.EncodeToken(se); err != nil {
				return err
			}
		case xml.CharData:
			trimmedData := strings.TrimSpace(string(se))
			if len(trimmedData) > 0 {
				maskedValue := mt.Method.Mask(trimmedData)
				maskedString := fmt.Sprintf("%v", maskedValue)

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
