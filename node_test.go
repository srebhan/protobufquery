package protobufquery

import (
	"testing"

	"github.com/doclambda/protobufquery/testcases/addressbook"
	"github.com/stretchr/testify/require"
)

var addressbookSample = &addressbook.AddressBook{
	People: []*addressbook.Person {
		{
			Name:  "John Doe",
			Id:    101,
			Email: "john@example.com",
		},
		{
			Name: "Jane Doe",
			Id:   102,
		},
		{
			Name:  "Jack Doe",
			Id:    201,
			Email: "jack@example.com",
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-555-5555", Type: addressbook.Person_WORK},
			},
		},
		{
			Name:  "Jack Buck",
			Id:    301,
			Email: "buck@example.com",
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-555-0000", Type: addressbook.Person_HOME},
				{Number: "555-555-0001", Type: addressbook.Person_MOBILE},
				{Number: "555-555-0002", Type: addressbook.Person_WORK},
			},
		},
		{
			Name:  "Janet Doe",
			Id:    1001,
			Email: "janet@example.com",
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-777-0000"},
				{Number: "555-777-0001", Type: addressbook.Person_HOME},
			},
		},
	},
	Tags: []string {"home", "private", "friends"},
}


func TestParseAddressBookXML(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)
	require.Len(t, doc.ChildNodes(), 6)

	xml := doc.OutputXML()
	expected := `<?xml version="1.0"?><people><name>John Doe</name><id>101</id><email>john@example.com</email></people><people><name>Jane Doe</name><id>102</id></people><people><name>Jack Doe</name><id>201</id><email>jack@example.com</email><phones><number>555-555-5555</number><type>2</type></phones></people><people><name>Jack Buck</name><id>301</id><email>buck@example.com</email><phones><number>555-555-0000</number><type>1</type></phones><phones><number>555-555-0001</number></phones><phones><number>555-555-0002</number><type>2</type></phones></people><people><name>Janet Doe</name><id>1001</id><email>janet@example.com</email><phones><number>555-777-0000</number></phones><phones><number>555-777-0001</number><type>1</type></phones></people><tags><element>home</element><element>private</element><element>friends</element></tags>`
	require.Equal(t, expected, xml)
}
