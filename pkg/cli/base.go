package cli

import (
	"context"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/log"
)

// Base is a struct that can be embedded in commands to provide common functionality.
type Base struct {
	Logger log.Logger      `internal:"ignore"`
	Ctx    context.Context `internal:"ignore"`
}
