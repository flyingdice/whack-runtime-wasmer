package internal

import "github.com/wasmerio/wasmer-go/wasmer"

type global struct {
	value interface{}
}

func (g *global) Bind(store *wasmer.Store) *wasmer.Global {
	return nil
	//gtype := pkg.NewGlobalType(pkg.I32, pkg.IMMUTABLE)
	//value := pkg.NewValue(g.value, pkg.I32)
	//return pkg.NewGlobal(store, gtype, value)
}

func NewGlobal(value interface{}) (*global, error) {
	return &global{
		value: value,
	}, nil
}
