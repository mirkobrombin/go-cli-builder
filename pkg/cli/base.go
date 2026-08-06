package cli

import (
	"context"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/log"
)

// Base is a struct that can be embedded in commands to provide common functionality.
//
// Container is nil unless the application was built with WithContainer.
type Base struct {
	Logger    log.Logger      `internal:"ignore"`
	Ctx       context.Context `internal:"ignore"`
	Container Container       `internal:"ignore"`
}
