# xmlbuilder
Go library to generate xml output using the builder pattern.

## Installing

```
go get -u github.com/j0ran/xmlbuilder
```

## Example

```golang
func main() {
	xml := xmlbuilder.New(os.Stdout)
	xml.InstructXML()
	xml.Element("root", "id", 1, "escape", "<escape \"this\">")
	{
		xml.Element("count")
		{
			for i := 1; i <= 3; i++ {
				xml.Attr(fmt.Sprintf("attr%d", i), i*i)
			}
			for i := 1; i < 10; i++ {
				xml.Tag("item", "position", i, "test", true)
			}
		}
		xml.End()

		xml.Chars("Some <mixed> content")

		xml.Element("test", "Mixed content", "id", 20, "extra", "extra attribute")
		{
			xml.Cdata("A cdata block")
			xml.Tag("tag", "Me")
		}
		xml.End()
	}
	xml.End()
}
```

Will produce:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<root escape="&lt;escape &#34;this&#34;&gt;" id="1">
  <count attr1="1" attr2="4" attr3="9">
    <item position="1" test="true" />
    <item position="2" test="true" />
    <item test="true" position="3" />
    <item position="4" test="true" />
    <item position="5" test="true" />
    <item position="6" test="true" />
    <item position="7" test="true" />
    <item position="8" test="true" />
    <item position="9" test="true" />
  </count>
  Some &lt;mixed&gt; content
  <test id="20" extra="extra attribute">
    Mixed content
    <![CDATA[A cdata block]]>
    <tag>Me</tag>
  </test>
</root>
```

# Documentation

Not much documentation yet. 

The `Tag` method is basicly a call to `Element` and then `End`, but it will put the output on a single line when pretty printing is enabled (which is the default).

For each call to `Element` you must make a matching call to `End`.

The library doesn't catch errors when you call the methods in the wrong order. It does escape text when needed for the Chars and attributes, but it doesn't check the contents for the names of the elements and attributes.

There is no explicit namespace support, but you can generate xml that uses namespaces by hardcoding it in. For example:

```golang
xml.Element("gml:Point",
		"gml:id", "p21",
		"srsName", "http://www.opengis.net/def/crs/EPSG/0/4326",
		"xmlns:gml", "http://www.opengis.net/gml")
{
	xml.Tag("gml:coordinates", 45.67, 88.56)
}
xml.End()
```
