// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lexer // import "sevki.org/graphql/lexer"

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"sevki.org/graphql/token"
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// lexer holds the state of the lexer.
type Lexer struct {
	Tokens chan token.Token // channel of scanned items
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

func (l *Lexer) LineBuffer() string {
	return string(l.buf)
}
func (l *Lexer) Pos() int {
	return l.pos
}
func New(name string, r io.Reader) *Lexer {

	l := &Lexer{
		r:      bufio.NewReader(r),
		name:   name,
		line:   1,
		Tokens: make(chan token.Token),
	}
	go l.run()
	return l

}

// errorf returns an error token and continues to scan.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.Tokens <- token.Token{token.Error, l.start, []byte(fmt.Sprintf(format, args...)), l.start, l.pos}
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

func (l *Lexer) emit(t token.Type) {
	if t == token.Newline {
		l.line++
	}
	s := l.input[l.start:l.pos]
	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%s:%d: emit %s\n", l.name, l.line, token.Token{t, l.line, []byte(s), l.start, l.pos})
	}
	if t != token.Newline {
		l.Tokens <- token.Token{
			t,
			l.line,
			[]byte(s),
			l.start,
			l.pos,
		}
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
			l.emit(token.LeftParen)
			return lexAny
		case unicode.IsDigit(r):
			return lexNumber
		case unicode.IsLetter(r):
			return lexAlphaNumeric
		case r == '$':
			return lexVariable
		case r == '#':
			return lexComment
		case r == '{':
			l.emit(token.LeftCurly)
			return lexAny
		case r == '[':
			l.emit(token.LeftBrac)
			return lexAny
		case r == '(':
			l.emit(token.LeftParen)
			return lexAny
		case r == ')':
			l.emit(token.RightParen)
			return lexAny
		case r == '}':
			l.emit(token.RightCurly)
			return lexAny
		case r == ']':
			l.emit(token.RightBrac)
			return lexAny
		case r == ':':
			l.emit(token.Colon)
			return lexAny
		case r == '|':
			l.emit(token.Pipe)
			return lexAny
		case r == '.':
			return lexPeriodOrElipsis
		case r == ',':
			l.ignore()
			return lexAny
		case r == '=':
			l.emit(token.Equal)
			return lexAny
		case r == '"':
			return lexQuote
		case r == '@':
			return lexDirective
		case isEndOfLine(r):
			l.emit(token.Newline)
			return lexAny
		case isSpace(r):
			return lexSpace

		}
	}

}

func lexPeriodOrElipsis(l *Lexer) stateFn {
	if l.peek() != '.' {
		l.emit(token.Period)
		return lexAny
	} else {
		l.next()
		if r := l.next(); r != '.' {
			l.errorf("Unexpected character inside period or elipsis in position %d:%d character %q.",
				l.line,
				l.pos,
				r)

		}
		l.emit(token.Elipsis)
		return lexAny
	}
}

func lexQuote(l *Lexer) stateFn {
	for l.peek() != '"' {
		l.next()
	}
	if r := l.next(); r == '"' {
		l.emit(token.Quote)
	} else {
		l.errorf("Unexpected character inside quote in position %d:%d character %q.",
			l.line,
			l.pos,
			r)
	}

	return lexAny
}

func lexVariable(l *Lexer) stateFn {
	l.ignore()
	for isAlphaNumeric(l.peek()) {
		l.next()
	}
	l.emit(token.Variable)
	return lexAny
}

func lexComment(l *Lexer) stateFn {
	for !isEndOfLine(l.peek()) {
		l.next()
	}
	//	l.emit(token.Comment)
	l.ignore()
	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexAlphaNumeric(l *Lexer) stateFn {
	for isAlphaNumeric(l.peek()) {
		l.next()
	}
	switch l.input[l.start:l.pos] {
	case "fragment":
		l.emit(token.FragmentStart)
		break
	case "query":
		l.emit(token.QueryStart)
		break
	case "mutation":
		l.emit(token.MutationStart)
		break
	case "on":
		l.emit(token.On)
		break
	case "true":
		l.emit(token.True)
		break
	case "false":
		l.emit(token.False)
		break
	default:
		l.emit(token.String)
	}

	return lexAny
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexDirective(l *Lexer) stateFn {
	l.ignore()
	for isAlphaNumeric(l.peek()) {
		l.next()
	}
	if r := l.peek(); r == '(' || r == ' ' {
		l.emit(token.Directive)
		l.ignore()
	} else {
		l.errorf("Unexpected character inside directive in position %d:%d character %q.",
			l.line,
			l.pos,
			r)
	}
	return lexAny
}

func lexNumber(l *Lexer) stateFn {
	emitee := token.Number
	for isValidNumber(l.peek()) {
		switch l.next() {
		case '.':
			emitee = token.Float
		case 'x':
			return lexHex
		}
	}
	l.emit(emitee)
	return lexAny
}

// lexSpace lexes a hexadecimal
func lexHex(l *Lexer) stateFn {
	for isValidHex(l.peek()) {
		l.next()
	}
	l.emit(token.Hex)
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
func isValidNumber(r rune) bool {
	return unicode.IsDigit(r) ||
		r == '-' ||
		r == '.' ||
		r == 'x'

}
func isValidHex(r rune) bool {
	return unicode.IsDigit(r) ||
		r == 'x' ||
		r == 'X' ||
		r == 'A' ||
		r == 'a' ||
		r == 'B' ||
		r == 'b' ||
		r == 'c' ||
		r == 'C' ||
		r == 'd' ||
		r == 'D' ||
		r == 'e' ||
		r == 'E' ||
		r == 'f' ||
		r == 'F'
}
