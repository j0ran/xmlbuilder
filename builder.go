package builder

import (
	"fmt"
	"io"
	"strings"
)

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#39;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&#34;",
)

type Builder struct {
	writer          io.Writer
	buildingElement bool
	attributes      map[string]string
	elements        []string
	indentString    string
	indent          string
	inline          bool
}

func s(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func New(writer io.Writer) *Builder {
	builder := new(Builder)
	builder.writer = writer
	builder.attributes = make(map[string]string)
	builder.indent = "  "
	return builder
}

func (b *Builder) Element(element string, args ...interface{}) *Builder {
	if b.buildingElement {
		b.outputElement(false, true)
	}

	b.buildingElement = true
	b.attributes = make(map[string]string)
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

func (b *Builder) Attr(name string, value interface{}) *Builder {
	if b.buildingElement {
		b.attributes[name] = s(value)
	}
	return b
}

func (b *Builder) End() *Builder {
	if b.buildingElement {
		b.outputElement(true, true)
	} else {
		if b.inline {
			fmt.Fprintf(b.writer, "</%s>\n", b.elements[len(b.elements)-1])
		} else {
			fmt.Fprintf(b.writer, "%s</%s>\n", b.doIndent(), b.elements[len(b.elements)-1])
		}
		b.elements = b.elements[:len(b.elements)-1]
	}
	return b
}

func (b *Builder) Tag(element string, args ...interface{}) *Builder {
	b.inline = true
	b.Element(element, args...).End()
	b.inline = false
	return b
}

func (b *Builder) Instruct(name string, args ...interface{}) *Builder {
	fmt.Fprintf(b.writer, "<?%s", name)
	for i := 0; i < len(args); i += 2 {
		fmt.Fprintf(b.writer, ` %s="%s"`, s(args[i+0]), htmlEscaper.Replace(s(args[i+1])))
	}
	fmt.Fprintln(b.writer, "?>")
	return b
}

func (b *Builder) InstructXML() *Builder {
	return b.Instruct("xml", "version", "1.0", "encoding", "UTF-8")
}

func (b *Builder) Chars(chars interface{}) *Builder {
	b.outputElement(false, !b.inline)
	if b.inline {
		fmt.Fprint(b.writer, htmlEscaper.Replace(s(chars)))
	} else {
		fmt.Fprintf(b.writer, "%s%s%s\n", b.doIndent(), b.indent, htmlEscaper.Replace(s(chars)))
	}
	return b
}

func (b *Builder) CharsNoEscape(chars interface{}) *Builder {
	b.outputElement(false, !b.inline)
	if b.inline {
		fmt.Fprint(b.writer, s(chars))
	} else {
		fmt.Fprintf(b.writer, "%s%s%s\n", b.doIndent(), b.indent, s(chars))
	}
	return b
}

func (b *Builder) Cdata(data interface{}) *Builder {
	b.outputElement(false, true)
	fmt.Fprintf(b.writer, "%s%s<![CDATA[%s]]>\n", b.doIndent(), b.indent, s(data))
	return b
}

func (b *Builder) Flush() {
	b.outputElement(true, true)
}

func (b *Builder) doIndent() string {
	indentValue := len(b.elements) - 1
	if len(b.indentString) != len(b.indent)*indentValue {
		b.indentString = strings.Repeat(b.indent, indentValue)
	}
	return b.indentString
}

func (b *Builder) outputElement(close bool, newline bool) {
	if b.buildingElement {
		fmt.Fprintf(b.writer, "%s<%s", b.doIndent(), b.elements[len(b.elements)-1])
		for key, value := range b.attributes {
			if key != "" && value != "" {
				fmt.Fprintf(b.writer, ` %s="%s"`, key, htmlEscaper.Replace(value))
			}
		}
		if close {
			b.elements = b.elements[:len(b.elements)-1]
			fmt.Fprint(b.writer, " />")
		} else {
			fmt.Fprint(b.writer, ">")
		}
		if newline {
			fmt.Fprintln(b.writer)
		}

		b.buildingElement = false
	}
}
