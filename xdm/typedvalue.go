package xdm

import "fmt"

// TypedValue is a node's typed value (XDM 3.1 section 5.12, dm:typed-value)
// given by the program that built the node, for a node whose typed value
// cannot be derived from its string value and type annotation.
//
// Schema assessment derives a typed value from the string value: an
// xs:integer attribute holding "07" atomizes to 7, and an attribute of a
// list type atomizes to one item per whitespace-separated token. A program
// that exposes its own objects as an XDM tree can hold values that this
// derivation cannot produce: a list whose items contain spaces, or a list
// whose string value is written in another notation, such as "[a, b c]" for
// the items "a" and "b c". Such a program records the typed value on the node
// with Typing.TypedValue, and the node keeps the string value it was given.
//
// A TypedValue holds its items and never changes after NewTypedValue returns,
// so the nodes of a tree and their copies can share one.
type TypedValue struct {
	items []*Atomic
}

// NewTypedValue returns a typed value of the given items, in order. No items
// is the empty sequence. A nil item panics: a typed value is a sequence of
// atomic values, and nil is not one.
//
// The slice is copied, so the caller may reuse it.
func NewTypedValue(items ...*Atomic) *TypedValue {
	for i, a := range items {
		if a == nil {
			panic(fmt.Sprintf("xdm: NewTypedValue: item %d is nil", i))
		}
	}
	return &TypedValue{items: append([]*Atomic(nil), items...)}
}

// Len returns the number of items.
func (v *TypedValue) Len() int { return len(v.items) }

// Items returns the items as a new sequence, which the caller may modify.
func (v *TypedValue) Items() Sequence {
	out := make(Sequence, len(v.items))
	for i, a := range v.items {
		out[i] = a
	}
	return out
}

// suppliedValue returns the typed value given to n with Typing.TypedValue,
// or nil when there is none. Like ListItem it is read only while n has a type
// annotation: the value is what the annotation means, and a node whose
// annotation was cleared atomizes as untyped.
func (n *Node) suppliedValue() *TypedValue {
	nm := n.typ().get()
	if nm.annotation == "" {
		return nil
	}
	return nm.value
}
