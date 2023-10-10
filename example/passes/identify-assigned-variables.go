// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: identify-assigned-variables : L6 -> L7
//
// This pass identifies which variables are assigned using set!. This
// is the first step in a process known as assignment conversion. We
// separate assigned varaibles from unassigned variables, and assigned
// variables are converted into reference cells that can be
// manipulated through primitives. In this compiler, we use the
// existing box type to create the cells (using the box primitive),
// reference the cells (using the unbox primitive), and mutating the
// cells (using the set-box! primitive). One of the reasons we perform
// assignment conversion is it allows multiple closures to capture the
// same mutable variable and all of the closures will see the same,
// up-to-date, value for that variable since they all simply contain a
// pointer to the reference cell. If we didn't do this conversion, we
// would need to figure out a different way to handle set! so that the
// updates are propagated to all the closures that close over the
// variable. The eventual effect of assignemnt conversion is the
// following:
//
//	(let ([x 5])
//	  (set! x (+ x 5))
//	  (+ x x)) =>
//	(let ([t0 5])
//	  (let ([x (box t0)])
//	    (primcall set-box! x (+ (unbox x) 5)
//	    (+ (unbox x) (unbox x))))
//
// (of course in this example, we could have simply shadowed x)
//
// This pass, however, is simply an analysis pass.  It gathers up the set of
// assigned variables and deposits them in an AssignedBody just inside their
// binding points.  The transformation in this pass is:
//
//	(let ([x 5] [y 7] [z 10])
//	  (set! x (+ x y))
//	  (+ x z)) =>
//	(let ([x 5] [y 7] [z 10])
//	  (assigned (x)
//	    (set! x (+ x y))
//	    (+ x z)))
//
// The key operations we depend on are:
//
//   - set-cons: to extend a set with a newly found assigned variable.
//   - intersect: to determine which assigned variables are bound by a
//     lambda, let, or letrec.
//   - difference: to remove assigned variables from a set once we
//     locate their binding form.
//   - union: to gather assigned variables from sub-expressions into a
//     single set.
//
// Note: we are using a relatively inefficient representation of sets
// here, simply representing them as lists and using our set-cons,
// intersect, difference, and union procedures to maintain their
// set-ness.  We could choose a more efficient set representation,
// perhaps leveraging insertion sort or something similar, or we could
// choose to represent our variables using a mutable record, with a
// field to indicate if it is assigned.  Either approach will improve
// the worst case performance of this pass, though the mutable record
// version will get us down to a linear cost, which is the best case
// for any pass in the current version of the nanopass framework.
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L6"
	"github.com/mdempsky/hermes/example/lang/L7"
)

func Entry(e0 L6.Expr) L7.Expr {
	e := Expr(e0)
	if !e.free.Empty() {
		builtin.Errorf(e.x, "found one or more unbound variables: %v", e.free)
	}
	return e.x
}

type expr struct {
	x    L7.Expr
	free builtin.Set[L7.Symbol] // aggregate with union
}

func Expr(e L6.Expr) expr {
	return expr{}
}

func Set(lhs L7.Symbol, rhs expr) expr {
	return expr{
		x:    rhs.x,
		free: builtin.NewSet(lhs).Union(rhs.free),
	}
}

func Lambda(params []L7.Symbol, body expr) expr {
	free, abody := bind(builtin.NewSet(params...), body)
	return expr{
		x: L7.Lambda{
			Params: params,
			Body:   abody,
		},
		free: free,
	}
}

func Let(bindings []L7.Binding, body expr) expr {
	free, abody := bind(boundSet(bindings), body)
	return expr{
		x: L7.Let{
			Bindings: bindings,
			Body:     abody,
		},
		free: free,
	}
}

func LetRec(bindings []L7.Binding, body expr) expr {
	free, abody := bind(boundSet(bindings), body)
	return expr{
		x: L7.LetRec{
			Bindings: bindings,
			Body:     abody,
		},
		free: free,
	}
}

func bind(bound builtin.Set[L7.Symbol], body expr) (free builtin.Set[L7.Symbol], abody L7.AssignedBody) {
	free = body.free.Difference(bound)
	abody = L7.AssignedBody{
		Names: builtin.Sorted(body.free.Intersect(bound)),
		Body:  body.x,
	}
	return
}

func boundSet(bindings []L7.Binding) builtin.Set[L7.Symbol] {
	return builtin.NewSet(builtin.Map(bindings, func(binding L7.Binding) L7.Symbol { return binding.Var })...)
}
