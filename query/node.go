// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type CallType

package query

type CallType int

const (
	Select CallType = iota
	Insert
	Update
	Delete
	Limit
	Having
	OrderBy
	GroupBy
	Lock
)

type Call struct {
	Type   CallType
	Name   []byte
	Params []Param
	Call   *Call
}

type Node struct {
	Name   []byte
	Parent *Node `json:"-"`
	Edges  []Node
	Call   *Call
}
