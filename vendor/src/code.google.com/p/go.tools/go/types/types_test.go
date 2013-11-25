// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains tests verifying the types associated with an AST after
// type checking.

package types_test

import (
	"go/ast"
	"go/parser"
	"testing"

	_ "code.google.com/p/go.tools/go/gcimporter"
	. "code.google.com/p/go.tools/go/types"
)

const filename = "<src>"

func makePkg(t *testing.T, src string) (*Package, error) {
	file, err := parser.ParseFile(fset, filename, src, parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}
	// use the package name as package path
	return Check(file.Name.Name, fset, []*ast.File{file})
}

type testEntry struct {
	src, str string
}

// dup returns a testEntry where both src and str are the same.
func dup(s string) testEntry {
	return testEntry{s, s}
}

// types that don't depend on any other type declarations
var independentTestTypes = []testEntry{
	// basic types
	dup("int"),
	dup("float32"),
	dup("string"),

	// arrays
	dup("[10]int"),

	// slices
	dup("[]int"),
	dup("[][]int"),

	// structs
	dup("struct{}"),
	dup("struct{x int}"),
	{`struct {
		x, y int
		z float32 "foo"
	}`, `struct{x int; y int; z float32 "foo"}`},
	{`struct {
		string
		elems []complex128
	}`, `struct{string; elems []complex128}`},

	// pointers
	dup("*int"),
	dup("***struct{}"),
	dup("*struct{a int; b float32}"),

	// functions
	dup("func()"),
	dup("func(x int)"),
	{"func(x, y int)", "func(x int, y int)"},
	{"func(x, y int, z string)", "func(x int, y int, z string)"},
	dup("func(int)"),
	{"func(int, string, byte)", "func(int, string, byte)"},

	dup("func() int"),
	{"func() (string)", "func() string"},
	dup("func() (u int)"),
	{"func() (u, v int, w string)", "func() (u int, v int, w string)"},

	dup("func(int) string"),
	dup("func(x int) string"),
	dup("func(x int) (u string)"),
	{"func(x, y int) (u string)", "func(x int, y int) (u string)"},

	dup("func(...int) string"),
	dup("func(x ...int) string"),
	dup("func(x ...int) (u string)"),
	{"func(x, y ...int) (u string)", "func(x int, y ...int) (u string)"},

	// interfaces
	dup("interface{}"),
	dup("interface{m()}"),
	dup(`interface{String() string; m(int) float32}`),

	// maps
	dup("map[string]int"),
	{"map[struct{x, y int}][]byte", "map[struct{x int; y int}][]byte"},

	// channels
	dup("chan int"),
	dup("chan<- func()"),
	dup("<-chan []func() int"),
}

// types that depend on other type declarations (src in TestTypes)
var dependentTestTypes = []testEntry{
	// interfaces
	dup(`interface{io.Reader; io.Writer}`),
	dup(`interface{m() int; io.Writer}`),
	{`interface{m() interface{T}}`, `interface{m() interface{p.T}}`},
}

func TestTypes(t *testing.T) {
	var tests []testEntry
	tests = append(tests, independentTestTypes...)
	tests = append(tests, dependentTestTypes...)

	for _, test := range tests {
		src := `package p; import "io"; type _ io.Writer; type T ` + test.src
		pkg, err := makePkg(t, src)
		if err != nil {
			t.Errorf("%s: %s", src, err)
			continue
		}
		typ := pkg.Scope().Lookup("T").Type().Underlying()
		str := typ.String()
		if str != test.str {
			t.Errorf("%s: got %s, want %s", test.src, str, test.str)
		}
	}
}

var testExprs = []testEntry{
	// basic type literals
	dup("x"),
	dup("true"),
	dup("42"),
	dup("3.1415"),
	dup("2.71828i"),
	dup(`'a'`),
	dup(`"foo"`),
	dup("`bar`"),

	// arbitrary expressions
	dup("&x"),
	dup("*&x"),
	dup("(x)"),
	dup("x + y"),
	dup("x + y * 10"),
	dup("t.foo"),
	dup("s[0]"),
	dup("s[x:y]"),
	dup("s[:y]"),
	dup("s[x:]"),
	dup("s[:]"),
	dup("f(1, 2.3)"),
	dup("-f(10, 20)"),
	dup("f(x + y, +3.1415)"),
	{"func(a, b int) {}", "(func literal)"},
	{"func(a, b int) []int {}(1, 2)[x]", "(func literal)(1, 2)[x]"},
	{"[]int{1, 2, 3}", "(composite literal)"},
	{"[]int{1, 2, 3}[x:]", "(composite literal)[x:]"},
	{"i.([]string)", "i.(<expr *ast.ArrayType>)"},
}

func TestExprs(t *testing.T) {
	for _, test := range testExprs {
		src := "package p; var _ = " + test.src + "; var (x, y int; s []string; f func(int, float32) int; i interface{}; t interface { foo() })"
		file, err := parser.ParseFile(fset, filename, src, parser.DeclarationErrors)
		if err != nil {
			t.Errorf("%s: %s", src, err)
			continue
		}
		str := ExprString(file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0])
		if str != test.str {
			t.Errorf("%s: got %s, want %s", test.src, str, test.str)
		}
	}
}