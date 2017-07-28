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

type blank struct{}

var Blank = blank{}

func (b blank) String() string {
	return ""
}

type Builder struct {
	writer          io.Writer
	buildingElement bool
	attributes      map[string]string
	attrnames       []string
	elements        []string
	indentString    string
	indent          string
	offset          int  // indent offset
	empty           bool // use empty elements
	inline          int
}

func s(v ...interface{}) string {
	return fmt.Sprint(v...)
}

// New creates a new xml builder. By default pretty print is enabled and the identation is
// two spaces.
func New(writer io.Writer) *Builder {
	builder := new(Builder)
	builder.writer = writer
	builder.attributes = make(map[string]string)
	builder.indent = "  "
	builder.empty = true
	return builder
}

func (b *Builder) Inline() *Builder {
	if b.inline == 0 {
		b.Flush()
		b.offset++
		fmt.Fprint(b.writer, b.doIndent())
		b.offset--
	}
	b.inline++
	return b
}

func (b *Builder) EndInline() *Builder {
	b.inline--
	if b.inline == 0 {
		fmt.Fprintln(b.writer)
	}
	return b
}

// Element defines a new element in the xml document.
func (b *Builder) Element(element string, args ...interface{}) *Builder {
	if b.buildingElement {
		b.outputElement(false)
	}

	b.buildingElement = true
	b.elements = append(b.elements, element)
	first := len(args) % 2
	for i := first; i < len(args); i += 2 {
		if a, v := s(args[i+0]), s(args[i+1]); a != "" && (v != "" || args[i+1] == Blank) {
			b.Attr(a, v)
		}
	}
	if first != 0 {
		b.Chars(args[0])
	}

	return b
}

// ElementNoEscape defines a new element in the xml document but doesn't escape the Chars
func (b *Builder) ElementNoEscape(element string, args ...interface{}) *Builder {
	if b.buildingElement {
		b.outputElement(false)
	}

	b.buildingElement = true
	b.elements = append(b.elements, element)
	first := len(args) % 2
	for i := first; i < len(args); i += 2 {
		if a, v := s(args[i+0]), s(args[i+1]); a != "" && (v != "" || args[i+1] == Blank) {
			b.Attr(a, v)
		}
	}
	if first != 0 {
		b.CharsNoEscape(args[0])
	}

	return b
}

// Attr will add an attribute to the current element being build, or when not building
// an element it will add attributes to the next element to be build.
func (b *Builder) Attr(name string, value interface{}) *Builder {
	if _, found := b.attributes[name]; !found {
		b.attrnames = append(b.attrnames, name)
	}
	b.attributes[name] = s(value)

	return b
}

// End will add a close tag that matches the the previous Element call.
func (b *Builder) End() *Builder {
	if b.buildingElement {
		b.outputElement(true)
	} else {
		if b.inline > 0 {
			fmt.Fprint(b.writer, "</", b.elements[len(b.elements)-1], ">")
		} else {
			fmt.Fprint(b.writer, b.doIndent(), "</", b.elements[len(b.elements)-1], ">\n")
		}
		b.elements = b.elements[:len(b.elements)-1]
	}
	return b
}

// Tag inserts an inline element and directly closes it.
func (b *Builder) Tag(element string, args ...interface{}) *Builder {
	b.Inline()
	b.Element(element, args...).End()
	b.EndInline()
	return b
}

// TagNoEscape inserts an inline element and directly closes it but doesn't escape the Chars.
func (b *Builder) TagNoEscape(element string, args ...interface{}) *Builder {
	b.Inline()
	b.ElementNoEscape(element, args...).End()
	b.EndInline()
	return b
}

// Instruct can be used to generate instruction tags.
func (b *Builder) Instruct(name string, args ...interface{}) *Builder {
	fmt.Fprint(b.writer, "<?", name)
	for i := 0; i < len(args); i += 2 {
		fmt.Fprint(b.writer, " ", args[i+0], `="`, attrEscaper.Replace(s(args[i+1])), `"`)
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
	fmt.Fprint(b.writer, "<!DOCTYPE ", doctype, ">\n")
	return b
}

// Offset will add the delta value to the given ident offset
func (b *Builder) Offset(delta int) *Builder {
	b.offset += delta
	return b
}

// Chars add characters to the document. It will also escape special characters.
func (b *Builder) Chars(chars ...interface{}) *Builder {
	b.outputElement(false)
	line := fmt.Sprint(chars...)
	if b.inline > 0 {
		fmt.Fprint(b.writer, htmlEscaper.Replace(line))
	} else {
		fmt.Fprint(b.writer, b.doIndent(), b.indent, htmlEscaper.Replace(line), "\n")
	}
	return b
}

// CharsNoEscape adds characters to the document without escaping special characters like <, & and >.
func (b *Builder) CharsNoEscape(chars ...interface{}) *Builder {
	b.outputElement(false)
	line := fmt.Sprint(chars...)
	if b.inline > 0 {
		fmt.Fprint(b.writer, line)
	} else {
		fmt.Fprint(b.writer, b.doIndent(), b.indent, line, "\n")
	}
	return b
}

// Cdata adds a cdata element to the output. The cdata endtoken "]]> should not appear in the input string.
// This function does not check this.
func (b *Builder) Cdata(data ...interface{}) *Builder {
	b.outputElement(false)
	line := fmt.Sprint(data...)
	if b.inline > 0 {
		fmt.Fprint(b.writer, "<![CDATA[", line, "]]>")
	} else {
		fmt.Fprint(b.writer, b.doIndent(), b.indent, "<![CDATA[", line, "]]>\n")
	}
	return b
}

func (b *Builder) doIndent() string {
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

// Empty is used to determine if empty elements should be used
// If true an empty element will be outputed as <br />, the default
// If false an empty element will be outputed as <br>
func (b *Builder) Empty(useEmpty bool) *Builder {
	b.empty = useEmpty
	return b
}

func (b *Builder) outputElement(close bool) {
	if b.buildingElement {
		buf := &bytes.Buffer{}
		if b.inline == 0 {
			buf.WriteString(b.doIndent())
		}
		buf.WriteRune('<')
		buf.WriteString(b.elements[len(b.elements)-1])
		for _, key := range b.attrnames {
			value := b.attributes[key]
			buf.WriteRune(' ')
			buf.WriteString(key)
			buf.WriteString(`="`)
			buf.WriteString(attrEscaper.Replace(value))
			buf.WriteString(`"`)
		}
		b.attributes = make(map[string]string)
		b.attrnames = b.attrnames[:0]
		if close {
			b.elements = b.elements[:len(b.elements)-1]
			if b.empty {
				buf.WriteString(" />")
			} else {
				buf.WriteRune('>')
			}
		} else {
			buf.WriteRune('>')
		}
		if b.inline == 0 {
			buf.WriteString("\n")
		}
		b.writer.Write(buf.Bytes())
		b.buildingElement = false
	}
}

// Flush will generate the current element that is being build.
func (b *Builder) Flush() *Builder {
	b.outputElement(false)
	return b
}
