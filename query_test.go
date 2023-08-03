package protobufquery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antchfx/xpath"
	"github.com/stretchr/testify/require"
)

func BenchmarkSelectorCache(b *testing.B) {
	DisableSelectorCache = false
	for i := 0; i < b.N; i++ {
		_, _ = getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func BenchmarkDisableSelectorCache(b *testing.B) {
	DisableSelectorCache = true
	for i := 0; i < b.N; i++ {
		_, _ = getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func TestNavigator(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	nav := CreateXPathNavigator(doc)
	nav.MoveToRoot()
	require.Equal(t, xpath.RootNode, nav.NodeType(), "node type is not RootNode")

	expectedPeoples := []map[string]interface{}{
		{
			"name":  "John Doe",
			"id":    "101",
			"email": "john@example.com",
			"age":   "42",
		},
		{
			"name": "Jane Doe",
			"id":   "102",
			"age":  "40",
		},
	}

	// Find the people node
	require.True(t, nav.MoveToChild())
	for nav.Current().Name != "people" && nav.MoveToNext() {
		// nothing to do
	}

	// Check the nodes against the expected values and check if we did not
	// miss any field
	for _, expected := range expectedPeoples {
		require.Equal(t, "people", nav.Current().Name)
		require.True(t, nav.MoveToChild())
		seen := make(map[string]bool, len(expected))
		for {
			// Compare the node value
			key := nav.Current().Name
			require.NotEmpty(t, key)
			expectedValue, found := expected[key]
			require.Truef(t, found, "key %q not found", key)
			require.Equal(t, expectedValue, nav.Value())

			// Remember which fields we saw for later reverse check
			seen[key] = true

			if !nav.MoveToNext() {
				break
			}
		}

		// Check if we miss an expected field
		for key := range expected {
			require.Truef(t, seen[key], "key %q not found in nodes", key)
		}
		require.True(t, nav.MoveToParent())
		require.True(t, nav.MoveToNext())
	}

	require.True(t, nav.MoveToParent())
	require.Equal(t, nav.Current().Type, DocumentNode, "expected 'DocumentNode'")
}

func TestQueryNames(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	expected := []string{
		"John Doe",
		"Jane Doe",
		"Jack Doe",
		"Jack Buck",
		"Janet Doe",
	}

	nodes, err := QueryAll(doc, "/descendant::*[name() = 'people']/name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}

	nodes, err = QueryAll(doc, "//name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}

	nodes, err = QueryAll(doc, "/people[3]/name")
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, "name", nodes[0].Name)
	require.EqualValues(t, expected[2], nodes[0].Value())
}

func TestQueryPhoneNumberFirst(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	expected := []string{
		"555-555-5555",
		"555-555-0000",
		"555-777-0000",
	}
	nodes, err := QueryAll(doc, "//phones[1]/number")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "number", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryPhoneNumberLast(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	expected := []string{
		"555-555-5555",
		"555-555-0002",
		"555-777-0001",
	}
	nodes, err := QueryAll(doc, "//phones[last()]/number")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "number", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryPhoneNoEmail(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	nodes, err := QueryAll(doc, "/people[not(email)]/id")
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, "id", nodes[0].Name)
	require.EqualValues(t, 102, nodes[0].Value())
}

func TestQueryPhoneAge(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	expected := []string{
		"John Doe",
		"Jane Doe",
		"Jack Buck",
	}
	nodes, err := QueryAll(doc, "/people[age > 18]/name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryJack(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	nodes, err := QueryAll(doc, "//people[contains(name, 'Jack')]/id")
	require.NoError(t, err)

	require.Len(t, nodes, 2)
	require.Equal(t, "id", nodes[0].Name)
	require.EqualValues(t, 201, nodes[0].Value())
	require.Equal(t, "id", nodes[1].Name)
	require.EqualValues(t, 301, nodes[1].Value())
}

func TestQueryExample(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	expected := []string{
		"John Doe: ",
		"Jane Doe: ",
		"Jack Doe: 555-555-5555",
		"Jack Buck: 555-555-0000,555-555-0001,555-555-0002",
		"Janet Doe: 555-777-0000,555-777-0001",
	}

	nodes, err := QueryAll(doc, "//people")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))

	for i, person := range nodes {
		name := FindOne(person, "name").InnerText()
		numbers := make([]string, 0)
		for _, node := range Find(person, "phones/number") {
			numbers = append(numbers, node.InnerText())
		}
		v := fmt.Sprintf("%s: %s", name, strings.Join(numbers, ","))
		require.EqualValues(t, expected[i], v)
	}
}
