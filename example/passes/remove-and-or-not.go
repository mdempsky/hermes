// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: remove-and-or-not : L1 -> L2
//
// This pass looks for references to and, or, and not and replaces it
// with the appropriate if expressions. This pass follows the standard
// expansions and has one small optimization:
//
//	(if (not e0) e1 e2) => (if e0 e2 e1)           [optimization]
//	(and)               => #t                      [from Scheme standard]
//	(or)                => #f                      [from Scheme standard]
//	(and e e* ...)      => (if e (and e* ...) #f)  [standard expansion]
//	(or e e* ...)       => (let ([t e])            [standard expansion -
//
//	(if t t (or e* ...)))  avoids computing e twice]
//
// Design decision: again kept separate from parse-and-rename to
// simplify the discussion of this pass (adding it to parse-and-rename
// doesn't really make parse-and-rename much more complicated, and if
// we had a macro system, which would likely be implemented in
// parse-and-rename, or before it, we would probably want and, or, and
// not defined as macros, rather than forms in the language, in which
// case this pass would be unnecessary).
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L1"
	"github.com/mdempsky/hermes/example/lang/L2"
)

func Entry(e L1.Expr) L2.Expr { return nil }

func Expr(e L1.Expr) L2.Expr {
	switch e := e.(type) {
	case struct {
		L1.If
		Cond struct {
			L1.Not
			X L2.Expr
		}
		Then, Else L2.Expr
	}:
		return L2.If{Cond: e.Cond.X, Then: e.Else, Else: e.Then}
	}

	return nil
}

func Not(x L2.Expr) L2.Expr {
	return L2.If{Cond: x, Then: L2.False{}, Else: L2.True{}}
}

func And(x []L2.Expr) L2.Expr {
	return builtin.FoldRight(x, L2.Expr(L2.True{}), func(first, rest L2.Expr) L2.Expr {
		return L2.If{Cond: first, Then: rest, Else: L2.False{}}
	})
}

func Or(x []L2.Expr) L2.Expr {
	return builtin.FoldRight(x, L2.Expr(L2.False{}), func(first, rest L2.Expr) L2.Expr {
		tmp := builtin.Fresh[L2.Symbol]()
		return L2.Let{
			Bindings: []L2.Binding{{Var: tmp, Val: first}},
			Body:     L2.If{Cond: tmp, Then: tmp, Else: rest},
		}
	})
}
