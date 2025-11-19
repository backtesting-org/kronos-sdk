package position

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

func (ds *dataStore) AddTradeToStrategy(strategyName strategy.StrategyName, trade connector.Trade) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getExecutions()
	execution := current[strategyName]

	if execution == nil {
		execution = &strategy.StrategyExecution{
			Orders: []connector.Order{},
			Trades: []connector.Trade{},
		}
	}

	// Add trade
	execution.Trades = append(execution.Trades, trade)

	// Store updated map
	updated := make(map[strategy.StrategyName]*strategy.StrategyExecution, len(current)+1)
	for k, v := range current {
		updated[k] = v
	}
	updated[strategyName] = execution
	ds.executions.Store(updated)
}

func (ds *dataStore) GetTradesForStrategy(strategyName strategy.StrategyName) []connector.Trade {
	executions := ds.getExecutions()
	execution := executions[strategyName]

	if execution == nil {
		return []connector.Trade{}
	}

	return execution.Trades
}
