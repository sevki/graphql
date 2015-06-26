// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"log"
	"testing"

	"sevki.org/graphql/query"
	"sevki.org/lib/prettyprint"
)

func TestSimpleQuery(t *testing.T) {
	t.Parallel()
	const qry = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend
  }
}`
	if ast, err := NewQuery([]byte(qry)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}

}
func TestComplexQuery(t *testing.T) {
	t.Parallel()
	const qry = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend,
    profilePicture(userid: "bla", size: 11) {
      uri,
      width,
      height
    }
  }
}`
	if ast, err := NewQuery([]byte(qry)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}
}
func TestCommentQuery(t *testing.T) {
	t.Parallel()
	const qry = `{
  user(id: 3500401) {
    id,
    name,
    isViewerFriend,
    # something interesting here 
    profilePicture(userid: "bla", size: 22) {
      uri,
      width,
      height
    }
  }
}`
	if ast, err := NewQuery([]byte(qry)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}
}
func TestVariabledQuery(t *testing.T) {
	t.Parallel()
	const qry = `{
  # asdasd
  user(id: 3500401) {
    id,
    name,
    isViewerFriend,
    # something interesting here 
    profilePicture(userid: $id, size: 33) {
      uri,
      width,
      height
    }
  }
}`
	if ast, err := NewQuery([]byte(qry)); err != nil {
		t.Error(err.Error())
	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}
}

func TestApplyContext(t *testing.T) {
	t.Parallel()
	const qry = `{
  user(id: $id) {
    id,
    name,
    isViewerFriend
  }
}`
	if ast, err := NewQuery([]byte(qry)); err != nil {
		t.Error(err.Error())
		ctx := make(map[string]interface{})
		ctx["id"] = 12
		_, err := query.ApplyContext(*ast, ctx)
		if err != nil {
			t.Error(err)
		}

	} else {
		log.Printf(prettyprint.AsJSON(*ast))
	}

}
