// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type ParamType

package query // import "sevki.org/graphql/query"

import (
	"fmt"
	"strconv"

	"sevki.org/graphql/lexer"
)

type Node struct {
	Name   UString
	Parent *Node `json:"-"`
	Edges  []Node
	Params map[string]Param
}

type ParamType int
type UString []uint8

const (
	Empty ParamType = iota
	Int
	String
	Variable
	Tuple
)

type IntParam struct {
	Value int
}
type VariableParam struct {
	Value string
}
type StringParam struct {
	Value UString
}

type TupleParam struct {
	Key   UString
	Value Param
}

type ArrayParam []ParamPrimitive

type Param interface {
	isParam()
	Set(lexer.Token) Param
	String() string
}
type ParamPrimitive interface {
	isPrim()
}

func (i IntParam) isParam()      {}
func (i TupleParam) isParam()    {}
func (i StringParam) isParam()   {}
func (i ArrayParam) isParam()    {}
func (i VariableParam) isParam() {}

func (i IntParam) isPrim()    {}
func (i StringParam) isPrim() {}

func (i IntParam) Set(t lexer.Token) Param {
	if inty, err := strconv.Atoi(string(t.Text)); err == nil {
		i.Value = inty
	}
	return i
}

func (i VariableParam) Set(t lexer.Token) Param {
	i.Value = string(t.Text)

	return i
}
func (i StringParam) Set(t lexer.Token) Param {

	if t.Type == lexer.Quote {
		i.Value = t.Text[1 : len(t.Text)-1]
	}
	return i
}

func (i TupleParam) Set(t lexer.Token) Param {

	switch t.Type {
	case lexer.Number:
		i.Value = IntParam{}
	case lexer.Quote:
		i.Value = StringParam{}
	}

	return i.Value.Set(t)
}

func (i ArrayParam) Set(lexer.Token) {}
func (i TupleParam) SetKey(t UString) {
	i.Key = t
}

func (i IntParam) String() (str string) {
	str = fmt.Sprintf("%d", i.Value)
	return
}

func (i StringParam) String() (str string) {
	str = fmt.Sprintf("%v", string(i.Value))
	return
}

func (i VariableParam) String() (str string) {
	str = fmt.Sprintf("%v", i.Value)
	return
}
