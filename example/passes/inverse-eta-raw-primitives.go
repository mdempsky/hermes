// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass : inverse-eta-raw-primitives : L3 -> L4
//
// Eta reduction recognizes a function that takes a set of arguments and
// passes those arguments directly to another function, and unwraps the
// function.  For instance, the function:
//
//	(lambda (x y) (f x y))
//
// can be eta reduced to:
//
//	f
//
// Eta reduction is not always a semantics preserving transformation because
// it can change the termination properties of the program, for instance a
// program that terminates, could turn into one that does not because a
// function is applied directly, rather than a function that might never be
// applied.
//
// In this pass, we are applying the inverse operation and adding a lambda
// wrapper when we see a primitive.  We do this so that primitives, which we
// are going to open code into a C-code equivalent, can still be treated as
// though it was a Scheme procedure.  This allows us to map over primitives,
// which would otherwise not be possible with our code generation.  Our
// transformation looks for primitives in call position, marking them as
// primitive calls, and primitives not in call position are eta-expanded to move
// them into call position.
//
//	(pr e* ...) => (primcall pr e* ...)
//	pr          => (lambda (x* ...) (primcall pr x* ...))
//
// Design decision: Another way to handle this would be to create a single
// function for each primitive, and lift these definitions to the top-level
// of the program, including just those primitives that are used.  This
// would avoid the potential to re-creating the same procedure over and over
// again, as we are now.
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L3"
	"github.com/mdempsky/hermes/example/lang/L4"
)

func Entry(e L3.Expr) L4.Expr { return nil }

func Expr(e L3.Expr) L4.Expr {
	switch e := e.(type) {
	case struct {
		L3.Apply
		Fun  L4.Primitive
		Args []L4.Expr
	}:
		return L4.PrimCall{Prim: e.Fun, Args: e.Args}
	case L3.Primitive:
		panic(builtin.Errorf(e, "unexpected primitive: %v", e))
	}
	return nil
}
