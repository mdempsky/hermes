// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lang

type omit any
type inherit any
type define any
type redefine any
type language any

// TODO(mdempsky): Do I want an explicit "entry" point?
//
// In scheme-to-c, it only changes once, from "Expr" to "Program".

// TODO(mdempsky): I'm adding some explicit "omits" that I think
// should be inferred automatically.

type Lsrc[
	Primitive define,
	Symbol define,

	Expr interface {
		*Primitive
		*Symbol
		*Const
		Quote(X Datum)
		IfThen(Cond, Then Expr)
		If(Cond, Then, Else Expr)
		Or(X []Expr)
		And(X []Expr)
		Not(X Expr)
		Begin(Init []Expr, Body Expr)
		Lambda(Params []Symbol, Init []Expr, Body Expr)
		Let(Bindings []Binding, Init []Expr, Body Expr)
		LetRec(Bindings []Binding, Init []Expr, Body Expr)
		Set(Var Symbol, Val Expr)
		Apply(Fun Expr, Args []Expr)
	},
	Binding struct {
		Var Symbol
		Val Expr
	},
	Const interface {
		True()
		False()
		Nil()
		Int(X int)
	},
	Datum interface {
		*Const
		Pair(Car, Cdr Datum)
		Vector(List []Datum)
	},
] language

// L1 removes one-armed if and adds the void primitive.
type L1[
	Primitive redefine, // adds Void
	Expr interface {
		*Primitive // keep embedding
		IfThen() omit
	},
] language

// L2 removes or, and, and not forms.
type L2[
	Expr interface {
		Or() omit
		And() omit
		Not() omit
	},
] language

// L3 removes multiple expressions from the body of lambda, let, and
// letrec (to be replaced with a single begin expression that contains
// the expressions from the body).
type L3[
	Symbol, Binding inherit,
	Expr interface {
		Lambda(Params []Symbol, Body Expr)
		Let(Bindings []Binding, Body Expr)
		LetRec(Bindings []Binding, Body Expr)
	},
] language

// L4 removes raw primitives (to be replaced with a lambda and a
// primitive call).
type L4[
	Primitive inherit,
	Expr interface {
		**Primitive // TODO(mdempsky): implicit?
		PrimCall(Prim Primitive, Args []Expr)
	},
] language

// L5 removes raw constants (to be replaced with quoted constant).
type L5[
	Const inherit,
	Expr interface {
		**Const
	},
	Datum interface {
		*Const // preserve
	},
] language

// L6 removes quoted datum (to be replaced with explicit calls to cons
// and make-vector+vector-set!).
type L6[
	Const inherit,
	Expr interface {
		Quote(X Const)
	},
] language

// L7 adds a listing of assigned variables to the body of the binding
// forms: let, letrec, and lambda.
type L7[
	Expr interface {
		Lambda(Params []Symbol, Body AssignedBody)
		Let(Bindings []Binding, Body AssignedBody)
		LetRec(Bindings []Binding, Body AssignedBody)
	},
	// TODO(mdempsky): This seems more fitting as an attribute grammar.
	AssignedBody struct {
		Names []Symbol
		Body  Expr
	},
	Symbol, Binding inherit,
] language

// L8 changes letrec bindings to only bind variables to lambdas.
type L8[
	Expr interface {
		*Symbol
		*LambdaExpr
		LetRec(Bindings []RecBinding, Body Expr)
	},
	RecBinding struct {
		Var Symbol
		Val LambdaExpr
	},
	// TODO(mdempsky): What would be the trade-offs of representing this
	// instead as "struct{ Params []Symbol; Body AssignedBody }"?
	//
	// I suppose if I did that, I wouldn't be allowed to embed it in
	// Expr, though I'm getting rid of that soon anyway.
	LambdaExpr interface {
		Lambda(Params []Symbol, Body AssignedBody)
	},
	Symbol, AssignedBody inherit,
] language

// L9 removes lambda expressions from expression context, effectively
// meaning we can only have lambdas bound in the right-hand-side of
// letrec expressions.
type L9[
	Expr interface {
		**LambdaExpr
	},
	LambdaExpr inherit,
] language

// L10 removes set! and assigned bodies (to be replaced by set-box!
// primcall for set!, and unbox primcall for references of assigned
// variables).
type L10[
	AssignedBody omit, // asserts that this form is gone
	Expr interface {
		Set() omit
		Let(Bindings []Binding, Body Expr)
	},
	LambdaExpr interface {
		Lambda(Params []Symbol, Body Expr)
	},
	Binding, Symbol inherit,
] language

// L11 add a list of free variables to the body of lambda expressions
// (starting closure conversion code).
type L11[
	Symbol, Expr inherit,
	LambdaExpr interface {
		Lambda(Params []Symbol, Body FreeBody)
	},
	FreeBody interface {
		Free(Free []Symbol, Body Expr)
	},
] language

// L12 removes the letrec form and adds closure and labels forms to
// replace it.  The closure form binds a variable to a label (code
// pointer) and its set of free variables, and the labels form binds
// labels (code pointer) to lambda expressions.
type L12[
	// TODO(mdempsky): AssignedBody should be gone here.
	AssignedBody omit,

	Symbol inherit,
	RecBinding inherit,

	Expr interface {
		LetRec() omit
		Label(Name Symbol)
		Closures(Closures []Closure, Body LabelsBody)
	},
	Closure struct {
		X Symbol
		L Symbol
		F []Symbol
	},
	LabelsBody interface {
		Labels(Bindings []RecBinding, Body Expr)
	},
] language

// L13 finishes closure conversion, removes the closures form,
// replacing it with primitive calls to deal with closure objects, and
// raises the labels from into the Expr non-terminal.
type L13[
	FreeBody omit,

	Primitive redefine, // adds Closure
	Symbol inherit,
	RecBinding inherit,

	Expr interface {
		Closures() omit
		Labels(Bindings []RecBinding, Body Expr)
	},
	LambdaExpr interface {
		Lambda(Params []Symbol, Body Expr)
	},
] language

// L14 removes labels form from the Expr nonterminal and puts a single
// labels form at the top.  Essentially this raises all of our closure
// converted functions to the top.
type L14[
	RecBinding, Symbol inherit,
	Program interface {
		Labels(Bindings []RecBinding, Entry Symbol)
	},
] language

// L15 moves simple expressions (constants and variable references)
// out of the Expr nonterminal, and replaces expressions referred to
// in calls and primcalls with simple expressions.  This effectively
// removes complex operands to calls and primcalls.
type L15[
	Const, Primitive, Symbol inherit,

	Expr interface {
		**Symbol // TODO(mdempsky): should be implicit
		*SimpleExpr
		PrimCall(Prim Primitive, Args []SimpleExpr)
		Apply(Fun SimpleExpr, Args []SimpleExpr)
	},
	SimpleExpr interface {
		*Symbol
		Label(Name Symbol)
		Quote(X Const)
	},
] language

// L16 separates the Expr nonterminal into the Value, Effect, and
// Predicate nonterminals.  This is needed to translate from our
// expression language into a language like C that has statements
// (effects) and expressions (values) and predicates that need to be
// simply values.
type L16[
	Expr omit, // should be absent anyway; asserts this
	Const, Symbol inherit,
	ValuePrim, EffectPrim, PredicatePrim define,

	SimpleExpr interface {
		*Symbol
		Label(Name Symbol)
		Quote(X Const)
	},
	Binding struct {
		Var Symbol
		Val Value
	},
	Value interface {
		*SimpleExpr
		IfValue(Cond Predicate, Then, Else Value)
		BeginValue(Init []Effect, X Value)
		LetValue(Bindings []Binding, Body Value)
		PrimValue(Prim ValuePrim, Args []SimpleExpr)
		ApplyValue(Fun SimpleExpr, Args []SimpleExpr)
	},
	Effect interface {
		Nop()
		IfEffect(Cond Predicate, Then, Else Effect)
		BeginEffect(Init []Effect, X Effect) // TODO(mdempsky): Why non-empty? For consistency I guess?
		LetEffect(Bindings []Binding, Body Effect)
		PrimEffect(Prim EffectPrim, Args []SimpleExpr)
		ApplyEffect(Fun SimpleExpr, Args []SimpleExpr)
	},
	Predicate interface {
		True()
		False()
		IfPred(Cond Predicate, Then, Else Predicate)
		BeginPred(Init []Effect, X Predicate)
		LetPred(Bindings []Binding, Body Predicate)
		PrimPred(Prim PredicatePrim, Args []SimpleExpr)
	},
	LambdaExpr interface {
		Lambda(Params []Symbol, Body Value)
	},
] language

// L17 removes the allocation primitives: cons, box, make-vector, and
// make-closure and adds a generic alloc form for specifying
// allocation.  It also adds raw integers for specifying type tags in
// the alloc form.
type L17[
	SimpleExpr inherit,
	ValuePrim, EffectPrim redefine,
	Value interface {
		Alloc(Tag int64, Size SimpleExpr)
	},
] language

// L18 removes let forms and replaces them with a top-level locals
// form that indicates which variables are bound in the function (so
// they can be listed at the top of our C function) and set! that do
// simple assignments.
type L18[
	Symbol inherit,
	Value interface {
		LetValue() omit
	},
	Effect interface {
		LetEffect() omit
		Set(Var Symbol, Val Value)
	},
	Predicate interface {
		LetPred() omit
	},
	LambdaExpr interface {
		Lambda(Params, Locals []Symbol, Body Value)
	},
] language

// L19 simplifies the right-hand-side of a set! so that it can
// contain, simple expression, allocations, primcalls, and function
// calls, but not ifs, or begins.
type L19[
	Symbol, ValuePrim inherit,
	SimpleExpr inherit,
	Value interface {
		**SimpleExpr // TODO(mdempsky): again, should be implied
		*Rhs
	},
	Rhs interface {
		*SimpleExpr
		Alloc(Tag int64, Size SimpleExpr)
		PrimValue(Prim ValuePrim, Args []SimpleExpr)
		ApplyValue(Fun SimpleExpr, Args []SimpleExpr)
	},
	Effect interface {
		Set(Lhs Symbol, Rhs Rhs)
	},
] language

// L20 removes begin from the predicate production (effectively
// forcing the if to only have if, true, false, and predicate
// primitive calls).
//
// TODO: removed this language because our push-if pass was buggy, and

// L21 removes quoted constants and replace it with our raw ptr
// representation (i.e. 64-bit integers)
type L21[
	// TODO(mdempsky): Const here is now unreachable.
	Datum omit,
	SimpleExpr interface {
		Quote() omit
		Int(Int int64)
	},
] language

// L22 removes the primcalls and replace them with mref (memory
// references), add, subtract, multiply, divide, shift-right,
// shift-left, logand, mset! (memory set), =, <, and <=.
//
// TODO: we should probably replace this with "machine" instructions
// instead, so that we can more easily extend the language and
// generate C code from it.
type L22[
	Rhs interface {
		PrimValue() omit
	},
	SimpleExpr interface {
		MRef(Ptr SimpleExpr, Index *SimpleExpr, Offset int64)
		Add(X, Y SimpleExpr)
		Subtract(X, Y SimpleExpr)
		Multiple(X, Y SimpleExpr)
		Divide(X, Y SimpleExpr)
		ShiftRight(X, Y SimpleExpr)
		ShiftLeft(X, Y SimpleExpr)
		LogicalAnd(X, Y SimpleExpr)
	},
	Effect interface {
		PrimEffect() omit
		MSet(Ptr SimpleExpr, Index *SimpleExpr, Offset int64, Data SimpleExpr)
	},
	Predicate interface {
		PrimPred() omit
		Eql(X, Y SimpleExpr)
		Lss(X, Y SimpleExpr)
		Leq(X, Y SimpleExpr)
	},
] language
