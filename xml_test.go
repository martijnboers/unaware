package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestXMLMasker_EmptyElement(t *testing.T) {
	inputXML := `<root></root>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with empty element error = %v", err)
	}

	if !strings.Contains(out.String(), "<root>") {
		t.Errorf("Expected output to contain <root> tag")
	}
}

func TestXMLMasker_Attributes(t *testing.T) {
	inputXML := `<person name="John Doe"></person>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with attributes error = %v", err)
	}

	// Note: The current implementation does not mask attributes.
	// This test is to ensure attributes are preserved.
	if !strings.Contains(out.String(), `name="John Doe"`) {
		t.Errorf("Expected attribute to be preserved")
	}
}

func TestXMLMasker_NestedElements(t *testing.T) {
	inputXML := `<root><parent><child>value</child></parent></root>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with nested elements error = %v", err)
	}

	if strings.Contains(out.String(), ">value<") {
		t.Errorf("Expected 'value' to be masked")
	}
}

func TestXMLMasker_MixedContent(t *testing.T) {
	inputXML := `<root>text<element>value</element>more text</root>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with mixed content error = %v", err)
	}

	if strings.Contains(out.String(), ">text<") {
		t.Errorf("Expected 'text' to be masked")
	}
	if strings.Contains(out.String(), ">value<") {
		t.Errorf("Expected 'value' to be masked")
	}
	if strings.Contains(out.String(), ">more text<") {
		t.Errorf("Expected 'more text' to be masked")
	}
}

func TestXMLMasker_SelfClosingTag(t *testing.T) {
	inputXML := `<root><child/></root>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with self-closing tag error = %v", err)
	}

	// The encoder might expand the self-closing tag, so we check for both forms.
	if !strings.Contains(out.String(), "<child/>") && !strings.Contains(out.String(), "<child></child>") {
		t.Errorf("Expected output to contain <child/> tag")
	}
}

func TestXMLMasker_CommentsAndProcInst(t *testing.T) {
	inputXML := `<root><!-- a comment --><?target instruction?><value>data</value></root>`
	m := NewConsistentMasker()
	xm := NewXMLMasker(m)

	var in bytes.Buffer
	in.WriteString(inputXML)
	var out bytes.Buffer

	err := xm.Mask(&in, &out)
	if err != nil {
		t.Fatalf("XMLMasker.Mask() with comments and proc inst error = %v", err)
	}

	if !strings.Contains(out.String(), "<!-- a comment -->") {
		t.Error("Expected comment to be preserved")
	}
	if !strings.Contains(out.String(), "<?target instruction?>") {
		t.Error("Expected processing instruction to be preserved")
	}
	if strings.Contains(out.String(), ">data<") {
		t.Error("Expected 'data' to be masked")
	}
}
