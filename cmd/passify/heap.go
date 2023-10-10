// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/types"
	"log"

	"golang.org/x/tools/go/ssa"
)

var sizes types.Sizes

type heap struct {
	allocs map[*ssa.Alloc]cell
}

type cell struct {
	typ   types.Type
	elems map[int64]ssa.Value
}

func (c *cell) init(base int64, typ types.Type) {
	switch typ := typ.Underlying().(type) {
	case *types.Array:
		elem := typ.Elem()
		elemSize := sizes.Sizeof(elem)
		if elemSize == 0 || typ.Len() == 0 {
			return
		}
		for i := int64(0); i < typ.Len(); i++ {
			c.init(base+i*elemSize, elem)
		}
	case *types.Struct:
		for i, offset := range offsets(typ) {
			c.init(base+offset, typ.Field(i).Type())
		}
	default:
		c.elems[base] = nil
	}
}

func offsets(typ *types.Struct) []int64 {
	fields := make([]*types.Var, typ.NumFields())
	for i := 0; i < typ.NumFields(); i++ {
		fields[i] = typ.Field(i)
	}
	return sizes.Offsetsof(fields)
}

func (h *heap) alloc(v *ssa.Alloc) {
	typ := v.Type().(*types.Pointer).Elem()
	c := cell{
		typ:   typ,
		elems: make(map[int64]ssa.Value),
	}
	c.init(0, typ)
	fmt.Printf("offsets: %v => %v\n", v.Type(), keys(c.elems))
	if h.allocs == nil {
		h.allocs = make(map[*ssa.Alloc]cell)
	}
	h.allocs[v] = c
}

func (h *heap) store(v *ssa.Store) {
	alloc, offset := evalPtr(v.Addr, 0)

	cell, ok := h.allocs[alloc]
	if !ok {
		log.Fatalf("no allocation for %v", alloc)
	}
	old, ok := cell.elems[offset]
	if !ok {
		log.Fatalf("offset %v not valid within %v (%v); %v", offset, alloc, alloc.Type(), keys(cell.elems))
	}
	if old != nil {
		log.Fatalf("offset %v within %v already initialized to %v", offset, alloc, old)
	}

	cell.elems[offset] = v.Val
}

func (h *heap) load(v ssa.Value) (ssa.Value, bool) {
	alloc, offset := evalPtr(v, 0)

	cell, ok := h.allocs[alloc]
	if !ok {
		log.Fatalf("no allocation for %v", alloc)
	}
	old, ok := cell.elems[offset]
	return old, ok && old != nil
}

func evalPtr(v ssa.Value, base int64) (*ssa.Alloc, int64) {
	switch v := v.(type) {
	case *ssa.Alloc:
		return v, base
	case *ssa.FieldAddr:
		typ := v.X.Type().(*types.Pointer).Elem().Underlying().(*types.Struct)
		return evalPtr(v.X, base+offsets(typ)[v.Field])
	case *ssa.IndexAddr:
		typ := v.X.Type().(*types.Pointer).Elem().Underlying().(*types.Array)
		return evalPtr(v.X, base+sizes.Sizeof(typ.Elem())*evalInt(v.Index))
	default:
		log.Fatalf("unexpected pointer: %v (%T)", v, v)
		panic(0)
	}
}

func evalInt(v ssa.Value) int64 {
	switch v := v.(type) {
	case *ssa.Const:
		return v.Int64()
	}
	log.Fatalf("unexpected int: %v (%T)", v, v)
	panic(0)
}

/*
	evalAddr := func(addr ssa.Value) *ssa.Value {
		orig := addr

		var i, n int64 = 0, 1
		switch addr0 := addr.(type) {
		case *ssa.FieldAddr:
			i = int64(addr0.Field)
			addr = addr0.X
		case *ssa.IndexAddr:
			i = addr0.Index.(*ssa.Const).Int64()
			addr = addr0.X
		}
		switch addr := addr.(type) {
		case *ssa.Alloc:
			if cell, ok := heap[addr]; ok {
				if i >= int64(len(cell)) {
					log.Fatalf("index %v for alloc %v\n", i, addr)
				}
				return &cell[i]
			}
			panic("missing cell for Alloc")
		}
		_ = n
		log.Fatalf("unknown instruction: %v (%T); %v", addr, addr, orig)
		panic(0)
	}


*/
