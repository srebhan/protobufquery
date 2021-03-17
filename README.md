protobufquery
<!-- ==== -->
<!-- [![Build Status](https://travis-ci.org/antchfx/protobufquery.svg?branch=master)](https://travis-ci.org/antchfx/protobufquery)
[![Coverage Status](https://coveralls.io/repos/github/antchfx/protobufquery/badge.svg?branch=master)](https://coveralls.io/github/antchfx/protobufquery?branch=master)
[![GoDoc](https://godoc.org/github.com/antchfx/protobufquery?status.svg)](https://godoc.org/github.com/antchfx/protobufquery)
[![Go Report Card](https://goreportcard.com/badge/github.com/antchfx/protobufquery)](https://goreportcard.com/report/github.com/antchfx/protobufquery) -->

Overview
===

Protobufquery is an XPath query package for ProtocolBuffer documents. It lets you extract data from parsed
ProtocolBuffer message through an XPath expression. Built-in XPath expression cache avoid re-compilation of
XPath expression for each query.

Getting Started
===

### Install Package
```
go get github.com/doclambda/protobufquery
```

#### Load ProtocolBuffer message from io.Reader.

**TODO**

```go
f, err := os.Open("./books.json")
doc, err := protobufquery.Parse(f)
```

#### Find authors of all books in the store.
```go
list := protobufquery.Find(doc, "store/book/*/author")
// or equal to
list := protobufquery.Find(doc, "//author")
// or by QueryAll()
nodes, err := protobufquery.QueryAll(doc, "//a")
```

#### Find the third book.

```go
book := protobufquery.Find(doc, "//book/*[3]")
```

#### Find the last book.

```go
book := protobufquery.Find(doc, "//book/*[last()]")
```

#### Find all books that have an isbn number.

```go
list := protobufquery.Find(doc, "//book/*[isbn]")
```

#### Find all books priced less than 10.

```go
list := protobufquery.Find(doc, "//book/*[price<10]")
```

Examples
===

**TODO**

```go
func main() {
	s := `{
		"name": "John",
		"age"      : 26,
		"address"  : {
		  "streetAddress": "naist street",
		  "city"         : "Nara",
		  "postalCode"   : "630-0192"
		},
		"phoneNumbers": [
		  {
			"type"  : "iPhone",
			"number": "0123-4567-8888"
		  },
		  {
			"type"  : "home",
			"number": "0123-4567-8910"
		  }
		]
	}`
	doc, err := protobufquery.Parse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	name := protobufquery.FindOne(doc, "name")
	fmt.Printf("name: %s\n", name.InnerText())
	var a []string
	for _, n := range protobufquery.Find(doc, "phoneNumbers/*/number") {
		a = append(a, n.InnerText())
	}
	fmt.Printf("phone number: %s\n", strings.Join(a, ","))
	if n := protobufquery.FindOne(doc, "address/streetAddress"); n != nil {
		fmt.Printf("address: %s\n", n.InnerText())
	}
}
```

Implement Principle
===
If you are familiar with XPath and XML, you can easily figure out how to
write your XPath expression.

```json
{
"name":"John",
"age":30,
"cars": [
	{ "name":"Ford", "models":[ "Fiesta", "Focus", "Mustang" ] },
	{ "name":"BMW", "models":[ "320", "X3", "X5" ] },
	{ "name":"Fiat", "models":[ "500", "Panda" ] }
]
}
```
The above ProtocolBuffer above will be convert by *protobufquery* to a structure similar to the XML document below:

```XML
<name>John</name>
<age>30</age>
<cars>
	<element>
		<name>Ford</name>
		<models>
			<element>Fiesta</element>
			<element>Focus</element>
			<element>Mustang</element>
		</models>		
	</element>
	<element>
		<name>BMW</name>
		<models>
			<element>320</element>
			<element>X3</element>
			<element>X5</element>
		</models>		
	</element>
	<element>
		<name>Fiat</name>
		<models>
			<element>500</element>
			<element>Panda</element>
		</models>		
	</element>
</cars>
```

Note: `element` is an anonymous element without name.

List of XPath query packages
===
|Name |Description |
|--------------------------|----------------|
|[htmlquery](https://github.com/antchfx/htmlquery) | XPath query package for the HTML document|
|[xmlquery](https://github.com/antchfx/xmlquery) | XPath query package for the XML document|
|[jsonquery](https://github.com/antchfx/jsonquery) | XPath query package for the JSON document|
|[protobufquery](https://github.com/doclambda/protobufquery) | XPath query package for ProtocolBuffer messages|
