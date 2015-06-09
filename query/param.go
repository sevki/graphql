// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type ParamType

package query

type ParamType int

const (
	Empty ParamType = iota
	Int
	String
	Tuple
)

type Param interface {
	Type() ParamType
}

type IntParam struct {
	Value int
}

func (i IntParam) Type() ParamType {
	return Int
}

type TupleParam struct {
	Key   []byte
	Value Param
}

func (i TupleParam) Type() ParamType {
	return Tuple
}

type StringParam struct {
	Value []byte
}

func (i StringParam) Type() ParamType {
	return String
}
