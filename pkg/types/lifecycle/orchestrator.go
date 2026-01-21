package lifecycle

import (
	"context"
)

// Orchestrator manages strategy execution lifecycle
type Orchestrator interface {
	// Start begins orchestration
	Start(ctx context.Context) error

	// Stop gracefully stops orchestration
	Stop(ctx context.Context) error
}
