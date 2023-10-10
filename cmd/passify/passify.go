// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmp"
	"fmt"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	files, err := filepath.Glob("passes/*.go")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		cfg := packages.Config{
			Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedImports | packages.NeedDeps,
		}
		pkgs, err := packages.Load(&cfg, "file="+file)
		if err != nil {
			log.Fatal(err)
		}
		do(filepath.Base(file), pkgs)
	}
}

func do(file string, pkgs []*packages.Package) {
	prog, ssaPkgs := ssautil.Packages(pkgs, 0)
	prog.Build()

	if len(pkgs) != 1 || len(ssaPkgs) != 1 {
		panic("weird")
	}
	pkg := ssaPkgs[0]
	sizes = pkgs[0].TypesSizes

	var Lsrc, Ldst *types.Package

	{
		entry := pkg.Func("Entry")
		sig := entry.Type().(*types.Signature)
		if sig.Params().Len() != 1 || sig.Results().Len() != 1 {
			log.Fatal("weird Entry signature:", sig)
		}

		// The source and destination entry types.
		src := sig.Params().At(0).Type().(*types.Named)
		dst := sig.Results().At(0).Type().(*types.Named)

		// Their corresponding packages.
		Lsrc = src.Obj().Pkg()
		Ldst = dst.Obj().Pkg()
		fmt.Printf("\n### %v: %v -> %v\n", file, Lsrc.Name(), Ldst.Name())
	}

	for _, name := range keys(pkg.Members) {
		fn, ok := pkg.Members[name].(*ssa.Function)
		if !ok || strings.HasPrefix(name, "init") || name == "Entry" || name[0] >= 'a' && name[0] <= 'z' {
			continue
		}
		fmt.Println(fn)

		// Okay, I want to figure out how to morph an Lsrc.IfThen into an L1.Expr.
		//
		// An Lsrc.IfThen has two Lsrc.Expr fields, and pass.IfThen takes two L1.Expr params.
		// So we recursively need a morphism from Lsrc.Expr to L1.Expr too.
		//
		// With that morphism, we can define the input parameters as morph[Lsrc.Expr->L1.Expr](Lsrc.IfThen.Cond) and morph[Lsrc.Expr->L1.Expr](Lsrc.IfThen.Then).
		//
		// We interpret the IfThen code, and see it constructs an L1.If value (of type L1.Expr).
		// Also, it constructs all of the fields, so we're happy.

		p := &ParamExpr{Name: "e::Lsrc.IfThen"}

		srcType := Lsrc.Scope().Lookup(name).(*types.TypeName)
		if srcUnder, ok := srcType.Type().(*types.Named).Underlying().(*types.Struct); ok {

			sig := fn.Type().(*types.Signature)
			if sig.Params().Len() != srcUnder.NumFields() {
				log.Fatalf("bad signature: cannot morph %v fields into %v parameters", srcUnder.NumFields(), sig.Params().Len())
			}

			env := make(map[ssa.Value]Expr)
			for i, param := range fn.Params {
				field := srcUnder.Field(i)

				env[param] = &MorphExpr{
					X: &ProjExpr{
						X:     p,
						Index: i,
						Field: field,
					},
					Src: field.Type(),
					Dst: param.Type(),
				}
			}
		}

		var h heap

		// Allocate heap memory.
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.Alloc:
					h.alloc(instr)
				}
			}
		}

		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.MapUpdate, *ssa.Go, *ssa.Defer, *ssa.Send, *ssa.MakeChan, *ssa.MakeMap, *ssa.RunDefers, *ssa.Select:
					log.Fatal("not supported", instr)
				case *ssa.Store:
					h.store(instr)
				}
			}
		}

		indent := "\t"

		var printExpr func(v ssa.Value)
		printExpr = func(v ssa.Value) {
			switch v := v.(type) {
			case *ssa.Function:
				fmt.Print(v.Name())
			case *ssa.Const:
				fmt.Print(v.Value)
			case *ssa.Call:
				printExpr(v.Call.Value)
				fmt.Printf("(")
				for i, arg := range v.Call.Args {
					if i > 0 {
						fmt.Printf(", ")
					}
					printExpr(arg)
				}
				fmt.Printf(")")
			case *ssa.Extract:
				if v.Index == 1 {
					if assert, ok := v.Tuple.(*ssa.TypeAssert); ok {
						fmt.Printf("is[%v]?(", assert.AssertedType)
						printExpr(assert.X)
						fmt.Printf(")")
						return
					}
				}

				printExpr(v.Tuple)
				fmt.Printf("@%v", v.Index)
			case *ssa.MakeInterface:
				printExpr(v.X)
			case *ssa.Parameter:
				fmt.Printf("PARAM.%v", v.Name())
			case *ssa.Slice:
				if x, ok := v.X.(*ssa.Alloc); ok {
					if c, ok := h.allocs[x]; ok {
						fmt.Printf("%v{", v.Type())
						for i, offset := range keys(c.elems) {
							if i > 0 {
								fmt.Printf(", ")
							}
							printExpr(c.elems[offset])
						}
						fmt.Printf("}")
						return
					}
				}
			case *ssa.FieldAddr:
				printExpr(v.X)
				fmt.Printf(".%v", v.X.Type().(*types.Pointer).Elem().Underlying().(*types.Struct).Field(v.Field).Name())
			case *ssa.UnOp:
				if v.Op == token.MUL {
					if x, ok := v.X.(*ssa.Alloc); ok {
						if c, ok := h.allocs[x]; ok {
							fmt.Printf("%v{", types.TypeString(c.typ, types.RelativeTo(pkgs[0].Types)))
							for i, offset := range keys(c.elems) {
								if i > 0 {
									fmt.Printf(", ")
								}
								fmt.Printf("%v: ", offset)
								printExpr(c.elems[offset])
							}
							fmt.Printf("}")
							return
						}
						fmt.Printf("unknown alloc: %v\n", x)
					}
					if v, ok := h.load(v.X); ok {
						printExpr(v)
						return
					}
				}
				fmt.Printf("%v ", v.Op)
				printExpr(v.X)
			default:
				fmt.Printf("[[%T]]", v)
			}
		}

		var walk func(b *ssa.BasicBlock)
		walk = func(b *ssa.BasicBlock) {
			switch last := b.Instrs[len(b.Instrs)-1].(type) {
			default:
				log.Fatalf("unexpected instruction: %v (%T)", last, last)
			case *ssa.If:
				old := indent
				fmt.Print(indent + "if (")
				printExpr(last.Cond)
				fmt.Print(") {\n")
				indent += "\t"
				walk(b.Succs[0])
				fmt.Print(old + "} else {\n")
				walk(b.Succs[1])
				fmt.Println(old + "}\n")
				indent = old
			case *ssa.Jump:
				walk(b.Succs[0])
			case *ssa.Panic:
				fmt.Print(indent + "panic(")
				printExpr(last.X)
				fmt.Print(")\n")
			case *ssa.Return:
				fmt.Print(indent + "return (")
				if isZero(last.Results[0]) {
					fmt.Print("ZERO")
				} else {
					printExpr(last.Results[0])
				}
				fmt.Print(")\n")
			}
		}
		fmt.Println("{{{")
		walk(fn.Blocks[0])
		fmt.Println("}}}")

		_ = p
	}
}

func isZero(v ssa.Value) bool {
	switch v := v.(type) {
	case *ssa.Const:
		if v.Value == nil {
			return true
		}
	}
	// TODO(mdempsky): Recognize more zero values.
	return false
}

type Expr interface {
	isExpr()
}

type ParamExpr struct {
	Name string
}

type ProjExpr struct {
	X     Expr
	Index int
	Field *types.Var
}

type MorphExpr struct {
	X Expr

	// What homomorphism we're applying here.
	Src, Dst types.Type
}

type LitExpr struct {
	Type  types.Type // should have underlying type Struct
	Elems []Expr
}

type FieldAddrExpr struct {
	Ptr   *LitExpr
	Index int
}

func (*ParamExpr) isExpr()     {}
func (*ProjExpr) isExpr()      {}
func (*MorphExpr) isExpr()     {}
func (*LitExpr) isExpr()       {}
func (*FieldAddrExpr) isExpr() {}

func keys[K cmp.Ordered, V any](m map[K]V) []K {
	res := make([]K, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}
