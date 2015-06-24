// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type Type

package lexer

import (
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

type Token struct {
	Type Type
	Line int
	Text []byte
}

type Type int

const eof = -1

const (
	EOF Type = iota
	Error
	Newline
	String
	Space
	Number
	LeftCurly
	RightCurly
	LeftParen
	RightParen
	LeftBrac
	RightBrac
	Quote
	Colon
	Comma
	Semicolon
	Period
	Comment
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// lexer holds the state of the lexer.
type Lexer struct {
	Tokens chan Token // channel of scanned items
	r      io.ByteReader
	done   bool
	name   string // the name of the input; used only for error reports
	buf    []byte
	input  string  // the line of text being scanned.
	state  stateFn // the next lexing function to enter
	line   int     // line number in input
	pos    int     // current position in the input
	start  int     // start position of this item
	width  int     // width of last rune read from input
}

func (l *Lexer) Pos() int {
	return l.pos
}
func New(name string, r io.ByteReader) *Lexer {
	l := &Lexer{
		r:      r,
		name:   name,
		line:   1,
		Tokens: make(chan Token),
	}
	go l.run()
	return l

}

// errorf returns an error token and continues to scan.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.Tokens <- Token{Error, l.start, []byte(fmt.Sprintf(format, args...))}
	return lexAny
}

// run runs the state machine for the Scanner.
func (l *Lexer) run() {
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.Tokens)
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if !l.done && int(l.pos) == len(l.input) {
		l.loadLine()
	}
	if len(l.input) == l.start {
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *Lexer) emit(t Type) {
	if t == Newline {
		l.line++
	}
	s := l.input[l.start:l.pos]
	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%s:%d: emit %s\n", l.name, l.line, Token{t, l.line, []byte(s)})
	}
	if t != Newline {
		l.Tokens <- Token{t, l.line, []byte(s)}
	}
	l.start = l.pos
	l.width = 0
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// loadLine reads the next line of input and stores it in (appends it to) the input.
// (l.input may have data left over when we are called.)
// It strips carriage returns to make subsequent processing simpler.
func (l *Lexer) loadLine() {
	l.buf = l.buf[:0]
	for {
		c, err := l.r.ReadByte()
		if err != nil {
			l.done = true
			break
		}
		if c != '\r' {
			l.buf = append(l.buf, c)
		}
		if c == '\n' {
			break
		}
	}
	l.input = l.input[l.start:l.pos] + string(l.buf)
	l.pos -= l.start
	l.start = 0
}

func lexAny(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return nil
		case r == '(':
			l.emit(LeftParen)
			return lexAny
		case unicode.IsDigit(r):
			return lexNumber
		case unicode.IsLetter(r):
			return lexAlphaNumeric
		case r == '#':
			return lexComment
		case r == '{':
			l.emit(LeftCurly)
			return lexAny
		case r == '[':
			l.emit(LeftBrac)
			return lexAny
		case r == '(':
			l.emit(LeftCurly)
			return lexAny
		case r == ')':
			l.emit(RightParen)
			return lexAny
		case r == '}':
			l.emit(RightCurly)
			return lexAny
		case r == ']':
			l.emit(RightBrac)
			return lexAny
		case r == ':':
			l.emit(Colon)
			return lexAny
		case r == '.':
			l.emit(Period)
			return lexAny
		case r == ',':
			l.emit(Comma)
			return lexAny
		case r == '"':
			return lexQuote
		case isEndOfLine(r):
			l.emit(Newline)
			return lexAny
		case isSpace(r):
			return lexSpace

		}
	}
	return nil
}
func lexQuote(l *Lexer) stateFn {
	for isValidQuote(l.peek()) {
		l.next()
	}
	if r := l.next(); r == '"' {
		l.emit(Quote)
	} else {
		l.errorf("Unexpected character inside quote in position %d:%d character %q.",
			l.line,
			l.pos,
			r)
	}

	return lexAny
}

func lexComment(l *Lexer) stateFn {
	for !isEndOfLine(l.peek()) {
		l.next()
	}
	l.emit(Comment)
	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexAlphaNumeric(l *Lexer) stateFn {
	for isAlphaNumeric(l.peek()) {
		l.next()
	}
	l.emit(String)
	return lexAny
}

func lexNumber(l *Lexer) stateFn {
	for unicode.IsDigit(l.peek()) {
		l.next()
	}
	l.emit(Number)
	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *Lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexAny
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isString(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return isString(r) || unicode.IsDigit(r)
}

func isValidQuote(r rune) bool {
	return isAlphaNumeric(r) || isSpace(r)
}
