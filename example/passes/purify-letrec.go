// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: purify-letrec : L7 -> L8
//
// This pass looks for places where letrec is used to bind something
// other than a lambda expression, or where a letrec bound variable is
// assigned and moves these bindings into let bindings. When the pass
// is done all of the letrecs in our program will be immutable and
// bind only lambda expressions. For instance, the following example:
//
//	(letrec ([f (lambda (g x) (g x))]
//	         [a 5]
//	         [b (+ 5 7)]
//	         [g (lambda (h) (f h 5))]
//	         [c (let ([x 10]) ((letrec ([zero? (lambda (n) (= n 0))]
//	                                    [f (lambda (n)
//	                                         (if (zero? n)
//	                                             1
//	                                             (* n (f (- n 1)))))])
//	                             f)
//	                            x))]
//	         [m 10]
//	         [z (lambda (x) x)])
//	  (set! z (lambda (x) (+ x x)))
//	  (set! m (+ m m))
//	  (+ (+ (+ (f z a) (f z b)) (f z c)) (g z))))
//	=>
//	(let ([z (quote #f)] [m '#f] [c '#f])
//	  (let ([b (+ '5 '7)] [a '5])
//	    (letrec ([g (lambda (h) (f h '5))]
//	             [f (lambda (g x) (g x))])
//	      (begin
//	        (set! z (lambda (x) x))
//	        (set! m '10)
//	        (set! c
//	          (let ([x '10])
//	            ((letrec ([f (lambda (n)
//	                              (if (zero? n)
//	                                  '1
//	                                  (* n (f (- n '1)))))]
//	                      [zero? (lambda (n) (= n '0))])
//	               f)
//	              x)))
//	        (begin
//	          (set! z (lambda (x) (+ x x)))
//	          (set! m (+ m m))
//	          (+ (+ (+ (f z a) (f z b)) (f z c)) (g z)))))))
//
// The algorithm for doing this is fairly simple.  We attempt to separate
// the bindings into simple bindings, lambda bindings, and complex bindings.
// Simple bindings bind a constant, a variable reference not bound in this
// letrec, the call to an effect free primitive, a begin that contains only
// simple expressions, or an if that contains only simple expressions to an
// immutable variable. The simple? predicate is used for determining when an
// expression is simple.  A lambda expression is simply a lambda, and a
// complex expression is any other expression.
//
// Design decision: There are many other approaches that we could use,
// including those described in the "Fixing Letrec: A Faithful Yet Efficient
// Implementation of Scheme’s Recursive Binding Construct" by Waddell, et.
// al. and "Fixing Letrec (reloaded)" by Ghuloum and Dybvig.  Earlier
// versions of Chez Scheme used the earlier paper, which documented how to
// properly handle R5RS letrecs, and newer versions use the latter paper
// which described how to properly handle R6RS letrec and letrec*.
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L7"
	"github.com/mdempsky/hermes/example/lang/L8"
)

func Entry(e L7.Expr) L8.Expr { return nil }

func LetRec(bindings []L8.Binding, body L8.AssignedBody) L8.Expr {
	// classify bindings as simple/lambda/complex.

	assigned := builtin.NewSet(body.Names...)

	type sort struct {
		simple, complex builtin.List[L8.Binding]
		lambdas         builtin.List[L8.RecBinding]
	}
	sorts := builtin.Sum(bindings, func(binding L8.Binding) sort {
		if !assigned.Has(binding.Var) {
			switch e := binding.Val.(type) {
			case L8.Lambda:
				return sort{
					lambdas: builtin.ListOf(L8.RecBinding{Var: binding.Var, Val: e}),
				}
			}
			// TODO(mdempsky): Recognize "simple" (side-effect-free)
			// expressions too: Quote, Symbols that are neither assigned nor
			// being bound here, PrimCall with an effect-free primitive, and
			// Begin and If that are recursively free of side effects.
		}

		return sort{
			complex: builtin.ListOf(binding),
		}
	})

	return L8.Let{
		Bindings: builtin.Map(sorts.complex.Slice(), func(binding L8.Binding) L8.Binding {
			return L8.Binding{
				Var: binding.Var,
				Val: L8.Quote{X: L8.False{}},
			}
		}),
		Body: L8.AssignedBody{
			Body: L8.Let{
				Bindings: sorts.simple.Slice(),
				Body: L8.AssignedBody{
					Body: L8.LetRec{
						Bindings: sorts.lambdas.Slice(),
						Body: L8.Begin{
							Init: builtin.MapIndex(sorts.complex.Slice(), func(i int, binding L8.Binding) L8.Expr {
								return L8.Set{Var: binding.Var, Val: binding.Val}
							}),
							Body: body.Body,
						},
					},
				},
			},
		},
	}
}
