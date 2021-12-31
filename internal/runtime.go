package internal

import (
	"github.com/flyingdice/whack-sdk/sdk/id"
	"github.com/flyingdice/whack-sdk/sdk/module"
	"github.com/flyingdice/whack-sdk/sdk/runtime/config"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// TODO (ahawker) Cache compiled modules?
// TODO (ahawker) Reactor support? Do we call start/init during creation or lazy?

// Runtime encapsulates all Wasmer state necessary to run WASM code.
type Runtime struct {
	engine  *wasmer.Engine
	env     *wasmer.WasiEnvironment
	imports *wasmer.ImportObject
	module  *wasmer.Module
	store   *wasmer.Store
}

// NewInstance creates a new Instance for the runtime.
func (r *Runtime) NewInstance() (*Instance, error) { return NewInstance(r) }

// NewRuntime creates a new runtime.
func NewRuntime(mod module.Module, cfg config.Config) (*Runtime, error) {
	// Create global state.
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compile raw module bytes into code (WAT).
	compiled, err := wasmer.NewModule(store, mod.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile mod")
	}

	// Create Wasi state environment.
	env, err := wasiEnv(mod.Wrn(), cfg.Wasi())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wasi env")
	}

	// TODO i need to create imports from myself
	// and from the whack-sdk which is the dep[

	// Create global imports with our extended environment.
	imports, err := importObject(env, store, compiled, cfg.Host().Imports())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create import object")
	}

	return &Runtime{
		env:     env,
		engine:  engine,
		imports: imports,
		module:  compiled,
		store:   store,
	}, nil
}

// wasiEnv creates a Wasi environment for the given runtime wasi config.
func wasiEnv(wrn id.WRN, config config.WasiConfig) (*wasmer.WasiEnvironment, error) {
	builder := wasmer.NewWasiStateBuilder(wrn.Name())

	// Arguments
	for _, arg := range config.Arguments() {
		builder.Argument(arg)
	}

	// Environment Variables
	for key, val := range config.EnvVars() {
		builder.Environment(key, val)
	}

	// Map directories
	for name, path := range config.Directories() {
		builder.MapDirectory(name, path)
	}

	// Stdio
	if config.CaptureStdout() {
		builder.CaptureStdout()
	}
	if config.CaptureStderr() {
		builder.CaptureStderr()
	}

	// Workdir
	builder.PreopenDirectory(config.Workdir())

	return builder.Finalize()
}
