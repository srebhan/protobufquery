package protobufquery

import (
	"bytes"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// A NodeType is the type of a Node.
type NodeType uint

const (
	// DocumentNode is a document object that, as the root of the document tree,
	// provides access to the entire XML document.
	DocumentNode NodeType = iota
	// ElementNode is an element.
	ElementNode
	// TextNode is the text content of a node.
	TextNode
)

// A Node consists of a NodeType and some Data (tag name for
// element nodes, content for text) and are part of a tree of Nodes.
type Node struct {
	Parent, PrevSibling, NextSibling, FirstChild, LastChild *Node

	Type NodeType
	Name string
	Data *protoreflect.Value

	level int
}

// Options controls the behavior of protobuf parsing
type Options struct {
	// PopulateDefaults determines whether to populate default values for proto3 messages
	PopulateDefaults bool
}

// DefaultOptions provides sensible defaults for parsing
// PopulateDefaults is false by default to maintain backward compatibility
var DefaultOptions = Options{
	PopulateDefaults: false, // Preserve existing behavior by default
}

// ChildNodes gets all child nodes of the node.
func (n *Node) ChildNodes() []*Node {
	var a []*Node
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		a = append(a, nn)
	}
	return a
}

// InnerText gets the value of the node and all its child nodes.
func (n *Node) InnerText() string {
	var output func(*bytes.Buffer, *Node)
	output = func(buf *bytes.Buffer, n *Node) {
		if n.Type == TextNode {
			buf.WriteString(n.Data.String())
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
	}
	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}

func outputXML(buf *bytes.Buffer, n *Node) {
	if n.Type == TextNode {
		buf.WriteString(n.Data.String())
		return
	}

	name := "element"
	if n.Name != "" {
		name = n.Name
	}
	buf.WriteString("<" + name + ">")
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		outputXML(buf, child)
	}
	buf.WriteString("</" + name + ">")
}

// OutputXML prints the XML string.
func (n *Node) OutputXML() string {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0"?>`)
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		outputXML(&buf, n)
	}
	return buf.String()
}

// SelectElement finds the first of child elements with the
// specified name.
func (n *Node) SelectElement(name string) *Node {
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		if nn.Name == name {
			return nn
		}
	}
	return nil
}

// Value return the value of the node itself or its 'TextElement' children.
// If `nil`, the value is either really `nil` or there is no matching child.
func (n *Node) Value() interface{} {
	if n.Type == TextNode {
		if n.Data == nil {
			return nil
		}
		return n.Data.Interface()
	}

	result := make([]interface{}, 0)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != TextNode || child.Data == nil {
			continue
		}
		result = append(result, child.Data.Interface())
	}

	if len(result) == 0 {
		return nil
	} else if len(result) == 1 {
		return result[0]
	}
	return result
}

// Parse ProtocolBuffer message with default options.
func Parse(msg protoreflect.Message) (*Node, error) {
	return ParseWithOptions(msg, DefaultOptions)
}

// ParseWithOptions parses a ProtocolBuffer message with custom options.
func ParseWithOptions(msg protoreflect.Message, opts Options) (*Node, error) {
	doc := &Node{Type: DocumentNode}
	visit(doc, msg, 1, opts)
	return doc, nil
}

func visit(parent *Node, msg protoreflect.Message, level int, opts Options) {
	desc := msg.Descriptor()
	fields := desc.Fields()

	// Process all fields in the message descriptor, not just present ones
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		if msg.Has(field) {
			// Field is present, use its actual value
			value := msg.Get(field)
			traverse(parent, field, value, level, opts)
		} else if opts.PopulateDefaults && desc.Syntax() == protoreflect.Proto3 {
			// Field is not present in proto3, use default value
			defaultValue := getDefaultValue(field)
			if defaultValue.IsValid() {
				traverse(parent, field, defaultValue, level, opts)
			}
		}
		// For proto2 or when PopulateDefaults is false, skip unset fields (existing behavior)
	}
}

func traverse(parent *Node, field protoreflect.FieldDescriptor, value protoreflect.Value, level int, opts Options) {
	node := &Node{Type: ElementNode, Name: string(field.Name()), level: level}
	nodeChildren := 0
	switch {
	case field.IsList():
		l := value.List()
		for i := 0; i < l.Len(); i++ {
			subNode := handleValue(field.Kind(), l.Get(i), level+1, opts)
			if subNode.Type == ElementNode {
				// Add element nodes directly to the parent
				subNode.Name = node.Name
				addNode(parent, subNode)
			} else {
				// Add basic nodes to the local collection node
				elementNode := &Node{Type: ElementNode, level: level + 1}
				subNode.level += 2
				addNode(elementNode, subNode)
				addNode(node, elementNode)
				nodeChildren++
			}
		}
	case field.IsMap():
		key := field.MapKey()
		value.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			subNode := handleValue(key.Kind(), v, level+1, opts)
			if subNode.Type == ElementNode {
				// Add element nodes directly to the parent
				subNode.Name = k.String()
				addNode(parent, subNode)
			} else {
				// Add basic nodes to the local collection node
				elementNode := &Node{Type: ElementNode, Name: k.String(), level: level + 1}
				subNode.level += 2
				addNode(elementNode, subNode)
				addNode(node, elementNode)
				nodeChildren++
			}
			return true
		})
	default:
		subNode := handleValue(field.Kind(), value, level+1, opts)
		if subNode.Type == ElementNode {
			// Add element nodes directly to the parent
			subNode.Name = node.Name
			addNode(parent, subNode)
		} else {
			// Add basic nodes to the local collection node
			addNode(node, subNode)
			nodeChildren++
		}
	}

	// Only add the node if it has children
	if nodeChildren > 0 {
		addNode(parent, node)
	}
}

func handleValue(kind protoreflect.Kind, value protoreflect.Value, level int, opts Options) *Node {
	var node *Node

	switch kind {
	case protoreflect.MessageKind:
		node = &Node{Type: ElementNode, level: level}
		visit(node, value.Message(), level+1, opts)
	default:
		node = &Node{Type: TextNode, Data: &value, level: level}
	}
	return node
}

func addNode(top, n *Node) {
	if n.level == top.level {
		top.NextSibling = n
		n.PrevSibling = top
		n.Parent = top.Parent
		if top.Parent != nil {
			top.Parent.LastChild = n
		}
	} else if n.level > top.level {
		n.Parent = top
		if top.FirstChild == nil {
			top.FirstChild = n
			top.LastChild = n
		} else {
			t := top.LastChild
			t.NextSibling = n
			n.PrevSibling = t
			top.LastChild = n
		}
	}
}

// getDefaultValue returns the appropriate default value for a field
func getDefaultValue(field protoreflect.FieldDescriptor) protoreflect.Value {
	// Skip lists and maps - they're empty by default and don't need explicit defaults
	if field.IsList() || field.IsMap() {
		return protoreflect.Value{}
	}

	switch field.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(false)
	case protoreflect.EnumKind:
		enumDesc := field.Enum()
		if enumDesc.Values().Len() > 0 {
			return protoreflect.ValueOfEnum(enumDesc.Values().Get(0).Number())
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(0)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(0)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(0)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(0)
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(0.0)
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(0.0)
	case protoreflect.StringKind:
		return protoreflect.ValueOfString("")
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte{})
	case protoreflect.MessageKind, protoreflect.GroupKind:
		// For nested messages, we don't populate defaults automatically
		// to avoid infinite recursion and unnecessary complexity.
		// If users need nested message defaults, they can call ParseWithOptions
		// on the nested message separately.
		return protoreflect.Value{}
	}

	return protoreflect.Value{}
}
