// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmp"
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := packages.Config{
		Mode: packages.NeedTypes,
	}
	pkgs, err := packages.Load(&cfg, ".")
	if err != nil {
		log.Fatal(err)
	}
	pkg := pkgs[0].Types
	scope := pkg.Scope()

	var langs []*types.Named
	for _, name := range scope.Names() {
		if !strings.HasPrefix(name, "L") {
			continue
		}
		lang := scope.Lookup(name).(*types.TypeName).Type().(*types.Named)
		langs = append(langs, lang)
	}

	slices.SortFunc(langs, func(li, lj *types.Named) int {
		return cmp.Compare(li.Obj().Pos(), lj.Obj().Pos())
	})

	base := "lang"
	os.MkdirAll(base, 0777)

	var L lang
	for _, l := range langs {
		L = L.extend(l.Obj().Name(), l.TypeParams())

		dir := filepath.Join(base, L.name)
		os.Mkdir(dir, 0777)

		buf := []byte(L.String())
		if buf1, err := format.Source(buf); err != nil {
			fmt.Printf("%v: format error: %v\n", L.name, err)
		} else {
			buf = buf1
		}
		os.WriteFile(filepath.Join(dir, L.name+".go"), buf, 0666)
	}
}

type lang struct {
	name string
	defs map[string]Define
}

type Define interface {
	copy() Define
}

type term struct {
	pass   string
	isAlso map[string]bool
}

func (t term) copy() Define { t.isAlso = nil; return &t }

type nonterm struct {
	embeds map[string]bool
	cons   map[string]*types.Signature
	str    *types.Struct
	isAlso map[string]bool
}

func (t nonterm) copy() Define { dup(&t.embeds); dup(&t.cons); t.isAlso = nil; return &t }

func dup[K comparable, V any](mp *map[K]V) {
	m := make(map[K]V, len(*mp))
	for k, v := range *mp {
		m[k] = v
	}
	*mp = m
}

func (L0 lang) extend(langName string, tparams *types.TypeParamList) (L lang) {
	L.name = langName

	L.defs = make(map[string]Define, len(L0.defs))
	for defName, def := range L0.defs {
		L.defs[defName] = def.copy()
	}

	commands := make(map[string][]string)
	var delta []*types.TypeParam

	for i := 0; i < tparams.Len(); i++ {
		tparam := tparams.At(i)
		name := tparam.Obj().Name()
		switch typ := tparam.Constraint().(type) {
		case *types.Named:
			command := typ.Obj().Name()
			commands[command] = append(commands[command], name)

		case *types.Interface:
			delta = append(delta, tparam)

		default:
			panic("unknown")
		}
	}

	take := func(key string) []string {
		res := commands[key]
		delete(commands, key)
		return res
	}

	for _, defName := range take("inherit") {
		if L.defs[defName] == nil {
			fmt.Printf("cannot inherit undefined %v\n", defName)
		}
	}

	for _, defName := range take("define") {
		if L.defs[defName] == nil {
			L.defs[defName] = &term{pass: L.name}
		} else {
			fmt.Printf("%v is already defined\n", defName)
		}
	}

	for _, defName := range take("redefine") {
		if _, ok := L.defs[defName].(*term); ok {
			L.defs[defName] = &term{pass: L.name}
		} else {
			fmt.Printf("%v is not a defined non-terminal\n", defName)
		}
	}

	var embedClobbers, consClobbers []string
	for _, typ := range delta {
		iface := typ.Constraint().(*types.Interface)
		if iface.IsImplicit() {
			// Product type.
			continue
		}

		for i := 0; i < iface.NumEmbeddeds(); i++ {
			embedType := iface.EmbeddedType(i)
			ptrs := 0
			for {
				ptr, ok := embedType.(*types.Pointer)
				if !ok {
					break
				}
				ptrs++
				embedType = ptr.Elem()
			}
			embed := embedType.(*types.TypeParam)
			embedClobbers = append(embedClobbers, embed.Obj().Name())
		}

		for i := 0; i < iface.NumMethods(); i++ {
			method := iface.Method(i)
			consClobbers = append(consClobbers, method.Name())
		}
	}

	for _, def := range L.defs {
		switch def := def.(type) {
		case *nonterm:
			for _, name := range embedClobbers {
				delete(def.embeds, name)
			}
			for _, name := range consClobbers {
				delete(def.cons, name)
			}
		}
	}

	for _, typ := range delta {
		defName := typ.Obj().Name()

		var nt *nonterm
		switch def := L.defs[defName].(type) {
		case nil:
			nt = &nonterm{
				embeds: make(map[string]bool),
				cons:   make(map[string]*types.Signature),
			}
			L.defs[defName] = nt
		case *nonterm:
			nt = def
		case *term:
			fmt.Printf("%v is already a terminal\n", defName)
		}

		iface := typ.Constraint().(*types.Interface)
		if iface.IsImplicit() {
			if nt.str != nil {
				// TODO(mdempsky): This is a redefinition. Should this require special syntax?
				// fmt.Printf("%v is already a product type\n", name)
			}
			nt.str = iface.EmbeddedType(0).(*types.Struct)
			continue
		}

		for i := 0; i < iface.NumEmbeddeds(); i++ {
			typ := iface.EmbeddedType(i)
			ptrs := 0
			for {
				ptr, ok := typ.(*types.Pointer)
				if !ok {
					break
				}
				typ = ptr.Elem()
				ptrs++
			}
			tparam := typ.(*types.TypeParam)
			embedName := tparam.Obj().Name()
			if ptrs == 1 {
				nt.embeds[embedName] = true
			}
		}

		for i := 0; i < iface.NumMethods(); i++ {
			con := iface.Method(i)
			conName := con.Name()
			sig := con.Type().(*types.Signature)

			if res := sig.Results(); res.Len() != 0 {
				if res.Len() != 1 || res.At(0).Type().(*types.Named).Obj().Name() != "omit" {
					fmt.Printf("unexpected signature result: %v\n", res)
				}
			} else {
				nt.cons[conName] = sig
			}
		}
	}

	for _, name := range take("omit") {
		delete(L.defs, name)
	}

	var flood func(nt *nonterm, also string)
	flood = func(nt *nonterm, also string) {
		if nt.isAlso[also] {
			return
		}
		if nt.isAlso == nil {
			nt.isAlso = make(map[string]bool)
		}
		nt.isAlso[also] = true

		for embedName := range nt.embeds {
			switch def := L.defs[embedName].(type) {
			case *term:
				if def.isAlso == nil {
					def.isAlso = make(map[string]bool)
				}
				def.isAlso[also] = true
			case *nonterm:
				flood(def, also)
			}
		}
	}
	for defName, def := range L.defs {
		switch def := def.(type) {
		case *nonterm:
			flood(def, defName)
		}
	}

	if len(commands) != 0 {
		fmt.Printf("unknown commands: %v\n", commands)
	}

	return
}

func keys[K cmp.Ordered, V any](m map[K]V) []K {
	res := make([]K, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}

func (L lang) String() string {
	var head, body, foot strings.Builder

	fmt.Fprintf(&head, "// Code generated by Hermes. DO NOT EDIT.\n\n")

	fmt.Fprintf(&head, "package %v\n", L.name)

	fmt.Fprintf(&head, "type terminal int\n\n")
	fmt.Fprintf(&head, "type (\n")

	is := func(typs []string, cons ...string) {
		for _, con := range cons {
			for _, typ := range typs {
				fmt.Fprintf(&foot, "\nfunc (%v) is%v() {}", con, typ)
			}
		}
	}

	for _, defName := range keys(L.defs) {
		switch def := L.defs[defName].(type) {
		default:
			panic("unknown def")
		case *term:
			fmt.Fprintf(&head, "\n\t%v terminal // from %v", defName, def.pass)
			is(keys(def.isAlso), defName)
		case *nonterm:
			if def.str != nil {
				if len(def.cons) != 0 && len(def.embeds) != 0 {
					fmt.Printf("%v is both a set and product type\n", defName)
				}
				fmt.Fprintf(&head, "\n\t%v %v", defName, def.str)
				continue
			}

			fmt.Fprintf(&body, "\n\ntype (\n\t%v interface{", defName)
			for _, key := range keys(def.isAlso) {
				if key != defName {
					fmt.Fprintf(&body, "%v; ", key)
				}
			}
			fmt.Fprintf(&body, "is%v()", defName)
			fmt.Fprintf(&body, "}")

			for _, conName := range keys(def.cons) {
				con := def.cons[conName]
				if res := con.Results(); res.Len() != 0 {
					if res.Len() != 1 || res.At(0).Type().(*types.Named).Obj().Name() != "omit" {
						fmt.Printf("expected omit: %v\n", res)
					}
					continue
				}
				fmt.Fprintf(&body, "\n\t%v struct{", conName)
				var prev types.Type
				for i := 0; i < con.Params().Len(); i++ {
					param := con.Params().At(i)
					if prev != nil {
						if !types.Identical(prev, param.Type()) {
							fmt.Fprintf(&body, " %v;", prev)
						} else {
							fmt.Fprintf(&body, ",")
						}
					}
					fmt.Fprintf(&body, " %s", param.Name())
					prev = param.Type()
				}
				if prev != nil {
					fmt.Fprintf(&body, " %v ", prev)
				}
				fmt.Fprintf(&body, "}")
			}
			fmt.Fprintf(&body, "\n)")

			is(keys(def.isAlso), keys(def.cons)...)
		}
	}

	head.WriteString("\n)")

	if body.Len() != 0 {
		head.WriteString("\n\n")
		head.WriteString(body.String())
	}

	if foot.Len() != 0 {
		head.WriteString("\n\n")
		head.WriteString(foot.String())
	}

	return head.String()
}

type keyword string

const (
	omit     keyword = "omit"
	inherit  keyword = "inherit"
	define   keyword = "define"
	redefine keyword = "redefine"
	language keyword = "language"
)
