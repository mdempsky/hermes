// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// pass: quote-constants : L4 -> L5
//
// A simple pass to find raw constants and wrap them in a quote.
//
//	c => (quote c)
//
// Design decision: This could simply be included in the next pass.
package pass

import (
	"github.com/mdempsky/hermes/example/lang/L4"
	"github.com/mdempsky/hermes/example/lang/L5"
)

func Entry(e L4.Expr) L5.Expr { return nil }

func Expr(e L4.Expr) L5.Expr {
	switch e := e.(type) {
	case L4.Const:
		return L5.Quote{X: e.(L5.Const)}
	}
	return nil
}
