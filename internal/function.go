package internal

import (
	"github.com/flyingdice/whack-sdk/sdk"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// ValueFunc represents a function that can be used with.
type ValueFunc func([]wasmer.Value) ([]wasmer.Value, error)

// HostFunc translates between ValueFunc and ffi.Func.
type HostFunc func(...wasmer.Value) (interface{}, error)

// function represents a Wasmer function created from a host function.
type function struct {
	name   string
	args   []wasmer.ValueKind
	retval []wasmer.ValueKind
	fn     ValueFunc
}

func (f *function) Name() string { return f.name }

// Bind creates a new function in the given store.
func (f *function) Bind(store *wasmer.Store) *wasmer.Function {
	args := wasmer.NewValueTypes(f.args...)
	retval := wasmer.NewValueTypes(f.retval...)
	fnType := wasmer.NewFunctionType(args, retval)
	return wasmer.NewFunction(store, fnType, f.fn)
}

// NewFunction creates a function for the given whack SDK function.
func NewFunction(fn sdk.Function) *function {
	// Variable number of int32 arguments.
	args := make([]wasmer.ValueKind, fn.NumIn())
	for i := 0; i < fn.NumIn(); i++ {
		args[i] = wasmer.I32
	}

	// Variable number (0, 1) of int32 return values.
	retval := make([]wasmer.ValueKind, fn.NumOut())
	for i := 0; i < fn.NumOut(); i++ {
		retval[i] = wasmer.I32
	}

	// Wasmer function with translations to be callable as a pkg import.
	vf := valueFunc(hostFunc(fn))

	return &function{
		name:   fn.Name(),
		args:   args,
		retval: retval,
		fn:     vf,
	}
}

// hostFunc decorates the given ffi.Func so it can be called by our pkg wrapper.
//
// This is responsible for calling the actual golang host function.
func hostFunc(fn sdk.Function) HostFunc {
	return func(args ...wasmer.Value) (interface{}, error) {
		hostArgs := hostFuncArgs(args...)
		return fn.Func()(hostArgs...)
	}
}

// hostFuncArgs translates array of pkg function arg types to golang types.
func hostFuncArgs(args ...wasmer.Value) []interface{} {
	retval := make([]interface{}, len(args))
	for i, arg := range args {
		retval[i] = arg.I32()
	}
	return retval
}

// valueFunc decorates the given HostFunc so it can be called by pkg.Function.
// This is responsible for translating a WASM function invocation into golang and back.
func valueFunc(fn HostFunc) ValueFunc {
	return func(args []wasmer.Value) ([]wasmer.Value, error) {
		result, err := fn(args...)
		if err != nil {
			return nil, err
		}

		retval := make([]wasmer.Value, 0)
		if result != nil {
			retval = append(retval, wasmer.NewValue(result, wasmer.I32))
		}

		return retval, nil
	}
}
