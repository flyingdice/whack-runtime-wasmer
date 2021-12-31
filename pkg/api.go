package pkg

import (
	"github.com/flyingdice/whack-runtime-wasmer/internal"
	"github.com/flyingdice/whack-sdk/pkg/sdk"
	"github.com/flyingdice/whack-sdk/pkg/sdk/runtime"
)

// New creates a new whack runtime using Wasmer (https://wasmer.io/).
func New(mod sdk.Module, cfg runtime.Config) (*internal.Runtime, error) {
	return internal.NewRuntime(mod, cfg)
}
