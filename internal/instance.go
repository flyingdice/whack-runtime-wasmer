package internal

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal/consts"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// Instance wraps an active Wasmer instance with a helpful
// API for interacting with it.
type Instance struct {
	env      *wasmer.WasiEnvironment
	instance *wasmer.Instance
}

// Stdout returns the stdout stream for the instance.
//
// This will be empty if CaptureStdout was not set on WasiStateBuilder.
func (i *Instance) Stdout() string { return string(i.env.ReadStdout()) }

// Stderr returns the stderr stream for the instance.
//
// This will be empty if CaptureStderr was not set on WasiStateBuilder.
func (i *Instance) Stderr() string { return string(i.env.ReadStderr()) }

// Call invokes an exported function by name with the given arguments.
func (i *Instance) Call(name string, args ...interface{}) (interface{}, error) {
	function, err := i.instance.Exports.GetFunction(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get exported function %s", name)
	}
	if function == nil {
		return nil, errors.Wrapf(err, "exported function %s not found", name)
	}
	return function(args...)
}

// Read memory of given length at a specific pointer location.
func (i *Instance) Read(ptr, length int32) ([]byte, error) {
	memory, err := i.instance.Exports.GetMemory(consts.ExportedMemoryName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exported memory")
	}

	data := memory.Data()
	if int(length) > len(data) {
		return nil, errors.Errorf("expected %d bytes; memory only %d bytes", length, len(data))
	}

	buffer := make([]byte, length)

	if read := copy(buffer, data[ptr:ptr+length]); int32(read) != length {
		return nil, errors.Errorf("expected to read %d; got %d", length, read)
	}

	return buffer, nil
}

// Write bytes to memory at a specific pointer location.
func (i *Instance) Write(ptr int32, bytes []byte) (int32, error) {
	memory, err := i.instance.Exports.GetMemory(consts.ExportedMemoryName)
	if err != nil {
		return -1, errors.Wrap(err, "failed to get exported memory")
	}

	length := int32(len(bytes))
	data := memory.Data()

	if written := copy(data[ptr:ptr+length], data); int32(written) != length {
		return 0, errors.Errorf("expected to write %d; got %d", length, written)
	}

	return length, nil
}

func NewInstance(rt *Runtime) (*Instance, error) {
	// Create Instance for module module.
	inst, err := wasmer.NewInstance(rt.module, rt.imports)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Instance")
	}

	// Fetch and invoke wasi start ffi if one exists (commands).
	// Note: Start functions take no arguments and have no return value.
	start, err := wasiStartFunc(inst)
	if err != nil {
		return nil, err
	}
	if start != nil {
		if _, err := start(); err != nil {
			return nil, errors.Wrap(err, "failed to invoke wasi start ffi")
		}
	}

	// Fetch and invoke wasi initialize ffi if one exists (reactors).
	// Note: Init functions take no arguments and have no return value.
	init, err := wasiInitFunc(inst)
	if err != nil {
		return nil, err
	}
	if init != nil {
		if _, err := init(); err != nil {
			return nil, errors.Wrap(err, "failed to invoke wasi init ffi")
		}
	}

	return &Instance{
		instance: inst,
		env:      rt.env,
	}, nil
}

// wasiStartFunc returns a NativeFunction start ffi for the Instance.
//
// If no NativeFunction is found, nil is returned without error.
//
// https://github.com/WebAssembly/WASI/blob/main/design/application-abi.md
func wasiStartFunc(instance *wasmer.Instance) (wasmer.NativeFunction, error) {
	start, err := instance.Exports.GetWasiStartFunction()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wasi start function")
	}
	if start != nil {
		return start, nil
	}
	start, err = instance.Exports.GetFunction(consts.StartFunctionName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get wasi %s exported function", consts.StartFunctionName)
	}
	if start != nil {
		return start, nil
	}
	return nil, nil
}

// wasiInitFunc returns a NativeFunction initialize ffi for the Instance.
//
// If no NativeFunction is found, nil is returned without error.
//
// https://github.com/WebAssembly/WASI/blob/main/design/application-abi.md
func wasiInitFunc(instance *wasmer.Instance) (wasmer.NativeFunction, error) {
	init, _ := instance.Exports.GetFunction(consts.InitFunctionName)
	if init != nil {
		return init, nil
	}
	return nil, nil
}
