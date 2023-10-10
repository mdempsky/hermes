// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: remove-one-armed-if : Lsrc -> L1
//
// This pass replaces the (if e0 e1) form with an if that will
// explicitly produce a void value when the predicate expression
// returns false. In other words:
//
//	(if e0 e1) => (if e0 e1 (void))
//
// Design descision: kept seperate from parse-and-rename to make it
// easier to understand how the nanopass framework can be used.
package pass

import (
	"github.com/mdempsky/hermes/example/lang/L1"
	"github.com/mdempsky/hermes/example/lang/Lsrc"
)

const VOID = 42

func Entry(Lsrc.Expr) L1.Expr { return nil }

func IfThen(Cond, Then L1.Expr) L1.Expr {
	return L1.If{Cond: Cond, Then: Then, Else: L1.Primitive(VOID)}
}
