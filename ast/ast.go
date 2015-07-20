// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type OperationType

// Package ast defines the data structures that are held in
// http://facebook.github.io/graphql/#sec-Grammar

package ast // import "graphql.co/graphql/ast"

import (
	"fmt"

	"graphql.co/graphql/token"
)

import (
	"strconv"
	"strings"
)

// Document as defined in
// http://facebook.github.io/graphql/#sec-Syntax.Document
type Document struct {
	Definitions []Definition
}

// Definition as defined in
// http://facebook.github.io/graphql/#Definition
type Definition interface {
	isDefinition()
	AddDirective(k string, args Arguments)
	AddSelection(s Selection)
}

// Operation as defined in
// http://facebook.github.io/graphql/#sec-Syntax.Operations
type Operation struct {
	OperationType OperationType
	Name          GraphQLName
	VariableDefinitions
	Directives
	SelectionSet
}

func (*Operation) isDefinition() {}
func (o *Operation) AddDirective(k string, args Arguments) {
	if o.Directives == nil {
		o.Directives = make(Directives)
	}
	o.Directives[k] = args
}
func (o *Operation) AddSelection(s Selection) {
	o.SelectionSet = append(o.SelectionSet, s)
}

// VariableDefinitions as defined in
// http://facebook.github.io/graphql/#VariableDefinition
type VariableDefinitions map[string]Variable

// OperationType as defined in
// http://facebook.github.io/graphql/#OperationType
type OperationType int

// OperationTypes as defined in
// http://facebook.github.io/graphql/#OperationType
const (
	Query OperationType = iota
	Mutation
)

// Selection as defined in
// http://facebook.github.io/graphql/#Selection
type Selection interface {
	isSelection()
	AddDirective(k string, args Arguments)
	AddSelection(s Selection)
}

// SelectionSet as defined in
// http://facebook.github.io/graphql/#SelectionSet
type SelectionSet []Selection

// Field as defined in http://facebook.github.io/graphql/#Field
type Field struct {
	Alias GraphQLName
	Name  GraphQLName
	Arguments
	Directives
	SelectionSet
	Parent Selection `json:"-"`
}

func (*Field) isSelection() {}
func (f *Field) AddDirective(k string, args Arguments) {
	if f.Directives == nil {
		f.Directives = make(Directives)
	}
	f.Directives[k] = args
}
func (f *Field) AddSelection(s Selection) {
	f.SelectionSet = append(f.SelectionSet, s)
}

// Fragment as defined in
// http://facebook.github.io/graphql/#FragmentDefinition
type Fragment struct {
	FragmentName GraphQLName
	Directives   Directives
	// TypeCondition as defined in
	// http://facebook.github.io/graphql/#TypeCondition
	TypeCondition GraphQLName
	SelectionSet  SelectionSet
	Parent        Selection `json:"-"`
}

func (*Fragment) isDefinition() {}
func (*Fragment) isSelection()  {}
func (f *Fragment) AddDirective(k string, args Arguments) {
	if f.Directives == nil {
		f.Directives = make(Directives)
	}
	f.Directives[k] = args
}
func (f *Fragment) AddSelection(s Selection) {
	f.SelectionSet = append(f.SelectionSet, s)
}

// Arguments as defined in
// http://facebook.github.io/graphql/#Arguments
type Arguments map[string]Value

// Directives as defined in
// http://facebook.github.io/graphql/#Directives
type Directives map[string]Arguments

// Type as defined in http://facebook.github.io/graphql/#Type
type Type string

// Types as defined in
// http://facebook.github.io/graphql/#sec-Syntax.Types
type TypeName GraphQLName
type ListType []Type
type NonNullType Type

// Variable as defined in http://facebook.github.io/graphql/#Variable
type Variable struct {
	Type         Type
	DefaultValue Value
}

// Value as defined in http://facebook.github.io/graphql/#sec-Values
type Value interface {
	isValue()
}

// Scalar as defined in http://facebook.github.io/graphql/#sec-Scalars
type Scalar interface {
	isScalar()
}

// GraphQLName as defined in http://facebook.github.io/graphql/#Name
type GraphQLName string

// GraphQLInt scalar type represents a signed 32‐bit numeric
// non‐fractional values. Response formats that support a 32‐bit
// integer or a number type should use that type to represent this
// scalar.
// GraphQLInt as defined in http://facebook.github.io/graphql/#sec-Int
type GraphQLInt int32

func (GraphQLInt) isValue()  {}
func (GraphQLInt) isScalar() {}

// GraphQLFloat  scalar type represents signed double‐precision
// fractional values as specified by IEEE 754. Response formats that
// support an appropriate double‐precision number type should use that
// type to represent this scalar.
// GraphQLFloat as defined in
// http://facebook.github.io/graphql/#sec-Float
type GraphQLFloat float32

func (GraphQLFloat) isValue()  {}
func (GraphQLFloat) isScalar() {}

// GraphQLString scalar type represents textual data, represented as
// UTF‐8 character sequences. The String type is most often used by
// GraphQL to represent free‐form human‐readable text. All response
// formats must support string representations, and that
// representation must be used here.
// GraphQLString as defined in
// http://facebook.github.io/graphql/#sec-String
type GraphQLString string

func (GraphQLString) isValue()  {}
func (GraphQLString) isScalar() {}

// GraphQLBoolean scalar type represents true or
// false. Response formats should use a built‐in boolean type if
// supported; otherwise, they should use their representation of the
// integers 1 and 0.
// GraphQLBoolean as defined in
// http://facebook.github.io/graphql/#sec-Boolean
type GraphQLBoolean bool

func (GraphQLBoolean) isValue()  {}
func (GraphQLBoolean) isScalar() {}

// GraphQLID scalar type represents a unique identifier, often
// used to refetch an object or as key for a cache. The ID type is
// serialized in the same way as a String; however, it is not intended
// to be human‐readable. While it is often numeric, it should always
// serialize as a String. GraphQLIDas defined in
// http://facebook.github.io/graphql/#sec-ID
// https://en.wikipedia.org/wiki/Globally_unique_identifier
type GraphQLID string

func (GraphQLID) isValue()  {}
func (GraphQLID) isScalar() {}

// ArrayValue as defined in
// http://facebook.github.io/graphql/#ArrayValue
type ArrayValue []Value

func (ArrayValue) isValue() {}

type GraphQLError string

func (GraphQLError) isValue() {}

func GraphQLValue(t token.Token) Value {
	switch t.Type {
	case token.True:
		return GraphQLBoolean(true)
	case token.False:
		return GraphQLBoolean(false)
	case token.Number:
		if i, err := strconv.Atoi(string(t.Text)); err !=
			nil {
			return GraphQLError(err.Error())
		} else {
			return GraphQLInt(i)
		}
	case token.Hex:
		var i int
		if i, err := fmt.Sscanf(string(t.Text), "%X", &i); err !=
			nil {
			return GraphQLError(err.Error())
		} else {
			return GraphQLInt(i)
		}
	case token.String:
		return GraphQLString(string(t.Text))
	case token.Quote:
		return GraphQLString(strings.Trim(string(t.Text), "\""))
	default:
		return GraphQLError(fmt.Sprintf("%s is not a GraphQL value", t.Type))
	}

}
