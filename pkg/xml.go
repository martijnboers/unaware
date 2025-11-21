package pkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// xmlProcessor is an unexported, internal struct that implements the processor interface for XML.
type xmlProcessor struct {
	// methodFactory creates new, thread-safe masker instances for each worker.
	methodFactory func() *masker
}

// newXMLProcessor is the internal constructor for the xmlProcessor.
func newXMLProcessor(strategy MaskingStrategy) *xmlProcessor {
	return &xmlProcessor{
		methodFactory: func() *masker {
			return newMasker(strategy)
		},
	}
}

// Process acts as the foreman for XML processing, deciding which strategy to use.
func (xp *xmlProcessor) Process(r io.Reader, w io.Writer) error {
	// We use a TeeReader to buffer the start of the stream for pattern detection.
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	decoder := xml.NewDecoder(tee)

	// --- Detect if the XML is a "list" of repeating elements ---
	root, _, _, ok := detectXMLListPattern(decoder)

	// Create a reader that prepends the consumed bytes, ensuring the next step sees the full stream.
	combinedReader := io.MultiReader(&buf, r)

	if ok {
		// Pattern found! Use the high-performance concurrent runner.
		runner := newConcurrentRunner(xp.methodFactory)
		chunkDecoder := xml.NewDecoder(combinedReader)
		chunkReader := xp.createXMLChunkReader(chunkDecoder, root.Name)
		assembler := &xmlAssembler{Root: root}
		return runner.Run(w, chunkReader, assembler)
	}

	// Pattern not found, fall back to the safe, single-threaded streamer.
	serialDecoder := xml.NewDecoder(combinedReader)
	return xp.processSerially(serialDecoder, w)
}

// --- xmlAssembler Implementation ---
// xmlAssembler knows how to write masked map chunks back into valid XML format.
type xmlAssembler struct {
	Root    xml.StartElement
	encoder *xml.Encoder
}

func (a *xmlAssembler) WriteStart(w io.Writer) error {
	a.encoder = xml.NewEncoder(w)
	a.encoder.Indent("", "  ")
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	return a.encoder.EncodeToken(a.Root)
}

func (a *xmlAssembler) WriteItem(w io.Writer, item any, isFirst bool) error {
	itemMap, ok := item.(map[string]any)
	if !ok {
		return fmt.Errorf("xml assembler expected map[string]any, but got %T", item)
	}
	// The map key is the element name, the value is the content map.
	for key, val := range itemMap {
		if valMap, ok := val.(map[string]any); ok {
			return mapToXML(key, valMap, a.encoder)
		}
	}
	return fmt.Errorf("unexpected structure in item map for XML assembler")
}

func (a *xmlAssembler) WriteEnd(w io.Writer) error {
	if err := a.encoder.EncodeToken(a.Root.End()); err != nil {
		return err
	}
	return a.encoder.Flush()
}

// mapToXML recursively writes a map back into valid XML tokens.
func mapToXML(key string, m map[string]any, enc *xml.Encoder) error {
	start := xml.StartElement{Name: xml.Name{Local: key}}
	// Extract and add attributes from the map
	for k, v := range m {
		if strings.HasPrefix(k, "-") {
			attrName := strings.TrimPrefix(k, "-")
			start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: attrName}, Value: fmt.Sprintf("%v", v)})
		}
	}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	// Handle character data content first
	if text, ok := m["#text"]; ok {
		if err := enc.EncodeToken(xml.CharData(fmt.Sprintf("%v", text))); err != nil {
			return err
		}
	}
	// Recursively handle child elements
	for k, v := range m {
		if !strings.HasPrefix(k, "-") && k != "#text" {
			if slice, ok := v.([]any); ok {
				for _, item := range slice {
					if itemMap, ok := item.(map[string]any); ok {
						if err := mapToXML(k, itemMap, enc); err != nil {
							return err
						}
					}
				}
			} else if nestedMap, ok := v.(map[string]any); ok {
				if err := mapToXML(k, nestedMap, enc); err != nil {
					return err
				}
			} else { // Handle simple key-value pairs
				if err := mapToXML(k, map[string]any{"#text": v}, enc); err != nil {
					return err
				}
			}
		}
	}
	return enc.EncodeToken(start.End())
}

// --- XML Detection and Chunking Logic ---

// detectXMLListPattern is a heuristic to check for a root element with repeating children.
func detectXMLListPattern(decoder *xml.Decoder) (xml.StartElement, xml.StartElement, xml.StartElement, bool) {
	var root, firstChild, secondChild xml.StartElement
	depth := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return root, firstChild, secondChild, false
		}
		switch se := token.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				root = se.Copy()
			} else if depth == 2 {
				if firstChild.Name.Local == "" {
					firstChild = se.Copy()
				} else {
					secondChild = se.Copy()
					if firstChild.Name.Local == secondChild.Name.Local {
						return root, firstChild, secondChild, true
					}
					return root, firstChild, secondChild, false
				}
			}
		case xml.EndElement:
			depth--
			if depth < 1 {
				return root, firstChild, secondChild, false
			}
		}
	}
}

// createXMLChunkReader creates the "Adaptor Tool" for XML.
func (xp *xmlProcessor) createXMLChunkReader(decoder *xml.Decoder, rootName xml.Name) chunkReader {
	var started bool
	return func() (any, error) {
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err // Includes io.EOF
			}
			switch se := token.(type) {
			case xml.StartElement:
				if !started {
					if se.Name.Local == rootName.Local && se.Name.Space == rootName.Space {
						started = true
						continue
					}
				}
				// We found a child element. Decode it and its contents into a map.
				elementMap, err := decodeElementToMap(decoder, se)
				if err != nil {
					return nil, err
				}
				// Wrap it in a map to preserve the element name for the assembler.
				return map[string]any{se.Name.Local: elementMap}, nil
			case xml.EndElement:
				if se.Name.Local == rootName.Local && se.Name.Space == rootName.Space {
					return nil, io.EOF // End of the root element
				}
			}
		}
	}
}

// decodeElementToMap reads one full XML element and converts it to a map[string]any.
func decodeElementToMap(decoder *xml.Decoder, start xml.StartElement) (map[string]any, error) {
	m := make(map[string]any)
	for _, attr := range start.Attr {
		m["-"+attr.Name.Local] = attr.Value
	}
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch se := token.(type) {
		case xml.StartElement:
			nestedMap, err := decodeElementToMap(decoder, se)
			if err != nil {
				return nil, err
			}
			key := se.Name.Local
			if existing, ok := m[key]; ok {
				if slice, ok := existing.([]any); ok {
					m[key] = append(slice, nestedMap)
				} else {
					m[key] = []any{existing, nestedMap}
				}
			} else {
				m[key] = nestedMap
			}
		case xml.CharData:
			text := strings.TrimSpace(string(se))
			if text != "" {
				m["#text"] = text
			}
		case xml.EndElement:
			if se.Name.Local == start.Name.Local && se.Name.Space == start.Name.Space {
				return m, nil
			}
		}
	}
}

// --- Serial Fallback Logic ---

// processSerially is the robust, single-threaded token streamer for XML.
func (xp *xmlProcessor) processSerially(decoder *xml.Decoder, w io.Writer) error {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	serialMasker := xp.methodFactory()
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
			startElem := se.Copy()
			for i := range startElem.Attr {
				attr := &startElem.Attr[i]
				maskedValue := serialMasker.mask(attr.Value)
				attr.Value = fmt.Sprintf("%v", maskedValue)
			}
			if err := encoder.EncodeToken(startElem); err != nil {
				return err
			}
		case xml.CharData:
			trimmedData := strings.TrimSpace(string(se))
			if len(trimmedData) > 0 {
				maskedValue := serialMasker.mask(trimmedData)
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
