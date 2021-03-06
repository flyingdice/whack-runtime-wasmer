package internal

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal/consts"
	"github.com/flyingdice/whack-sdk/pkg/sdk"
	"github.com/flyingdice/whack-sdk/pkg/sdk/runtime"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// importObject builds a wasmer.ImportObject that exposes host imports
// (functions, globals) to WASM code.
func importObject(
	runtime *Runtime,
	instanceWrn sdk.WRN,
	imports runtime.HostImports,
) (*wasmer.ImportObject, error) {
	// Create new global imports object.
	obj, err := runtime.env.GenerateImportObject(runtime.store, runtime.module)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create import object")
	}

	// Extend import object with our host imports.
	ext := hostImports(runtime, instanceWrn, imports)
	obj.Register(consts.Namespace, ext)

	return obj, nil
}

// hostImports builds wasmer.IntoExtern object that can be registered
// with the import object to be accessible by WASM code.
func hostImports(
	runtime *Runtime,
	instanceWrn sdk.WRN,
	imports runtime.HostImports,
) map[string]wasmer.IntoExtern {
	ext := map[string]wasmer.IntoExtern{}

	for name, fn := range importFunctions(runtime, instanceWrn, imports) {
		ext[name] = fn.Bind(runtime.store)
	}

	for name, gbl := range importGlobals(runtime, instanceWrn, imports) {
		ext[name] = gbl.Bind(runtime.store)
	}

	return ext
}

// importFunctions builds a map of host functions that can be accessed by WASM code.
func importFunctions(runtime *Runtime, instanceWrn sdk.WRN, imports runtime.HostImports) map[string]*function {
	functions := make(map[string]*function)
	for _, fn := range imports.Functions() {
		functions[fn.Name()] = NewFunction(runtime, instanceWrn, fn)
	}
	return functions
}

// importGlobals builds a map of globals that can be accessed by WASM code.
func importGlobals(runtime *Runtime, instanceWrn sdk.WRN, imports runtime.HostImports) map[string]*global {
	globals := make(map[string]*global)
	return globals
}
