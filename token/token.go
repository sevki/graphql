// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package token // import "graphql.co/token"

type Token struct {
	Type  Type
	Line  int
	Text  []byte
	Start int
	End   int
}

type Type int

const (
	EOF Type = iota
	Error
	Newline
	String
	Space
	Number
	Float
	Hex
	LeftCurly
	RightCurly
	LeftParen
	RightParen
	LeftBrac
	RightBrac
	Quote
	Equal
	Colon
	Comma
	Semicolon
	Period
	Comment
	Pipe
	Variable
	Elipsis
	Key
	Directive
	FragmentStart
	QueryStart
	MutationStart
	On
	True
	False
)
