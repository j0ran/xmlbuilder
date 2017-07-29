package xmlbuilder

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func readString(filename string) string {
	buf, _ := ioutil.ReadFile(filename)
	return string(buf)
}

func TestTag(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Tag("Joran")
	if a, b := "<Joran />\n", buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func TestAttributes(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Tag("person", "name", "Joran", "age", 40)
	if a, b := `<person name="Joran" age="40" />`+"\n", buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func TestElementAttr(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person")
	xml.Attr("name", "Joran")
	xml.Attr("age", 40)
	xml.End()
	if a, b := `<person name="Joran" age="40" />`+"\n", buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
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
	if a, b := readString("test/nested_element.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func TestChars(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person", "name", "Joran").Chars("Hello!").End()
	if a, b := readString("test/chars.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Element("person", "Hello!", "name", "Joran").End()
	if a, b := readString("test/chars.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Tag("person", "Hello!", "name", "Joran")
	if a, b := readString("test/chars_inline.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func TestAttr(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf)
	xml.Element("person", "name", "Joran")
	xml.Attr("age", 40)
	xml.End()
	if a, b := `<person name="Joran" age="40" />`+"\n", buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}

	buf = &bytes.Buffer{}
	xml = New(buf)
	xml.Attr("age", 40)
	xml.Element("person", "name", "Joran")
	xml.End()
	if a, b := `<person age="40" name="Joran" />`+"\n", buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
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
	if a, b := readString("test/instructxml.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func TestNotPretty(t *testing.T) {
	buf := &bytes.Buffer{}
	xml := New(buf).Inline()
	xml.Element("address")
	{
		xml.Attr("id", 12)
		xml.Tag("street", "Some street")
		xml.Tag("city", "Eindhoven")
		xml.Tag("phone", "1298376142", "type", "mobile")
	}
	xml.End()
	if a, b := readString("test/not_pretty.xml"), buf.String(); a != b {
		t.Errorf("%s and %s are not equal", a, b)
	}
}

func Example() {
	xml := New(os.Stdout)
	xml.Element("people")
	{
		xml.Element("person", "id", 1)
		{
			xml.Tag("name", "Joran")
			xml.Tag("age", 40)
		}
		xml.End()
	}
	xml.End()
	// Output:
	// <people>
	//   <person id="1">
	//     <name>Joran</name>
	//     <age>40</age>
	//   </person>
	// </people>
}

func ExampleDoctype() {
	xml := New(os.Stdout)
	xml.Doctype(DoctypeHTML5)
	xml.Inline().Element("p").Chars("Hello").End().EndInline()
	// Output:
	// <!DOCTYPE html>
	// <p>Hello</p>
}

func ExampleInline() {
	xml := New(os.Stdout)
	xml.Doctype(DoctypeHTML5)
	xml.Element("ul")
	{
		xml.Inline().Element("li").Chars("Hello ").Tag("b", "there").End().EndInline()
		xml.Inline().Element("li").Chars("Test ").Tag("b", "this").End().EndInline()
	}
	xml.End()
	// Output:
	// <!DOCTYPE html>
	// <ul>
	//   <li>Hello <b>there</b></li>
	//   <li>Test <b>this</b></li>
	// </ul>
}
