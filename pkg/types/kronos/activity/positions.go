package activity

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// Positions provides read-only access to position data
type Positions interface {
	// Strategy queries
	GetStrategyExecution(strategy strategy.StrategyName) *strategy.StrategyExecution
	GetAllStrategyExecutions() map[strategy.StrategyName]*strategy.StrategyExecution

	// Order queries
	GetStrategyForOrder(orderID string) (strategy.StrategyName, bool)
	GetTotalOrderCount() int64

	// Trade queries
	GetTradesForStrategy(strategy strategy.StrategyName) []connector.Trade
}
