// Code generated by Hermes. DO NOT EDIT.

package L5

type terminal int

type (
	Binding struct {
		Var Symbol
		Val Expr
	}
	Primitive terminal // from L1
	Symbol    terminal // from Lsrc
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
	Lambda struct {
		Params []Symbol
		Body   Expr
	}
	Let struct {
		Bindings []Binding
		Body     Expr
	}
	LetRec struct {
		Bindings []Binding
		Body     Expr
	}
	PrimCall struct {
		Prim Primitive
		Args []Expr
	}
	Quote struct{ X Datum }
	Set   struct {
		Var Symbol
		Val Expr
	}
)

func (False) isConst()   {}
func (False) isDatum()   {}
func (Int) isConst()     {}
func (Int) isDatum()     {}
func (Nil) isConst()     {}
func (Nil) isDatum()     {}
func (True) isConst()    {}
func (True) isDatum()    {}
func (Pair) isDatum()    {}
func (Vector) isDatum()  {}
func (Apply) isExpr()    {}
func (Begin) isExpr()    {}
func (If) isExpr()       {}
func (Lambda) isExpr()   {}
func (Let) isExpr()      {}
func (LetRec) isExpr()   {}
func (PrimCall) isExpr() {}
func (Quote) isExpr()    {}
func (Set) isExpr()      {}
func (Symbol) isExpr()   {}
