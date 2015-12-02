package xpath

import (
	"fmt"

	"github.com/lestrrat/go-libxml2/clib"
	"github.com/lestrrat/go-libxml2/node"
)

// Pointer returns the underlying C struct
func (x Object) Pointer() uintptr {
	return x.ptr
}

// Type returns the clib.XPathObjectType
func (x Object) Type() clib.XPathObjectType {
	return clib.XPathObjectType(clib.XMLXPathObjectType(x))
}

// Number returns the floatval component of the Object as float64
func (x Object) Number() float64 {
	return clib.XMLXPathObjectFloat64(x)
}

// Bool returns the boolval component of the Object
func (x Object) Bool() bool {
	return clib.XMLXPathObjectBool(x)
}

var WrapNodeFunc func(uintptr) (node.Node, error)

// NodeList returns the list of nodes included in this Object
func (x Object) NodeList() node.List {
	if WrapNodeFunc == nil {
		panic("WarapNodeFunc not initialized. read XXX for details")
	}

	nl, err := clib.XMLXPathObjectNodeList(x)
	if err != nil {
		return nil
	}

	ret := make([]node.Node, len(nl))
	for i, p := range nl {
		n, err := WrapNodeFunc(p)
		if err != nil {
			return nil
		}
		ret[i] = n
	}
	return ret
}

// String returns the stringified value of the nodes included in
// this Object. If the Object is anything other than a
// NodeSet, then we fallback to using fmt.Sprintf to generate
// some sort of readable output
func (x Object) String() string {
	switch x.Type() {
	case NodeSetType:
		nl := x.NodeList()
		if nl == nil {
			return ""
		}
		if x.ForceLiteral {
			s, err := nl.Literal()
			if err == nil {
				return s
			}
			return ""
		}
		return nl.NodeValue()
	default:
		return fmt.Sprintf("%v", x)
	}
}

// Free releases the underlying C structs
func (x *Object) Free() {
	clib.XMLXPathFreeObject(x)
}

// NewExpression compiles the given XPath expression string
func NewExpression(s string) (*Expression, error) {
	ptr, err := clib.XMLXPathCompile(s)
	if err != nil {
		return nil, err
	}

	return &Expression{ptr: ptr}, nil
}

// Pointer returns the underlying C struct
func (x *Expression) Pointer() uintptr {
	return x.ptr
}

// Free releases the underlying C structs in the Expression
func (x *Expression) Free() {
	clib.XMLXPathFreeCompExpr(x)
}

// NewContext creates a new Context, optionally providing
// with a context node.
//
// Note that although we are specifying `n... Node` for the argument,
// only the first, node is considered for the context node
func NewContext(n ...node.Node) (*Context, error) {
	var node node.Node
	if len(n) > 0 {
		node = n[0]
	}

	ctxptr, err := clib.XMLXPathNewContext(node)
	if err != nil {
		return nil, err
	}

	return &Context{ptr: ctxptr}, nil
}

func (x *Context) Pointer() uintptr {
	return x.ptr
}

// SetContextNode sets or resets the context node which
// XPath expressions will be evaluated against.
func (x *Context) SetContextNode(n node.Node) error {
	return clib.XMLXPathContextSetContextNode(x, n)
}

// Exists compiles and evaluates the xpath expression, and returns
// true if a corresponding node exists
func (x *Context) Exists(xpath string) bool {
	list := NodeList(x.Find(xpath))
	if list == nil {
		return false
	}

	return len(list) > 0
}

// Free releases the underlying C structs in the XPath
func (x *Context) Free() {
	clib.XMLXPathFreeContext(x)
}

func (x *Context) evalXPathExpr(expr *Expression) (node.XPathResult, error) {
	res, err := clib.XMLEvalXPath(x, expr)
	if err != nil {
		return nil, err
	}

	return &Object{ptr: res}, nil
}

// Find evaluates the expression s against the nodes registered
// in x. It returns the resulting data evaluated to an Result.
//
// You MUST call Free() on the Result, or you will leak memory
// If you don't really care for errors and just want to grab the
// value of Result, checkout xpath.String(), xpath.Number(), xpath.Bool()
// et al.
func (x *Context) Find(s string) (node.XPathResult, error) {
	expr, err := NewExpression(s)
	if err != nil {
		return nil, err
	}
	defer expr.Free()

	return x.FindExpr(expr)
}

// FindExpr evaluates the given XPath expression and returns an Object.
// You must call `Free()` on this returned object
//
// You MUST call Free() on the Result, or you will leak memory
func (x *Context) FindExpr(expr *Expression) (node.XPathResult, error) {
	o, err := x.evalXPathExpr(expr)
	if err != nil {
		return nil, err
	}
	//	res.ForceLiteral = true
	return o, err
}

// LookupNamespaceURI looksup the namespace URI associated with prefix
func (x *Context) LookupNamespaceURI(prefix string) (string, error) {
	return clib.XMLXPathNSLookup(x, prefix)
}

// RegisterNS registers a namespace so it can be used in an Expression
func (x *Context) RegisterNS(name, nsuri string) error {
	return clib.XMLXPathRegisterNS(x, name, nsuri)
}

