package pkg

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal"
	"github.com/flyingdice/whack-sdk/sdk/module"
	"github.com/flyingdice/whack-sdk/sdk/runtime/config"
)

// New creates a new Whack runtime using wasmer (https://wasmer.io/).
func New(mod module.Module, cfg config.Config) (*internal.Runtime, error) {
	return internal.NewRuntime(mod, cfg)
}
