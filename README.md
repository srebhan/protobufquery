protobufquery
====
[![Build Status](https://travis-ci.org/antchfx/protobufquery.svg?branch=master)](https://travis-ci.org/antchfx/protobufquery)
[![Coverage Status](https://coveralls.io/repos/github/antchfx/protobufquery/badge.svg?branch=master)](https://coveralls.io/github/antchfx/protobufquery?branch=master)
[![GoDoc](https://godoc.org/github.com/antchfx/protobufquery?status.svg)](https://godoc.org/github.com/antchfx/protobufquery)
[![Go Report Card](https://goreportcard.com/badge/github.com/antchfx/protobufquery)](https://goreportcard.com/report/github.com/antchfx/protobufquery)

Overview
===

protobufquery is an XPath query package for JSON document, lets you extract data from JSON documents through an XPath expression. Built-in XPath expression cache avoid re-compile XPath expression each query.

Getting Started
===

### Install Package
```
go get github.com/antchfx/protobufquery
```

#### Load JSON document from URL.

```go
doc, err := protobufquery.LoadURL("http://www.example.com/feed?json")
```

#### Load JSON document from string.

```go
s :=`{
    "name":"John",
    "age":31, 
    "city":"New York" 
    }`
doc, err := protobufquery.Parse(strings.NewReader(s))
```

#### Load JSON document from io.Reader.

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
The above JSON document will be convert to similar to XML document by the *protobufquery*, like below:

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

Notes: `element` is empty element that have no any name.

List of XPath query packages
===
|Name |Description |
|--------------------------|----------------|
|[htmlquery](https://github.com/antchfx/htmlquery) | XPath query package for the HTML document|
|[xmlquery](https://github.com/antchfx/xmlquery) | XPath query package for the XML document|
|[protobufquery](https://github.com/antchfx/protobufquery) | XPath query package for the JSON document|
