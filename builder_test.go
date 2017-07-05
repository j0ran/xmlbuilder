package xmlbuilder

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"reflect"
	"testing"
)

func tokenEquals(a, b xml.Token) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	switch ta := a.(type) {
	case xml.StartElement: // ignore order of attributes
		tb := b.(xml.StartElement)
		if tb.Name != ta.Name {
			return false
		}
		attra := make(map[xml.Name]string)
		for _, a := range ta.Attr {
			attra[a.Name] = a.Value
		}
		attrb := make(map[xml.Name]string)
		for _, a := range tb.Attr {
			attrb[a.Name] = a.Value
		}
		if !reflect.DeepEqual(attra, attrb) {
			return false
		}
	default:
		if !reflect.DeepEqual(a, b) {
			return false
		}
	}
	return true
}

func assertXMLEquals(t *testing.T, a, b string) {
	xmla := xml.NewDecoder(bytes.NewBufferString(a))
	xmlb := xml.NewDecoder(bytes.NewBufferString(b))

	tokena, _ := xmla.Token()
	tokenb, _ := xmlb.Token()
	for tokena != nil || tokenb != nil {
		if !tokenEquals(tokena, tokenb) {
			t.Errorf("%s and %s are not equal", a, b)
			return
		}
		tokena, _ = xmla.Token()
		tokenb, _ = xmlb.Token()
	}
}

func readString(filename string) string {
	buf, _ := ioutil.ReadFile(filename)
	return string(buf)
}

func TestTag(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Tag("Joran")
	assertXMLEquals(t, "<Joran />\n", buf.String())
}

func TestAttributes(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Tag("person", "name", "Joran", "age", 40)
	assertXMLEquals(t, `<person name="Joran" age="40" />`+"\n", buf.String())
}

func TestElementAttr(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person")
	xml.Attr("name", "Joran")
	xml.Attr("age", 40)
	xml.End()
	assertXMLEquals(t, `<person name="Joran" age="40" />`+"\n", buf.String())
}

func TestNestedElement(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person", "name", "Joran")
	{
		xml.Tag("tel", "nr", "1276536271")
		xml.Element("tel", "nr", "1232123212").End()
	}
	xml.End()
	assertXMLEquals(t, readString("test/nested_element.xml"), buf.String())
}

func TestChars(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person", "name", "Joran").Chars("Hello!").End()
	assertXMLEquals(t, readString("test/chars.xml"), buf.String())

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Element("person", "Hello!", "name", "Joran").End()
	assertXMLEquals(t, readString("test/chars.xml"), buf.String())

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Tag("person", "Hello!", "name", "Joran")
	assertXMLEquals(t, readString("test/chars_inline.xml"), buf.String())
}

func TestAttr(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person", "name", "Joran")
	xml.Attr("age", 40)
	xml.End()
	assertXMLEquals(t, `<person name="Joran" age="40" />`+"\n", buf.String())

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Attr("age", 40)
	xml.Element("person", "name", "Joran")
	xml.End()
	assertXMLEquals(t, `<person name="Joran" age="40" />`+"\n", buf.String())
}

func TestInstructXML(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.InstructXML()
	xml.Element("address")
	{
		xml.Attr("id", 12)
		xml.Tag("street", "Some street")
		xml.Tag("city", "Eindhoven")
		xml.Tag("phone", "1298376142", "type", "mobile")
	}
	xml.End()
	assertXMLEquals(t, readString("test/instructxml.xml"), buf.String())
}

func TestNotPretty(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf).Pretty(false)
	xml.Element("address")
	{
		xml.Attr("id", 12)
		xml.Tag("street", "Some street")
		xml.Tag("city", "Eindhoven")
		xml.Tag("phone", "1298376142", "type", "mobile")
	}
	xml.End()
	assertXMLEquals(t, readString("test/not_pretty.xml"), buf.String())
}
