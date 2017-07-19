package xmlbuilder

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

var (
	htmlEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
	)
	attrEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&#34;",
	)
)

const (
	DoctypeHTML5               = "html"
	DoctypeHTML4Strict         = `HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"`
	DoctypeHTML4Transitional   = `HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd"`
	DoctypeHTML4Frameset       = `HTML PUBLIC "-//W3C//DTD HTML 4.01 Frameset//EN" "http://www.w3.org/TR/html4/frameset.dtd"`
	DoctypeXHTML10Strict       = `html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"`
	DoctypeXHTML10Transitional = `html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"`
	DoctypeXHTML10Frameset     = `html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd"`
	DoctypeXHTML11             = `html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"`
)

type Builder struct {
	writer          io.Writer
	buildingElement bool
	attributes      map[string]string
	elements        []string
	indentString    string
	indent          string
	offset          int // indent offset
	inline          bool
	pretty          bool
}

func s(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// New creates a new xml builder. By default pretty print is enabled and the identation is
// two spaces.
func New(writer io.Writer) *Builder {
	builder := new(Builder)
	builder.writer = writer
	builder.attributes = make(map[string]string)
	builder.indent = "  "
	builder.pretty = true
	return builder
}

// Element defines a new element in the xml document.
func (b *Builder) Element(element string, args ...interface{}) *Builder {
	if b.buildingElement {
		b.outputElement(false, b.pretty)
	}

	b.buildingElement = true
	b.elements = append(b.elements, element)
	first := len(args) % 2
	for i := first; i < len(args); i += 2 {
		b.attributes[s(args[i+0])] = s(args[i+1])
	}
	if first != 0 {
		b.Chars(args[0])
	}

	return b
}

// Attr will add an attribute to the current element being build, or when not building
// an element it will add attributes to the next element to be build.
func (b *Builder) Attr(name string, value interface{}) *Builder {
	b.attributes[name] = s(value)
	return b
}

// End will add a close tag that matches the the previous Element call.
func (b *Builder) End() *Builder {
	if b.buildingElement {
		b.outputElement(true, b.pretty)
	} else {
		newline := "\n"
		if !b.pretty {
			newline = ""
		}
		if b.inline {
			fmt.Fprintf(b.writer, "</%s>%s", b.elements[len(b.elements)-1], newline)
		} else {
			fmt.Fprintf(b.writer, "%s</%s>%s", b.doIndent(), b.elements[len(b.elements)-1], newline)
		}
		b.elements = b.elements[:len(b.elements)-1]
	}
	return b
}

// Tag inserts an inline element and directly closes it.
func (b *Builder) Tag(element string, args ...interface{}) *Builder {
	b.inline = true
	b.Element(element, args...).End()
	b.inline = false
	return b
}

// Instruct can be used to generate instruction tags.
func (b *Builder) Instruct(name string, args ...interface{}) *Builder {
	fmt.Fprintf(b.writer, "<?%s", name)
	for i := 0; i < len(args); i += 2 {
		fmt.Fprintf(b.writer, ` %v="%s"`, args[i+0], attrEscaper.Replace(s(args[i+1])))
	}
	fmt.Fprintln(b.writer, "?>")
	return b
}

// InstructXML outputs an default xml instruction to be used at the beginning of an xml document.
// This is the same as calling Instruct("xml", "version", "1.0", "encoding", "UTF-8")
func (b *Builder) InstructXML() *Builder {
	return b.Instruct("xml", "version", "1.0", "encoding", "UTF-8")
}

func (b *Builder) Doctype(doctype string) *Builder {
	fmt.Fprintf(b.writer, "<!DOCTYPE %s>\n", doctype)
	return b
}

// Offset will add the delta value to the given ident offset
func (b *Builder) Offset(delta int) *Builder {
	b.offset += delta
	return b
}

// Chars add characters to the document. It will also escape special characters.
func (b *Builder) Chars(chars interface{}) *Builder {
	b.outputElement(false, b.pretty && !b.inline)
	if b.inline || !b.pretty {
		fmt.Fprint(b.writer, htmlEscaper.Replace(s(chars)))
	} else {
		fmt.Fprintf(b.writer, "%s%s%s\n", b.doIndent(), b.indent, htmlEscaper.Replace(s(chars)))
	}
	return b
}

// CharsNoEscape adds characters to the document without escaping special characters like <, & and >.
func (b *Builder) CharsNoEscape(chars interface{}) *Builder {
	b.outputElement(false, b.pretty && !b.inline)
	if b.inline || !b.pretty {
		fmt.Fprint(b.writer, s(chars))
	} else {
		fmt.Fprintf(b.writer, "%s%s%s\n", b.doIndent(), b.indent, s(chars))
	}
	return b
}

// Cdata adds a cdata element to the output. The cdata endtoken "]]> should not appear in the input string.
// This function does not check this.
func (b *Builder) Cdata(data interface{}) *Builder {
	b.outputElement(false, b.pretty)
	newline := "\n"
	if !b.pretty {
		newline = ""
	}
	fmt.Fprintf(b.writer, "%s%s<![CDATA[%s]]>%s", b.doIndent(), b.indent, s(data), newline)
	return b
}

func (b *Builder) doIndent() string {
	if !b.pretty { // pretty print is off, no indent
		return ""
	}
	indentValue := len(b.elements) + b.offset - 1
	if len(b.indentString) != len(b.indent)*indentValue {
		b.indentString = strings.Repeat(b.indent, indentValue)
	}
	return b.indentString
}

// Indent is used to set the indent string
func (b *Builder) Indent(indent string) *Builder {
	b.indent = indent
	return b
}

func (b *Builder) outputElement(close bool, newline bool) {
	if b.buildingElement {
		buf := &bytes.Buffer{}
		buf.WriteString(b.doIndent())
		buf.WriteRune('<')
		buf.WriteString(b.elements[len(b.elements)-1])
		for key, value := range b.attributes {
			if key != "" && value != "" {
				buf.WriteRune(' ')
				buf.WriteString(key)
				buf.WriteString(`="`)
				buf.WriteString(attrEscaper.Replace(value))
				buf.WriteString(`"`)
			}
		}
		b.attributes = make(map[string]string)
		if close {
			b.elements = b.elements[:len(b.elements)-1]
			buf.WriteString(" />")
		} else {
			buf.WriteRune('>')
		}
		if newline {
			buf.WriteString("\n")
		}
		b.writer.Write(buf.Bytes())
		b.buildingElement = false
	}
}

// Pretty is used to turn on and off the pretty printing of xml
func (b *Builder) Pretty(pretty bool) *Builder {
	b.pretty = pretty
	return b
}
