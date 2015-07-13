// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lexer // import "sevki.org/graphql/lexer"

import (
	"fmt"
	"os"
	"testing"
)

func TestQuery(t *testing.T) {
	ks, _ := os.Open("../tests/complex-as-possible.graphql")
	l := New("sq", ks)
	for {
		token := <-l.Tokens
		if token.Type != Newline {
			fmt.Printf("%s => %s\n", token.Type, token.Text)
		}
		if token.Type == EOF || token.Type == Error {
			break
		}
	}

}
