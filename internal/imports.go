package internal

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal/consts"
	"github.com/flyingdice/whack-sdk/sdk/runtime/config"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// importObject builds a wasmer.ImportObject that exposes host imports
// (functions, globals) to WASM code.
func importObject(
	env *wasmer.WasiEnvironment,
	store *wasmer.Store,
	mod *wasmer.Module,
	imports config.HostImports,
) (*wasmer.ImportObject, error) {
	// Create new global imports object.
	obj, err := env.GenerateImportObject(store, mod)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create import object")
	}

	// Extend import object with our host imports.
	ext := hostImports(store, imports)
	obj.Register(consts.Namespace, ext)

	return obj, nil
}

// hostImports builds wasmer.IntoExtern object that can be registered
// with the import object to be accessible by WASM code.
func hostImports(
	store *wasmer.Store,
	imports config.HostImports,
) map[string]wasmer.IntoExtern {
	ext := map[string]wasmer.IntoExtern{}

	for name, fn := range importFunctions(imports) {
		ext[name] = fn.Bind(store)
	}

	for name, gbl := range importGlobals(imports) {
		ext[name] = gbl.Bind(store)
	}

	return ext
}

// importFunctions builds a map of host functions that can be accessed by WASM code.
func importFunctions(imports config.HostImports) map[string]*function {
	functions := make(map[string]*function)
	for _, fn := range imports.Functions() {
		functions[fn.Name()] = NewFunction(fn)
	}
	return functions
}

// importGlobals builds a map of globals that can be accessed by WASM code.
func importGlobals(imports config.HostImports) map[string]*global {
	globals := make(map[string]*global)
	return globals
}
