// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"log"
	"testing"

	"sevki.org/lib/prettyprint"
)

func TestSimpleQuery(t *testing.T) {
	t.Parallel()
	const query = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend
  }
}`
	if ast, err := NewQuery([]byte(query)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}

}
func TestComplexQuery(t *testing.T) {
	t.Parallel()
	const query = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend,
    profilePicture(userid: "bla", size: 50) {
      uri,
      width,
      height
    }
  }
}`
	if ast, err := NewQuery([]byte(query)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}
}
