// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: convert-assignments : L9 -> L10
//
// This pass completes the assignment conversion process that we
// started in identify-assigned-variables. We use the assigned
// variable list through our previous passes to make decisions about
// how bindings were separated. Now, we are ready to change these
// explicitly to the box, unbox, and set-box! primitive calls
// described in the identified-assigned-variable pass. We also
// introduce new temporaries to contain the value that will be put in
// the box. This is largely because we don't want our representation
// of assigned variables to be observable from inside the program, and
// if we were to introduce an operator like call/cc to our
// implementation, then the order our variables were setup could
// potentially be identified by seeing that the allocation and
// computation of the values are intermixed. Instead, we want all the
// computation to happen, then the allocation, and then the allocated
// locations are updated with the values.
//
// Our transform thus looks like the following:
//
//	(let ([x0 e0] [x1 e1] ... [xa0 ea0] [xa1 xa0] ...)
//	  (assigned (xa0 xa1 ...)
//	    body))
//	=>
//	(let ([x0 e0] [x1 e1] ... [t0 ea0] [t1 ea1] ...)
//	  (let ([xa0 (primcall box t0)] [xa1 (primcall box t1)] ...)
//	    body^))
//
//	(lambda (x0 x1 ... xa0 xa1 ...) (assigned (xa0 xa1 ...) body))
//	=>
//	(lambda (x0 x1 ... t0 t1 ...)
//	  (let ([xa0 (primcall box t0)] [xa1 (primcall box t1)] ...)
//	    body^))
//
// where
//
//	(set! xa0 e) => (primcall set-box! xa0 e^)
//
// and
//
//	xa0 => (primcall unbox xa0)
//
// in body^ and e^.
//
// We could choose another data structure, or even create a new data
// structure to perform the conversion with, however, we've choosen
// the box because it contains exactly one cell, and takes up just one
// word in memory, where as our pair and vector take at least two
// words in memory. This decision might be different if we had other
// constraints on how we lay out memory.
package pass

import (
	"github.com/mdempsky/hermes/example/lang/L10"
	"github.com/mdempsky/hermes/example/lang/L9"
)

const Unbox = 100

func Entry(e L9.Expr) L10.Expr { return nil }

type env struct {
	box map[L10.Symbol]L10.Symbol
}

func (env) Expr(e L9.Expr) L10.Expr {
	return nil
}

func (env env) Symbol(x L10.Symbol) L10.Expr {
	if box, ok := env.box[x]; ok {
		return L10.PrimCall{Prim: Unbox, Args: []L10.Expr{box}}
	}
	return x
}

// TODO(mdempsky): This isn't done.

func (env env) Let(bindings []L10.Binding, abody L9.AssignedBody) L10.Expr {
	return L10.Let{
		Bindings: bindings,
		Body:     env.Expr(abody.Body),
	}
}

func (env env) Lambda(params []L10.Symbol, abody L9.AssignedBody) L10.LambdaExpr {
	return L10.Lambda{
		Params: params,
		Body:   env.Expr(abody.Body),
	}
}
