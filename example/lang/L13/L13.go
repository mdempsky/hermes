// Code generated by Hermes. DO NOT EDIT.

package L13

type terminal int

type (
	Binding struct {
		Var Symbol
		Val Expr
	}
	Closure struct {
		X Symbol
		L Symbol
		F []Symbol
	}
	Primitive  terminal // from L13
	RecBinding struct {
		Var Symbol
		Val LambdaExpr
	}
	Symbol terminal // from Lsrc
)

type (
	Const interface {
		Datum
		isConst()
	}
	False struct{}
	Int   struct{ X int }
	Nil   struct{}
	True  struct{}
)

type (
	Datum  interface{ isDatum() }
	Pair   struct{ Car, Cdr Datum }
	Vector struct{ List []Datum }
)

type (
	Expr  interface{ isExpr() }
	Apply struct {
		Fun  Expr
		Args []Expr
	}
	Begin struct {
		Init []Expr
		Body Expr
	}
	If     struct{ Cond, Then, Else Expr }
	Label  struct{ Name Symbol }
	Labels struct {
		Bindings []RecBinding
		Body     Expr
	}
	Let struct {
		Bindings []Binding
		Body     Expr
	}
	PrimCall struct {
		Prim Primitive
		Args []Expr
	}
	Quote struct{ X Const }
)

type (
	LabelsBody interface{ isLabelsBody() }
)

type (
	LambdaExpr interface{ isLambdaExpr() }
	Lambda     struct {
		Params []Symbol
		Body   Expr
	}
)

func (False) isConst()       {}
func (False) isDatum()       {}
func (Int) isConst()         {}
func (Int) isDatum()         {}
func (Nil) isConst()         {}
func (Nil) isDatum()         {}
func (True) isConst()        {}
func (True) isDatum()        {}
func (Pair) isDatum()        {}
func (Vector) isDatum()      {}
func (Apply) isExpr()        {}
func (Begin) isExpr()        {}
func (If) isExpr()           {}
func (Label) isExpr()        {}
func (Labels) isExpr()       {}
func (Let) isExpr()          {}
func (PrimCall) isExpr()     {}
func (Quote) isExpr()        {}
func (Lambda) isLambdaExpr() {}
func (Symbol) isExpr()       {}
