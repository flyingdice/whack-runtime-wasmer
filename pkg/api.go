package pkg

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal"
	"github.com/flyingdice/whack-sdk/sdk/module"
	"github.com/flyingdice/whack-sdk/sdk/runtime"
)

// New creates a new whack runtime using Wasmer (https://wasmer.io/).
func New(mod module.Module, cfg runtime.Config) (*internal.Runtime, error) {
	return internal.NewRuntime(mod, cfg)
}
