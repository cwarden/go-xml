package xpath

import (
	"testing"

	"github.com/knroy/go-xml/v2/xdm"
)

// TestSuppliedTypedValue evaluates expressions over an attribute whose typed
// value was supplied with xdm.Typing.TypedValue: the items take part in
// comparisons and fn:data as given, and fn:string still returns the string
// value the attribute was built with.
func TestSuppliedTypedValue(t *testing.T) {
	tree := xdm.NewTree()
	e := tree.Root.AppendElement(xdm.QName{Local: "e"})
	a := e.AppendAttr(xdm.QName{Local: "args"}, "[String, Map<String, Integer>]")
	a.ApplyTyping(xdm.Typing{TypeAnnotation: xdm.AnnotationName("urn:test", "list"),
		TypedValue: xdm.NewTypedValue(xdm.NewString("String"), xdm.NewString("Map<String, Integer>"))})
	n := e.AppendAttr(xdm.QName{Local: "none"}, "[]")
	n.ApplyTyping(xdm.Typing{TypeAnnotation: xdm.AnnotationName("urn:test", "list"),
		TypedValue: xdm.NewTypedValue()})
	tree.Finalize()

	for _, tc := range []struct{ expr, want string }{
		{`count(@args)`, "1"},
		{`count(data(@args))`, "2"},
		{`@args = 'Map<String, Integer>'`, "true"},
		{`@args = 'Map<String,'`, "false"},
		{`string-join(data(@args), '|')`, "String|Map<String, Integer>"},
		{`string(@args)`, "[String, Map<String, Integer>]"},
		{`data(@args) instance of xs:string+`, "true"},
		{`count(data(@none))`, "0"},
		{`@none = ''`, "false"},
	} {
		c, err := CompileWith(tc.expr, CompileOptions{Version: XPath31})
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		got, err := c.EvalString(NewContext(e, Builtins()))
		if err != nil {
			t.Errorf("%s: %v", tc.expr, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s = %q, want %q", tc.expr, got, tc.want)
		}
	}
}
