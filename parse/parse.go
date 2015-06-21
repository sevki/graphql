// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package parse

import (
	"bytes"
	"fmt"
	"io"

	"sevki.org/graphql/lexer"
	"sevki.org/graphql/query"
)

type Parser struct {
	name    string
	lexer   *lexer.Lexer
	state   stateFn
	peekTok lexer.Token
	curTok  lexer.Token
	line    int
	Error   error
	root    *query.Node
	ptr     *query.Node
}

type stateFn func(*Parser) stateFn

func (p *Parser) peek() lexer.Token {
	return p.peekTok
}
func (p *Parser) next() lexer.Token {
	tok := p.peekTok
	p.peekTok = <-p.lexer.Tokens
	p.curTok = tok
	//	log.Printf("%s(%s)\t=>\t%s(%s)\t%+v\n", tok.Type, tok.Text, p.peek().Type, p.peek().Text, p.rootQuery)
	if tok.Type == lexer.Error {
		p.errorf("%q", tok)
	}

	return tok
}
func (p *Parser) errorf(format string, args ...interface{}) {
	p.curTok = lexer.Token{Type: lexer.Error}
	p.peekTok = lexer.Token{Type: lexer.EOF}
	p.Error = fmt.Errorf(format, args...)
	//	log.panicln(p.Error.Error())
}
func New(name string, r io.ByteReader) *Parser {
	p := &Parser{
		name:  name,
		line:  0,
		lexer: lexer.New(name, r),
	}
	p.run()

	return p
}

// run runs the state machine for the Scanner.
func (p *Parser) run() {
	p.next()
	// we start in the GraphQuery State
	for p.state = parseGraph; p.state != nil; {
		p.state = p.state(p)

	}
}
func parseAny(p *Parser) stateFn {
	switch p.next().Type {
	case lexer.Period:
		return nil
	case lexer.LeftCurly:
		return parseEdges
	case lexer.Comma:
		p.ptr = p.ptr.Parent
		return parseEdges
	default:
		return nil
	}
	return nil
}
func parseGraph(p *Parser) stateFn {
	if p.next().Type == lexer.LeftCurly {
		// We are in the RootQuery
		p.root = &query.Node{}
		p.ptr = p.root
		return parseNode
	} else {
		p.errorf("Expected first item (%s) but got %s", lexer.LeftCurly, p.curTok.Type)
		return nil
	}

}

func parseNode(p *Parser) stateFn {
	if t := p.next(); t.Type == lexer.String {
		p.ptr.Name = t.Text
		if p.peek().Type == lexer.LeftParen {
			return parseParams
		}
	} else {
		p.panic("We were expecting a node here")
	}
	return parseAny
}
func parseEdges(p *Parser) stateFn {
	t := p.peek()
	if t.Type != lexer.String {
		p.panic("This is not a valid edge name")
	}
	p.ptr.Edges = append(p.ptr.Edges, query.Node{})
	k := p.ptr
	p.ptr = &k.Edges[len(k.Edges)-1]
	p.ptr.Parent = k
	return parseNode
}
func parseParams(p *Parser) stateFn {
	p.ptr.Params = make(map[string]query.Param)
	for {
		t := p.next()

		switch t.Type {
		case lexer.RightParen:
			return parseAny
		case lexer.Comma, lexer.LeftParen:
			return parseParams
		}
		var param query.Param
		// If our next token is a tuple
		if p.peek().Type != lexer.Colon || p.peek().Type != lexer.Comma {
			p.panic("Params need to be named.")
		}
		key := t.Text

		p.next() // advance colon

		switch p.peek().Type {
		case lexer.Number:
			param = query.IntParam{}
		case lexer.Quote:
			param = query.StringParam{}
		}

		if t = p.next(); true {
			param = param.Set(t)
		} else {
			p.panic("Unsupported param value.")
			return nil
		}

		p.ptr.Params[string(key)] = param
	}
	return parseAny
}
func parseArray(p *Parser) stateFn {
	return parseAny
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

func NewQuery(query []byte) (*query.Node, error) {
	sq := bytes.NewBuffer([]byte(query))
	var p *Parser
	if p = New("sq", sq); p.curTok.Type == lexer.Error {
		return nil, p.Error
	} else {

		return p.root, nil
	}
}

// Decode decodes a graphql query.
func (p *Parser) Decode(i interface{}) (err error) {
	if p.curTok.Type == lexer.Error {
		return p.Error
	}
	i = p.root
	return nil
}
