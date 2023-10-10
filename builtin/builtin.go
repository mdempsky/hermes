// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package builtin provides the Hermes builtin functions.
package builtin

import "cmp"

// Fresh returns a new, unique identifier.
func Fresh[ID any]() ID { panic(0) }

// FoldRight returns fn(s[len-1], fn(s[len-2], ... fn(s[0], init))).
func FoldLeft[In, Out any](s []In, init Out, fn func(In, Out) Out) Out { panic(0) }

// FoldRight returns fn(s[0], fn(s[1], ... fn(s[len-1], init))).
func FoldRight[In, Out any](s []In, init Out, fn func(In, Out) Out) Out { panic(0) }

// Map returns []Out{fn(s[0]), fn(s[1]), ..., fn(s[len-1])}.
func Map[In, Out any]([]In, func(In) Out) []Out { panic(0) }

// MapIndex returns []Out{fn(0, s[0]), fn(1, s[1]), ..., fn(len-1, s[len-1])}.
func MapIndex[In, Out any]([]In, func(int, In) Out) []Out { panic(0) }

// Errorf reports an error at the given position.
func Errorf(where any, msg string, args ...any) any { panic(0) }

type List[Elem any] opaque

func ListOf[Elem any](elems ...Elem) List[Elem]            { panic(0) }
func Cons[Elem any](head Elem, tail List[Elem]) List[Elem] { panic(0) }
func (List[Elem]) Slice() []Elem                           { panic(0) }

// TODO(mdempsky): Out must be a semigroup.
func Sum[In, Out any]([]In, func(In) Out) Out { panic(0) }

type Set[Elem comparable] opaque

func NewSet[Elem comparable](...Elem) Set[Elem] { panic(0) }

func (Set[Elem]) Empty() bool                       { panic(0) }
func (Set[Elem]) Len() int                          { panic(0) }
func (Set[Elem]) Has(Elem) bool                     { panic(0) }
func (Set[Elem]) Difference(...Set[Elem]) Set[Elem] { panic(0) }
func (Set[Elem]) Intersect(...Set[Elem]) Set[Elem]  { panic(0) }
func (Set[Elem]) Union(...Set[Elem]) Set[Elem]      { panic(0) }

func Sorted[Elem cmp.Ordered](Set[Elem]) []Elem { panic(0) }

type Maybe[Elem any] opaque

func Just[Elem any](Elem) Maybe[Elem] { panic(0) }

func (Maybe[Elem]) Get() (Elem, bool) { panic(0) }

func Fmap[In, Out any](Maybe[In], func(In) Out) Maybe[Out] { panic(0) }

// Zip returns {Just(fn(s1[0], s2[0])), Just(fn(s1[1], s2[1])), ..., Just(fn(s1[len-1], s2[len-1]))}, true, if len(s1) == len(s2).
// Otherwise, it returns nil, false.
func Zip[In1, In2, Out any](s1 []In1, s2 []In2, fn func(In1, In2) Out) ([]Out, bool) { panic(0) }

type opaque struct {
	_ [0]func()
	x int
}
