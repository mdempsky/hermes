// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: remove-complex-constants : L5 -> L6
//
// Lifts creation of constants composed of vectors or pairs outside
// the body of the program and makes their creation explicit.  In
// place of the constants, a temporary variable reference is created.
// The total transform looks something like the following:
//
//	(letrec ([add-pair-parts (lambda (p) (+ (car p) (cdr p)))])
//	  (+ (add-pair-parts '(4 . 5)) (add-pair-parts '(6 .7)))) =>
//	(let ([t0 (cons 4 5)] [t1 (cons 6 7)])
//	  (letrec ([add-pair-parts (lambda (p) (+ (car p) (cdr p)))])
//	    (+ (add-pair-parts t0) (add-pair-parts t1))))
//
// Design decision: Another possibility is to simply convert the
// constants into their memory-layed out variations, rather than
// treating it in pieces like this.  We could extend our C run-time
// support to know about these pre-layed out items so that we do not
// need to construct them when the program starts running.
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L5"
	"github.com/mdempsky/hermes/example/lang/L6"
)

const (
	Cons = 100 + iota
	MakeVector
	VectorSet
)

type expr struct {
	x    L6.Expr
	data builtin.List[L6.Binding]
}

func Entry(e0 L5.Expr) L6.Expr {
	e := Expr(e0)
	return let(e.data.Slice(), e.x)
}

func Expr(e L5.Expr) expr {
	switch e := e.(type) {
	case L5.Quote:
		switch x := e.X.(type) {
		case L5.Const:
			return quote(x)
		}
		x := Datum(e.X)
		tmp := builtin.Fresh[L6.Symbol]()
		return expr{
			x:    tmp,
			data: builtin.Cons(L6.Binding{Var: tmp, Val: x.x}, x.data),
		}
	}
	return expr{}
}

func Datum(e L5.Datum) expr {
	switch e := e.(type) {
	case L5.Const:
		return quote(e)
	}
	return expr{}
}

func Pair(car, cdr L6.Expr) expr {
	return expr{
		x: L6.PrimCall{Prim: Cons, Args: []L6.Expr{car, cdr}},
	}
}

func Vector(elems []L6.Expr) expr {
	tmp := builtin.Fresh[L6.Symbol]()
	return expr{
		x: L6.Let{
			Bindings: []L6.Binding{{Var: tmp, Val: L6.PrimCall{Prim: MakeVector, Args: []L6.Expr{L6.Quote{X: L6.Int{len(elems)}}}}}},
			Body: L6.Begin{
				Init: builtin.MapIndex(elems, func(i int, elem L6.Expr) L6.Expr {
					return L6.PrimCall{Prim: VectorSet, Args: []L6.Expr{tmp, L6.Quote{L6.Int{i}}, elem}}
				}),
				Body: tmp,
			},
		},
	}
}

func let(bindings []L6.Binding, body L6.Expr) L6.Expr {
	if len(bindings) == 0 {
		return body
	}
	return L6.Let{Bindings: bindings, Body: body}
}

func quote(x L5.Const) expr {
	return expr{
		x: L6.Quote{X: x.(L6.Const)},
	}
}
