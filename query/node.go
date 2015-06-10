// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type ParamType

package query

type Node struct {
	Name   UString
	Parent *Node `json:"-"`
	Edges  []Node
	Params []Param
}

type ParamType int
type UString []byte

const (
	Empty ParamType = iota
	Int
	String
	Tuple
)

type IntParam struct {
	Value int
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
}
type ParamPrimitive interface {
	isPrim()
}

func (i IntParam) isParam()    {}
func (i TupleParam) isParam()  {}
func (i StringParam) isParam() {}
func (i ArrayParam) isParam()  {}

func (i IntParam) isPrim()    {}
func (i StringParam) isPrim() {}
