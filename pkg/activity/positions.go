package activity

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	storeActivity "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// positions wraps the internal position store with read-only access
type positions struct {
	store storeActivity.Positions
}

// NewPositions creates a new read-only positions accessor
func NewPositions(store storeActivity.Positions) kronosActivity.Positions {
	return &positions{store: store}
}

// GetStrategyExecution retrieves the execution for a strategy
func (p *positions) GetStrategyExecution(strategyName strategy.StrategyName) *strategy.StrategyExecution {
	return p.store.GetStrategyExecution(strategyName)
}

// GetAllStrategyExecutions retrieves all strategy executions
func (p *positions) GetAllStrategyExecutions() map[strategy.StrategyName]*strategy.StrategyExecution {
	return p.store.GetAllStrategyExecutions()
}

// GetStrategyForOrder finds which strategy owns a given order
func (p *positions) GetStrategyForOrder(orderID string) (strategy.StrategyName, bool) {
	return p.store.GetStrategyForOrder(orderID)
}

// GetTotalOrderCount returns the total number of orders across all strategies
func (p *positions) GetTotalOrderCount() int64 {
	return p.store.GetTotalOrderCount()
}

// GetTradesForStrategy retrieves all trades for a strategy
func (p *positions) GetTradesForStrategy(strategyName strategy.StrategyName) []connector.Trade {
	return p.store.GetTradesForStrategy(strategyName)
}

var _ kronosActivity.Positions = (*positions)(nil)
