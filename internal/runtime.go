package internal

import (
	"github.com/flyingdice/whack-sdk/pkg/sdk"
	"github.com/flyingdice/whack-sdk/pkg/sdk/runtime"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

var _ sdk.Runtime = (*Runtime)(nil)

// TODO (ahawker) Cache compiled modules?
// TODO (ahawker) Reactor support? Do we call start/init during creation or lazy?

// Runtime encapsulates all Wasmer state necessary to run WASM code.
type Runtime struct {
	cfg    runtime.Config
	engine *wasmer.Engine
	env    *wasmer.WasiEnvironment
	module *wasmer.Module
	store  *wasmer.Store
	pool   sdk.InstancePool
}

// New creates a new Instance for the runtime.
func (r *Runtime) New() (sdk.Instance, error) {
	instance, err := NewInstance(r, r.cfg.Host().Imports())
	if err != nil {
		return nil, err
	}
	return r.pool.Set(instance)
}

// Get returns the pool instance for the given wrn.
func (r *Runtime) Get(wrn sdk.WRN) (sdk.Instance, error) {
	return r.pool.Get(wrn)
}

// NewRuntime creates a new runtime.
func NewRuntime(mod sdk.Module, cfg runtime.Config) (*Runtime, error) {
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

	// Create runtime instance pool.
	pool, err := runtime.NewPool()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create instance pool")
	}

	return &Runtime{
		env:    env,
		engine: engine,
		module: compiled,
		store:  store,
		cfg:    cfg,
		pool:   pool,
	}, nil
}

// wasiEnv creates a Wasi environment for the given runtime wasi config.
func wasiEnv(wrn sdk.WRN, config runtime.WasiConfig) (*wasmer.WasiEnvironment, error) {
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
