// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser // import "graphql.co/parser"

import (
	"bytes"
	"runtime"

	"strings"

	"github.com/fatih/color"
	"graphql.co/ast"
	"graphql.co/token"
)

func caller() string {
	caller, _, _, _ := runtime.Caller(2)
	name := strings.Split(runtime.FuncForPC(caller).Name(), ".")
	call := name[len(name)-1]

	if len(call) < 6 {
		return call
	} else {
		return call[5:]
	}

}

func firstCaller() string {
	caller, _, _, _ := runtime.Caller(1)
	name := strings.Split(runtime.FuncForPC(caller).Name(), ".")
	call := name[len(name)-1]

	if len(call) < 6 {
		return call
	} else {
		return call[5:]
	}

}
func arrow(buf string, tok token.Token) string {
	ret := ""
	for i := 0; i < len(string(buf)); i++ {
		if i >= tok.Start && i <= tok.End {
			ret += "^"
			continue
		} else {
			ret += " "

		}
		switch i {

		case tok.Start - 1, tok.Start - 2, tok.Start - 3:
			ret += ">"
			break
		case tok.End + 1, tok.End + 2, tok.End + 3:
			ret += "<"
			break
		default:
			ret += " "
		}
	}
	return ret
}
func (p *Parser) expect(t token.Token, expected token.Type) bool {
	if t.Type != expected {
		name := caller()
		red := color.New(color.FgRed).SprintFunc()
		errf := ""
		errf += red("While parsing %s were expecting %s but got %s at %d:%d.")
		errf += "\n%s\n%s"
		p.errorf(errf,
			name,
			expected,
			p.curTok.Type,
			p.curTok.Line,
			t.Start,
			strings.Trim(p.lexer.LineBuffer(), "\n"),
			red(arrow(p.lexer.LineBuffer(), t)),
		)
		return false
	} else {
		return true
	}
}

func (p *Parser) panic(message string) {
	p.errorf("%s\nIllegal element '%s' (of type %s) at line %d, character %d\n",
		message,
		p.curTok.Text,
		p.curTok.Type,
		p.curTok.Line,
		p.lexer.Pos(),
	)
}

func NewQuery(query []byte) (*ast.Document, error) {
	var doc ast.Document
	sq := bytes.NewBuffer([]byte(query))
	p := New("sq", sq)
	if p.Decode(&doc); p.curTok.Type == token.Error {
		return nil, p.Error
	} else {
		return p.Document, nil
	}
}

// Decode decodes a graphql ast.
func (p *Parser) Decode(i interface{}) (err error) {
	p.Document = (i.(*ast.Document))
	p.run()
	if p.curTok.Type == token.Error {
		return p.Error
	}

	return nil
}
