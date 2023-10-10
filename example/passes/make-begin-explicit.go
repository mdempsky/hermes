// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: make-being-explicit : L2 -> L3
//
// This pass takes the L2 let, letrec, and lambda expressions (which
// have bodies that can contain more than one expression), and
// converts them into bodies with a single expression, wrapped in a
// begin if necessary. To avoid polluting the output with extra
// begins that contain only one expression the build-begin helper
// checks to see if there is more then one expression and if there is
// builds a begin.
//
// Effectively this does the following:
//
//	(let ([x* e*] ...) body0 body* ... body1) =>
//	  (let ([x* e*] ...) (begin body0 body* ... body1))
//	(letrec ([x* e*] ...) body0 body* ... body1) =>
//	  (letrec ([x* e*] ...) (begin body0 body* ... body1))
//	(lambda (x* ...) body0 body* ... body1) =>
//	  (lambda (x* ...) (begin body0 body* ... body1))
//
// Design Decision: This could have been included with
// rename-and-parse, without making it significantly more
// compilicated, but was separated out to continue with simpler
// nanopass passes to help make it more obvious what is going on here.
package pass

import (
	"github.com/mdempsky/hermes/example/lang/L2"
	"github.com/mdempsky/hermes/example/lang/L3"
)

func Entry(e L2.Expr) L3.Expr { return nil }

func Let(bindings []L3.Binding, init []L3.Expr, body L3.Expr) L3.Expr {
	return L3.Let{Bindings: bindings, Body: begin(init, body)}
}
func LetRec(bindings []L3.Binding, init []L3.Expr, body L3.Expr) L3.Expr {
	return L3.Let{Bindings: bindings, Body: begin(init, body)}
}
func Lambda(params []L3.Symbol, init []L3.Expr, body L3.Expr) L3.Expr {
	return L3.Lambda{Params: params, Body: begin(init, body)}
}

func begin(init []L3.Expr, body L3.Expr) L3.Expr {
	if len(init) == 0 {
		return body
	}
	return L3.Begin{Init: init, Body: body}
}
