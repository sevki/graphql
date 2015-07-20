// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package parser // import "graphql.co/parser"

import (
	"fmt"
	"io"

	"graphql.co/ast"
	"graphql.co/lexer"
	"graphql.co/token"
)

type Parser struct {
	name     string
	lexer    *lexer.Lexer
	state    stateFn
	peekTok  token.Token
	curTok   token.Token
	line     int
	Error    error
	Document *ast.Document
	ptr      ast.Selection
	prnt     ast.Selection
}

type stateFn func(*Parser) stateFn

func (p *Parser) peek() token.Token {
	return p.peekTok
}
func (p *Parser) next() token.Token {
	tok := p.peekTok
	p.peekTok = <-p.lexer.Tokens
	p.curTok = tok

	// yellow := color.New(color.FgYellow).SprintFunc()
	// green := color.New(color.FgGreen).SprintFunc()
	// blue := color.New(color.FgCyan).SprintFunc()

	// log.Printf("%s: %s(%s)\t=>\t%s(%s)\t\n",
	// 	blue(caller()),
	// 	green(tok.Type),
	// 	tok.Text,
	// 	yellow(p.peek().Type),
	// 	p.peek().Text,
	// )
	if tok.Type == token.Error {
		p.errorf("%q", tok)
	}

	return tok
}
func (p *Parser) errorf(format string, args ...interface{}) {
	p.curTok = token.Token{Type: token.Error}
	p.peekTok = token.Token{Type: token.EOF}
	p.Error = fmt.Errorf(format, args...)
	//	log.panicln(p.Error.Error())
}
func New(name string, r io.Reader) *Parser {
	var doc ast.Document

	p := &Parser{
		name:     name,
		line:     0,
		lexer:    lexer.New(name, r),
		Document: &doc,
	}
	return p
}

// run runs the state machine for the Scanner.
func (p *Parser) run() {
	p.next()
	// we start in the GraphQuery State
	for p.state = parseDocument; p.state != nil; {
		p.state = p.state(p)

	}
}

// parseDocumnent
func parseDocument(p *Parser) stateFn {

	for p.peek().Type != token.EOF {
		return parseDefinition
	}
	return nil
}

// parseDefinition
func parseDefinition(p *Parser) stateFn {
	switch p.peek().Type {
	case token.EOF:
		return nil
	case token.LeftCurly, token.QueryStart:
		return parseOperation
	case token.MutationStart:
		return parseOperation
	case token.FragmentStart:
		return parseFragmentDefinition
	default:
		p.expect(p.peekTok, token.LeftCurly)
		return nil
	}
}

// parseOperation
func parseOperation(p *Parser) stateFn {
	t := p.next()
	var op ast.Operation

	if t.Type == token.MutationStart {
		op.OperationType = ast.Mutation
	} else {
		op.OperationType = ast.Query
	}
	if t.Type == token.QueryStart {
		op.Name = ast.GraphQLName(p.next().Text)
	}
	p.Document.Definitions = append(p.Document.Definitions, &op)

	return parseVariables
}

func parseSelection(p *Parser) stateFn {
	switch p.peek().Type {
	case token.Elipsis:
		return parseFragment
	case token.String:
		return parseField
	case token.RightCurly:
		return parseRightCurly
	default:
		return nil
	}
}
func parseSelectionSet(p *Parser) stateFn {

	if p.expect(p.next(), token.LeftCurly) {
		switch p.ptr.(type) {
		case *ast.Field:
			p.ptr.(*ast.Field).Parent = p.prnt
		case *ast.Fragment:
			p.ptr.(*ast.Fragment).Parent = p.prnt
		}
		p.prnt = p.ptr
		return parseSelection
	}
	return nil

}
func parseRightCurly(p *Parser) stateFn {

	p.next()
	if p.prnt == nil {
		return parseDefinition
	}
	switch p.prnt.(type) {
	case *ast.Field:
		p.prnt = p.prnt.(*ast.Field).Parent
	case *ast.Fragment:
		p.prnt = p.prnt.(*ast.Fragment).Parent
	}
	p.ptr = p.prnt
	switch p.peek().Type {
	case token.Elipsis, token.String, token.RightCurly:
		return parseSelection
	default:
		return parseDocument
	}
}
func parseField(p *Parser) stateFn {

	t := p.next()
	if !p.expect(t, token.String) {
		return nil
	}
	var field ast.Field
	field.Parent = p.prnt
	if p.peek().Type == token.Colon {
		field.Alias = ast.GraphQLName(t.Text)
		p.next()
		field.Name = ast.GraphQLName(p.next().Text)
	} else {
		field.Name = ast.GraphQLName(t.Text)
	}

	if p.prnt == nil {
		def := p.Document.Definitions[len(p.Document.Definitions)-1]
		def.AddSelection(&field)
	} else {
		p.prnt.AddSelection(&field)
	}
	if p.peek().Type == token.Elipsis {
		return parseSelection
	} else {
		p.ptr = &field
		return parseArguments

	}

}
func parseFragment(p *Parser) stateFn {
	if !p.expect(p.next(), token.Elipsis) {
		return nil
	}
	if p.peek().Type == token.On {
		return parseInlineFragment
	} else if p.peek().Type == token.String {
		return parseFragmentSpread
	} else {
		return nil
	}
}
func parseFragmentSpread(p *Parser) stateFn {

	t := p.next()
	frag := ast.Fragment{FragmentName: ast.GraphQLName(t.Text)}
	p.prnt.AddSelection(&frag)
	p.ptr = &frag

	return parseDirectives
}

func parseInlineFragment(p *Parser) stateFn {
	p.next()
	t := p.next()

	frag := ast.Fragment{TypeCondition: ast.GraphQLName(t.Text)}
	p.prnt.AddSelection(&frag)
	p.ptr = &frag
	if p.peek().Type == token.Directive {
		return parseDirectives
	}
	return nil
}
func parseDirectives(p *Parser) stateFn {
	//	log.Println(firstCaller())
	for p.peek().Type == token.Directive {
		t := p.next()
		if !p.expect(t, token.Directive) {
			return nil
		}
		if p.peek().Type == token.LeftParen {

			if p.ptr == nil {
				op := p.Document.Definitions[len(p.Document.Definitions)-1].(*ast.Operation)
				op.AddDirective(string(t.Text), p.parseArguments())
			} else {
				p.ptr.AddDirective(string(t.Text), p.parseArguments())
			}
		} else {
			if p.ptr == nil {
				op := p.Document.Definitions[len(p.Document.Definitions)-1].(*ast.Operation)
				op.AddDirective(string(t.Text), nil)
			} else {
				p.ptr.AddDirective(string(t.Text), nil)
			}
		}

	}

	if p.peek().Type == token.LeftCurly {
		return parseSelectionSet
	} else if p.peek().Type == token.RightCurly {
		return parseRightCurly
	} else {
		return parseSelection
	}
}

func parseArguments(p *Parser) stateFn {

	p.ptr.(*ast.Field).Arguments = p.parseArguments()

	return parseDirectives

}

//--------------------------------------------------------------
// parseArguments
func (p *Parser) parseArguments() ast.Arguments {
	if p.peek().Type == token.LeftParen {
		p.next()
		args := make(ast.Arguments)
		for p.peek().Type != token.RightParen {

			key := p.next()
			if !p.expect(key, token.String) {
				break
			}

			//followed by colon
			if !p.expect(p.next(), token.Colon) {
				break
			}

			if p.peek().Type == token.LeftBrac {
				var ary ast.ArrayValue
				p.next() // advance left brac
				for t := p.next(); t.Type != token.RightBrac; t = p.next() {
					ary = append(ary, ast.GraphQLValue(t))
				}
				args[string(key.Text)] = ary
			} else {
				value := p.next()

				args[string(key.Text)] = ast.GraphQLValue(value)
			}
		}
		p.next() // right paren
		return args
	}
	return nil
}
func parseVariables(p *Parser) stateFn {
	if p.next().Type == token.LeftParen {
		op := p.Document.Definitions[len(p.Document.Definitions)-1].(*ast.Operation)
		op.VariableDefinitions = make(ast.VariableDefinitions)

		for p.peek().Type != token.RightParen {

			varName := p.next()
			if !p.expect(varName, token.Variable) {
				return nil
			}

			//followed by colon
			if !p.expect(p.next(), token.Colon) {
				return nil
			}

			typeName := p.next()
			if !p.expect(typeName, token.String) {
				return nil
			}

			varb := ast.Variable{Type: ast.Type(typeName.Text)}

			//followed by Equals
			if p.peek().Type == token.Equal {
				p.next()
				varb.DefaultValue = ast.GraphQLString(p.next().Text)
			}

			op.VariableDefinitions[string(varName.Text)] = varb
		}
	}
	p.next() // right paren
	return parseDirectives
}

func parseFragmentDefinition(p *Parser) stateFn {
	if !p.expect(p.next(), token.FragmentStart) {
		return nil
	}
	var frag ast.Fragment
	t := p.next()
	if !p.expect(t, token.String) {
		return nil
	} else {
		frag.FragmentName = ast.GraphQLName(string(t.Text))
	}
	p.Document.Definitions = append(p.Document.Definitions, &frag)
	if !p.expect(p.next(), token.On) {
		return nil
	} else {
		t = p.next()
		if !p.expect(t, token.String) {
			return nil
		} else {
			frag.TypeCondition = ast.GraphQLName(string(t.Text))
		}
	}
	return parseDirectives
}
