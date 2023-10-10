// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: remove-anonymous-lambda : L8 -> L9
//
// since we are generating a C function for each Scheme lambda, we need to
// have a name for each of these lambdas.  In addition we need a name to use
// as the code pointer label, so that we can lift the lambdas to the top
// level of the program.  The transformation is fairly simple.  If we find a
// lambda in expression position (i.e. not in the right-hand side of a
// letrec binding) then we wrap a letrec around it that gives it a new name.
//
//	(letrec ([l* (lambda (x** ...) body*)] ...) body) => (no change)
//	(letrec ([l* (lambda (x** ...) body*)] ...) body)
//
//	(lambda (x* ...) body) => (letrec ([t0 (lambda (x* ...) body)]) t0)
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L8"
	"github.com/mdempsky/hermes/example/lang/L9"
)

func Entry(e L8.Expr) L9.Expr { return nil }

func Lambda(params []L9.Symbol, body L9.AssignedBody) L9.Expr {
	tmp := builtin.Fresh[L9.Symbol]()
	return L9.LetRec{
		Bindings: []L9.RecBinding{{Var: tmp, Val: L9.Lambda{Params: params, Body: body}}},
		Body:     tmp,
	}
}
