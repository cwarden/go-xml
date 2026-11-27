package xdm

import (
	"strings"
	"testing"
)

// listAttr builds <e a="[A, Map<String, Integer>]"/> with a supplied typed
// value of two strings, the second containing a space, and returns the
// element and the attribute. The string value is written in a notation that
// is not a whitespace-separated list, so a typed value derived from it would
// be wrong in both the number of items and their text.
func listAttr(t *testing.T) (*Node, *Node) {
	t.Helper()
	tree := NewTree()
	e := tree.Root.AppendElement(QName{Local: "e"})
	a := e.AppendAttr(QName{Local: "a"}, "[A, Map<String, Integer>]")
	a.ApplyTyping(Typing{TypeAnnotation: AnnotationName("urn:test", "list"),
		TypedValue: NewTypedValue(NewString("A"), NewString("Map<String, Integer>"))})
	tree.Finalize()
	return e, a
}

// itemsOf renders a sequence of atomic values as "type:value" pairs.
func itemsOf(t *testing.T, seq Sequence) string {
	t.Helper()
	var parts []string
	for _, it := range seq {
		a, ok := it.(*Atomic)
		if !ok {
			t.Fatalf("item %T is not atomic", it)
		}
		parts = append(parts, a.Type.String()+":"+a.String())
	}
	return strings.Join(parts, "|")
}

// A supplied typed value is the node's typed value as given: its items are
// not re-derived from the string value, and the string value is not changed.
func TestSuppliedTypedValueAtomizes(t *testing.T) {
	_, a := listAttr(t)
	const want = "xs:string:A|xs:string:Map<String, Integer>"
	if got := itemsOf(t, Atomize(Sequence{a})); got != want {
		t.Errorf("Atomize = %s, want %s", got, want)
	}
	seq, err := AtomizeChecked(Sequence{a})
	if err != nil {
		t.Fatal(err)
	}
	if got := itemsOf(t, seq); got != want {
		t.Errorf("AtomizeChecked = %s, want %s", got, want)
	}
	if got, ok := a.AtomizeList(); !ok || itemsOf(t, got) != want {
		t.Errorf("AtomizeList = %s, %v", itemsOf(t, got), ok)
	}
	if got := a.StringValue(); got != "[A, Map<String, Integer>]" {
		t.Errorf("StringValue = %q", got)
	}
}

// AtomizeList returns a new sequence on each call, so a caller that modifies
// it cannot change the node's typed value.
func TestSuppliedTypedValueIsNotShared(t *testing.T) {
	_, a := listAttr(t)
	got, _ := a.AtomizeList()
	got[0] = NewString("changed")
	if again, _ := a.AtomizeList(); again[0].(*Atomic).String() != "A" {
		t.Errorf("modifying a result changed the typed value to %v", again[0])
	}

	items := []*Atomic{NewString("x")}
	v := NewTypedValue(items...)
	items[0] = NewString("y")
	if got := v.Items()[0].(*Atomic).String(); got != "x" {
		t.Errorf("NewTypedValue kept the caller's slice: item is %q", got)
	}
}

// A typed value of one item is what Node.Atomize returns; an empty one
// atomizes to the empty sequence.
func TestSuppliedTypedValueLengths(t *testing.T) {
	tree := NewTree()
	e := tree.Root.AppendElement(QName{Local: "e"})
	one := e.AppendAttr(QName{Local: "one"}, "0x07")
	one.ApplyTyping(Typing{TypeAnnotation: "integer", TypedValue: NewTypedValue(NewInteger(7))})
	none := e.AppendAttr(QName{Local: "none"}, "[]")
	none.ApplyTyping(Typing{TypeAnnotation: AnnotationName("urn:test", "list"), TypedValue: NewTypedValue()})
	tree.Finalize()

	if got := one.Atomize(); got.Type != TypeInteger || got.String() != "7" {
		t.Errorf("Atomize = %s %s, want xs:integer 7", got.Type, got.String())
	}
	if got := Atomize(Sequence{none}); len(got) != 0 {
		t.Errorf("an empty typed value atomized to %d items", len(got))
	}
}

// The typed value is read only while the node has an annotation, and any new
// annotation replaces it, as it replaces the rest of the typing it came with.
func TestSuppliedTypedValueFollowsTheAnnotation(t *testing.T) {
	_, a := listAttr(t)
	a.SetTypeAnnotation("string")
	if got := itemsOf(t, Atomize(Sequence{a})); got != "xs:string:[A, Map<String, Integer>]" {
		t.Errorf("after re-annotation Atomize = %s", got)
	}
	if TypingOf(a).TypedValue != nil {
		t.Error("SetTypeAnnotation kept the supplied typed value")
	}

	_, b := listAttr(t)
	b.SetTypeAnnotation("")
	if got := itemsOf(t, Atomize(Sequence{b})); got != "xs:untypedAtomic:[A, Map<String, Integer>]" {
		t.Errorf("after clearing the annotation Atomize = %s", got)
	}

	tree := NewTree()
	c := tree.Root.AppendElement(QName{Local: "e"}).AppendAttr(QName{Local: "c"}, "v")
	c.ApplyTyping(Typing{TypedValue: NewTypedValue(NewString("x"), NewString("y"))})
	if got := itemsOf(t, Atomize(Sequence{c})); got != "xs:untypedAtomic:v" {
		t.Errorf("a typed value without an annotation was used: %s", got)
	}
}

// A copy keeps the typed value, whether the tree is copied or the typing is
// copied onto a node built separately.
func TestSuppliedTypedValueIsCopied(t *testing.T) {
	e, a := listAttr(t)
	const want = "xs:string:A|xs:string:Map<String, Integer>"

	c := Copy(e)
	if got := itemsOf(t, Atomize(Sequence{c.AttrAt(0)})); got != want {
		t.Errorf("Copy: Atomize = %s, want %s", got, want)
	}

	tree := NewTree()
	d := tree.Root.AppendElement(QName{Local: "e"}).AppendAttr(QName{Local: "a"}, a.Value())
	d.CopyTypingFrom(a)
	if got := itemsOf(t, Atomize(Sequence{d})); got != want {
		t.Errorf("CopyTypingFrom: Atomize = %s, want %s", got, want)
	}
}

func TestNewTypedValueRefusesNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewTypedValue accepted a nil item")
		}
	}()
	NewTypedValue(NewString("a"), nil)
}
