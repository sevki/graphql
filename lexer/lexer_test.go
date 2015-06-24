// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lexer // import "sevki.org/graphql/lexer"

import (
	"bytes"
	"fmt"
	"testing"
)

const testQuery = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend,
    category(name: "best") {
      id,
      description
    },
    firends.Limit(20) {
      id
      user(id: $id) {
         name
      }
    }
    # This is tircky
    profilePicture(size: 50)  {
      uri,
      width,
      height
    }
  }
}`

func TestQuery(t *testing.T) {
	sq := bytes.NewBuffer([]byte(testQuery))
	l := New("sq", sq)
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
