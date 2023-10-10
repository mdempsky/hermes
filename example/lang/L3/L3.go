// Code generated by Hermes. DO NOT EDIT.

package L3

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
		Expr
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
	Quote struct{ X Datum }
	Set   struct {
		Var Symbol
		Val Expr
	}
)

func (False) isConst()    {}
func (False) isDatum()    {}
func (False) isExpr()     {}
func (Int) isConst()      {}
func (Int) isDatum()      {}
func (Int) isExpr()       {}
func (Nil) isConst()      {}
func (Nil) isDatum()      {}
func (Nil) isExpr()       {}
func (True) isConst()     {}
func (True) isDatum()     {}
func (True) isExpr()      {}
func (Pair) isDatum()     {}
func (Vector) isDatum()   {}
func (Apply) isExpr()     {}
func (Begin) isExpr()     {}
func (If) isExpr()        {}
func (Lambda) isExpr()    {}
func (Let) isExpr()       {}
func (LetRec) isExpr()    {}
func (Quote) isExpr()     {}
func (Set) isExpr()       {}
func (Primitive) isExpr() {}
func (Symbol) isExpr()    {}
