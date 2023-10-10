// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: optimize-direct-call : L8 -> L8
//
// One of our simplest optimizations, we convert a directly applied
// lambdas into a let. This allows us to avoid the creation of a
// closure for the let, and allows us instead to create a local
// binding within a function. The transform is simple:
//
//	((lambda (x* ...) body) e* ...) => (let ([x* e*] ...) body)
//	where (length x*) == (length e*)
package pass

import (
	"github.com/mdempsky/hermes/builtin"
	"github.com/mdempsky/hermes/example/lang/L8"
)

func Entry(e L8.Expr) L8.Expr { return nil }

func Expr(e L8.Expr) L8.Expr {
	switch e := e.(type) {
	case struct {
		L8.Apply
		Fun L8.Lambda
	}:
		if bindings, ok := builtin.Zip(e.Fun.Params, e.Args, bind); ok {
			return L8.Let{Bindings: bindings, Body: e.Fun.Body}
		}
	}
	return nil
}

func bind(param L8.Symbol, arg L8.Expr) L8.Binding {
	return L8.Binding{Var: param, Val: arg}
}
