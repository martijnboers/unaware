package main

import (
	"encoding/xml"
	"fmt"
	"io"
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
		case xml.CharData:
			trimmedData := strings.TrimSpace(string(se))
			if len(trimmedData) > 0 {
				maskedValue := xm.masker.Mask(trimmedData)
				maskedString := fmt.Sprintf("%v", maskedValue)
				if err := encoder.EncodeToken(xml.CharData(maskedString)); err != nil {
					return err
				}
			} else {
				if err := encoder.EncodeToken(se); err != nil {
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