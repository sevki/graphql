// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lexer // import "graphql.co/lexer"

import (
	"fmt"
	"os"
	"testing"

	"graphql.co/token"
)

func TestQuery(t *testing.T) {
	ks, _ := os.Open("../tests/complex-as-possible.graphql")
	l := New("sq", ks)
	for {
		tok := <-l.Tokens
		if tok.Type != token.Newline {
			fmt.Printf("%s => %s\n", tok.Type, tok.Text)
		}
		if tok.Type == token.EOF || tok.Type == token.Error {
			break
		}
	}

}
