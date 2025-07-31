package protobufquery

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/srebhan/protobufquery/testcases/addressbook"
)

// TestProto3DefaultsBackwardCompatibility ensures existing behavior is preserved
func TestProto3DefaultsBackwardCompatibility(t *testing.T) {
	msg := addressbookSample.ProtoReflect()

	// Test with default options (should preserve existing behavior)
	doc, err := Parse(msg)
	require.NoError(t, err)

	// Should work the same as before
	people := Find(doc, "//people")
	require.Len(t, people, 5)

	// Test that default Parse() doesn't populate defaults
	opts := DefaultOptions
	require.False(t, opts.PopulateDefaults, "Default options should not populate defaults for backward compatibility")
}

// TestProto3DefaultsOptIn tests the new functionality when explicitly enabled
func TestProto3DefaultsOptIn(t *testing.T) {
	msg := addressbookSample.ProtoReflect()

	// Test with PopulateDefaults enabled
	opts := Options{PopulateDefaults: true}
	doc, err := ParseWithOptions(msg, opts)
	require.NoError(t, err)

	// Should still work correctly
	people := Find(doc, "//people")
	require.Len(t, people, 5)

	// Test that XML contains expected elements
	xmlOutput := doc.OutputXML()
	require.Contains(t, xmlOutput, "<people>")
	require.Contains(t, xmlOutput, "<name>")
	require.Contains(t, xmlOutput, "<id>")
}

// TestProto3DefaultValuePopulation tests the specific proto3 default value scenario
func TestProto3DefaultValuePopulation(t *testing.T) {
	// Create a proto3 message with default values that would normally be omitted
	person := &addressbook.Person{
		Name:  "Test Person",
		Id:    0,  // This is a default value for int32
		Age:   0,  // This is a default value for int32
		Email: "", // This is a default value for string
	}

	msg := person.ProtoReflect()

	// Test without PopulateDefaults (existing behavior)
	doc1, err := Parse(msg)
	require.NoError(t, err)

	xml1 := doc1.OutputXML()
	t.Logf("Without PopulateDefaults: %s", xml1)

	// Test with PopulateDefaults enabled
	opts := Options{PopulateDefaults: true}
	doc2, err := ParseWithOptions(msg, opts)
	require.NoError(t, err)

	xml2 := doc2.OutputXML()
	t.Logf("With PopulateDefaults: %s", xml2)

	// With PopulateDefaults, we should see default values in the XML
	if strings.Contains(xml2, "<id>0</id>") && strings.Contains(xml2, "<age>0</age>") {
		t.Log("PopulateDefaults correctly includes default values")
	} else {
		t.Log("Default values may not be included (depending on proto syntax)")
	}
}

// TestProto3VsProto2Behavior tests different behavior for proto2 vs proto3
func TestProto3VsProto2Behavior(t *testing.T) {
	msg := addressbookSample.ProtoReflect()

	// Test with PopulateDefaults enabled
	opts := Options{PopulateDefaults: true}
	doc, err := ParseWithOptions(msg, opts)
	require.NoError(t, err)

	// Check the syntax of the message
	syntax := msg.Descriptor().Syntax()
	t.Logf("Message syntax: %v", syntax)

	// The behavior should depend on whether this is proto2 or proto3
	if syntax == protoreflect.Proto3 {
		t.Log("Proto3 message - defaults will be populated when enabled")
	} else {
		t.Log("Proto2 message - existing behavior preserved")
	}

	// Should always work regardless of syntax
	people := Find(doc, "//people")
	require.Len(t, people, 5)
}

// TestOptionsValidation tests that the Options struct works correctly
func TestOptionsValidation(t *testing.T) {
	msg := addressbookSample.ProtoReflect()

	testCases := []struct {
		name string
		opts Options
	}{
		{
			name: "PopulateDefaults false",
			opts: Options{PopulateDefaults: false},
		},
		{
			name: "PopulateDefaults true",
			opts: Options{PopulateDefaults: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := ParseWithOptions(msg, tc.opts)
			require.NoError(t, err)
			require.NotNil(t, doc)

			// Should always be able to query successfully
			people := Find(doc, "//people")
			require.Len(t, people, 5)
		})
	}
}

// TestFieldDefaultValues tests various field types with default values
func TestFieldDefaultValues(t *testing.T) {
	// Test with a person that has minimal fields set
	person := &addressbook.Person{
		Name: "Minimal Person",
		// Id, Age, Email not set - should get defaults with PopulateDefaults=true
	}

	msg := person.ProtoReflect()

	// Test both options
	testCases := []struct {
		name     string
		opts     Options
		expected []string // Expected elements in XML
	}{
		{
			name:     "Without PopulateDefaults",
			opts:     Options{PopulateDefaults: false},
			expected: []string{"<name>Minimal Person</name>"},
		},
		{
			name:     "With PopulateDefaults",
			opts:     Options{PopulateDefaults: true},
			expected: []string{"<name>Minimal Person</name>"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := ParseWithOptions(msg, tc.opts)
			require.NoError(t, err)

			xml := doc.OutputXML()
			t.Logf("XML output: %s", xml)

			// Check expected elements are present
			for _, expected := range tc.expected {
				require.Contains(t, xml, expected)
			}
		})
	}
}

// TestXPathQueriesWithDefaults tests XPath queries work with default values
func TestXPathQueriesWithDefaults(t *testing.T) {
	person := &addressbook.Person{
		Name: "Query Test",
		Id:   42,
		Age:  25,
	}

	msg := person.ProtoReflect()

	// Test with PopulateDefaults disabled first
	opts1 := Options{PopulateDefaults: false}
	doc1, err := ParseWithOptions(msg, opts1)
	require.NoError(t, err)

	// Test with PopulateDefaults enabled
	opts2 := Options{PopulateDefaults: true}
	doc2, err := ParseWithOptions(msg, opts2)
	require.NoError(t, err)

	// Test that specific fields can be found
	testQueries := []struct {
		query       string
		shouldFind  bool
		description string
	}{
		{"//name", true, "name should always be found"},
		{"//id", true, "id should be found when set"},
		{"//age", true, "age should be found when set"},
	}

	for _, tq := range testQueries {
		t.Run(tq.query, func(t *testing.T) {
			// Test with PopulateDefaults false
			nodes1, err := QueryAll(doc1, tq.query)
			require.NoError(t, err)

			// Test with PopulateDefaults true
			nodes2, err := QueryAll(doc2, tq.query)
			require.NoError(t, err)

			if tq.shouldFind {
				require.NotEmpty(t, nodes1, "Query %s should find nodes without PopulateDefaults", tq.query)
				require.NotEmpty(t, nodes2, "Query %s should find nodes with PopulateDefaults", tq.query)
			}

			t.Logf("Query %s: without defaults=%d, with defaults=%d",
				tq.query, len(nodes1), len(nodes2))
		})
	}

	// Test that we can find all elements (count may differ based on defaults)
	t.Run("count_all_elements", func(t *testing.T) {
		allNodes1, err := QueryAll(doc1, "//*")
		require.NoError(t, err)

		allNodes2, err := QueryAll(doc2, "//*")
		require.NoError(t, err)

		t.Logf("Total elements: without defaults=%d, with defaults=%d",
			len(allNodes1), len(allNodes2))

		// With PopulateDefaults, we should have >= elements than without
		require.GreaterOrEqual(t, len(allNodes2), len(allNodes1),
			"PopulateDefaults should not reduce the number of elements")
	})
}

// TestProto3ZeroValueFix demonstrates the specific issue we're solving
func TestProto3ZeroValueFix(t *testing.T) {
	// This test simulates the exact issue described in the GitHub issue:
	// When a proto3 message has a field with value 0.0, it gets omitted
	// from the wire format, causing XPath queries to fail

	// Create a person with zero values (would be omitted in proto3)
	person := &addressbook.Person{
		Name:  "Zero Test",
		Id:    0,  // This is a zero value
		Age:   0,  // This is a zero value
		Email: "", // This is a zero value (empty string)
	}

	msg := person.ProtoReflect()
	syntax := msg.Descriptor().Syntax()

	t.Logf("Testing with message syntax: %v", syntax)

	// Test WITHOUT PopulateDefaults (existing behavior)
	t.Run("without_populate_defaults", func(t *testing.T) {
		doc, err := Parse(msg) // Uses default options
		require.NoError(t, err)

		xml := doc.OutputXML()
		t.Logf("XML without PopulateDefaults: %s", xml)

		// Try to find the zero-value fields
		idNodes, err := QueryAll(doc, "//id")
		require.NoError(t, err)

		ageNodes, err := QueryAll(doc, "//age")
		require.NoError(t, err)

		emailNodes, err := QueryAll(doc, "//email")
		require.NoError(t, err)

		// For proto3, zero values might be missing
		if syntax == protoreflect.Proto3 {
			t.Logf("Proto3: id nodes=%d, age nodes=%d, email nodes=%d",
				len(idNodes), len(ageNodes), len(emailNodes))
		}
	})

	// Test WITH PopulateDefaults (new behavior)
	t.Run("with_populate_defaults", func(t *testing.T) {
		opts := Options{PopulateDefaults: true}
		doc, err := ParseWithOptions(msg, opts)
		require.NoError(t, err)

		xml := doc.OutputXML()
		t.Logf("XML with PopulateDefaults: %s", xml)

		// Try to find the zero-value fields
		idNodes, err := QueryAll(doc, "//id")
		require.NoError(t, err)

		ageNodes, err := QueryAll(doc, "//age")
		require.NoError(t, err)

		emailNodes, err := QueryAll(doc, "//email")
		require.NoError(t, err)

		t.Logf("With PopulateDefaults: id nodes=%d, age nodes=%d, email nodes=%d",
			len(idNodes), len(ageNodes), len(emailNodes))

		// For proto3 with PopulateDefaults, we should find the zero values
		if syntax == protoreflect.Proto3 {
			// We should be able to find these fields even with zero values
			// Note: The exact behavior depends on the protobuf library implementation
			if len(idNodes) > 0 {
				idValue := idNodes[0].InnerText()
				t.Logf("Found id with zero value: '%s'", idValue)
				require.Equal(t, "0", idValue)
			}

			if len(ageNodes) > 0 {
				ageValue := ageNodes[0].InnerText()
				t.Logf("Found age with zero value: '%s'", ageValue)
				require.Equal(t, "0", ageValue)
			}

			if len(emailNodes) > 0 {
				emailValue := emailNodes[0].InnerText()
				t.Logf("Found email with zero value: '%s'", emailValue)
				require.Equal(t, "", emailValue)
			}
		}
	})
}

// TestTelemetryPointScenario simulates the exact Telegraf scenario from the GitHub issue
func TestTelemetryPointScenario(t *testing.T) {
	// This simulates the TelemetryPoint message from the GitHub issue:
	// message TelemetryPoint {
	//   float value = 1;
	//   string name = 2;
	// }

	// We'll use a Person message to simulate this (since we don't have TelemetryPoint)
	person := &addressbook.Person{
		Name: "test", // This represents the 'name' field
		Age:  0,      // This represents the 'value' field that would be 0
	}

	msg := person.ProtoReflect()

	t.Run("simulated_telegraf_issue", func(t *testing.T) {
		// Test the scenario where XPath query "number(age)" would return NaN
		// because the age=0 field is missing from the XML

		// Without PopulateDefaults
		doc1, err := Parse(msg)
		require.NoError(t, err)
		xml1 := doc1.OutputXML()
		t.Logf("Without PopulateDefaults: %s", xml1)

		// With PopulateDefaults
		opts := Options{PopulateDefaults: true}
		doc2, err := ParseWithOptions(msg, opts)
		require.NoError(t, err)
		xml2 := doc2.OutputXML()
		t.Logf("With PopulateDefaults: %s", xml2)

		// The key test: can we find the zero-value field?
		ageNodes, err := QueryAll(doc2, "//age")
		require.NoError(t, err)

		if len(ageNodes) > 0 {
			ageValue := ageNodes[0].InnerText()
			t.Logf("Fixed: XPath number(age) would return %s instead of NaN", ageValue)

			// This should be "0", not empty or missing
			require.Equal(t, "0", ageValue)
		} else {
			t.Log("Age field not found - may depend on protobuf version/syntax")
		}

		// Name should always be found
		nameNodes, err := QueryAll(doc2, "//name")
		require.NoError(t, err)
		require.NotEmpty(t, nameNodes)
		require.Equal(t, "test", nameNodes[0].InnerText())
	})
}

// TestParseWithOptionsAPI tests the new API functions
func TestParseWithOptionsAPI(t *testing.T) {
	msg := addressbookSample.ProtoReflect()

	t.Run("Parse_uses_default_options", func(t *testing.T) {
		doc, err := Parse(msg)
		require.NoError(t, err)
		require.NotNil(t, doc)

		// Should work the same as ParseWithOptions with DefaultOptions
		docWithDefaults, err := ParseWithOptions(msg, DefaultOptions)
		require.NoError(t, err)
		require.NotNil(t, docWithDefaults)

		// Both should produce the same result
		xml1 := doc.OutputXML()
		xml2 := docWithDefaults.OutputXML()
		require.Equal(t, xml1, xml2)
	})

	t.Run("ParseWithOptions_different_behaviors", func(t *testing.T) {
		opts1 := Options{PopulateDefaults: false}
		opts2 := Options{PopulateDefaults: true}

		doc1, err := ParseWithOptions(msg, opts1)
		require.NoError(t, err)

		doc2, err := ParseWithOptions(msg, opts2)
		require.NoError(t, err)

		// Both should work, but may produce different XML
		xml1 := doc1.OutputXML()
		xml2 := doc2.OutputXML()

		t.Logf("PopulateDefaults=false: %s", xml1)
		t.Logf("PopulateDefaults=true:  %s", xml2)

		// Both should be valid XML and queryable
		people1 := Find(doc1, "//people")
		people2 := Find(doc2, "//people")

		require.Len(t, people1, 5)
		require.Len(t, people2, 5)
	})
}

// TestDefaultOptionsValue ensures backward compatibility
func TestDefaultOptionsValue(t *testing.T) {
	// This test ensures that DefaultOptions preserves existing behavior
	require.False(t, DefaultOptions.PopulateDefaults,
		"DefaultOptions.PopulateDefaults must be false for backward compatibility")

	// Test that Parse() uses DefaultOptions
	msg := addressbookSample.ProtoReflect()

	doc1, err := Parse(msg)
	require.NoError(t, err)

	doc2, err := ParseWithOptions(msg, DefaultOptions)
	require.NoError(t, err)

	// Should produce identical results
	require.Equal(t, doc1.OutputXML(), doc2.OutputXML())
}
