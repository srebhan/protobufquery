package protobufquery

import (
	"testing"

	"github.com/antchfx/xpath"

	"github.com/stretchr/testify/require"
)

type keyValue struct {
	key string
	value string
	missing bool
}

type keyValueList []keyValue

func BenchmarkSelectorCache(b *testing.B) {
	DisableSelectorCache = false
	for i := 0; i < b.N; i++ {
		getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func BenchmarkDisableSelectorCache(b *testing.B) {
	DisableSelectorCache = true
	for i := 0; i < b.N; i++ {
		getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func TestNavigator(t *testing.T) {
	msg := addressbookSample.ProtoReflect()
	doc, err := Parse(msg)
	require.NoError(t, err)

	nav := CreateXPathNavigator(doc)
	nav.MoveToRoot()
	if nav.NodeType() != xpath.RootNode {
		t.Fatal("node type is not RootNode")
	}

	expectedPeople := []keyValueList {
		[]keyValue {
			{key: "name",  value: "John Doe"},
			{key: "id",    value: "101"},
			{key: "email", value: "john@example.com"},
		},
		[]keyValue {
			{key: "name",  value: "Jane Doe"},
			{key: "id",    value: "102"},
			{key: "email", value: "", missing: true},
		},
	}
	require.True(t, nav.MoveToChild())
	for _, keyvalues := range expectedPeople {
		require.Equal(t, "people", nav.Current().Name)
		require.True(t, nav.MoveToChild())
		for i, v := range keyvalues {
			if !v.missing {
				if i > 0 {
					require.True(t, nav.MoveToNext())
				}
				require.Equal(t, v.key, nav.Current().Name)
				require.Equal(t, v.value, nav.Value())
			} else {
				if i > 0 {
					if ! nav.MoveToNext() {
						// There is no other node so we passed the test
						continue
					}
				}
				require.NotEqual(t, v.key, nav.Current().Name)
				if i > 0 {
					require.True(t, nav.MoveToPrevious())
				}
			}
		}
		require.True(t, nav.MoveToParent())
		require.True(t, nav.MoveToNext())
	}

	require.True(t, nav.MoveToParent())
	require.Equal(t, nav.Current().Type, DocumentNode, "expected 'DocumentNode'")
}

func TestQuery(t *testing.T) {
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
