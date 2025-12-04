package health

import (
	"time"

	healthTypes "github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

// HealthStore aggregates error reports from both stores into a unified view
type healthStore struct {
	connectorErrors healthTypes.ConnectorErrorStore
	coordinatorData healthTypes.CoordinatorHealthStore
	timeProvider    temporal.TimeProvider
	startedAt       time.Time
}

// NewHealthStore creates a unified health reporter
func NewHealthStore(
	timeProvider temporal.TimeProvider,
	connectorErrors healthTypes.ConnectorErrorStore,
	coordinatorData healthTypes.CoordinatorHealthStore,
) healthTypes.HealthStore {
	return &healthStore{
		timeProvider:    timeProvider,
		connectorErrors: connectorErrors,
		coordinatorData: coordinatorData,
		startedAt:       timeProvider.Now(),
	}
}

// GetSystemHealth returns aggregated health report combining both stores
func (h *healthStore) GetSystemHealth() *healthTypes.SystemHealthReport {
	connReport := h.connectorErrors.GetErrorReport()
	dataReport := h.coordinatorData.GetErrorReport()

	hasErrors := len(connReport.Errors) > 0 || len(dataReport.Errors) > 0
	overallState := healthTypes.StateConnected
	if hasErrors {
		overallState = healthTypes.StateDegraded
	}

	return &healthTypes.SystemHealthReport{
		OverallState:    overallState,
		ConnectorErrors: connReport,
		DataFlowErrors:  dataReport,
		StartedAt:       h.startedAt,
		HasErrors:       hasErrors,
	}
}
