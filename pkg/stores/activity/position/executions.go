package position

import (
	"fmt"

	portfolioTypes "github.com/backtesting-org/kronos-sdk/pkg/types/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

func (ds *dataStore) StoreStrategyExecution(strategyName strategy.StrategyName, execution *strategy.StrategyExecution) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getExecutions()
	updated := make(portfolioTypes.StrategyExecutionMap, len(current)+1)
	for k, v := range current {
		updated[k] = v
	}
	updated[strategyName] = execution
	ds.executions.Store(updated)
}

func (ds *dataStore) GetStrategyExecution(strategyName strategy.StrategyName) *strategy.StrategyExecution {
	executions := ds.getExecutions()
	return executions[strategyName]
}

func (ds *dataStore) UpdateStrategyExecution(strategyName strategy.StrategyName, updateFunc func(*strategy.StrategyExecution)) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getExecutions()
	execution := current[strategyName]
	if execution == nil {
		return fmt.Errorf("strategy execution not found: %s", strategyName)
	}

	// Apply update
	updateFunc(execution)

	// Store updated map
	updated := make(portfolioTypes.StrategyExecutionMap, len(current))
	for k, v := range current {
		updated[k] = v
	}
	updated[strategyName] = execution
	ds.executions.Store(updated)

	return nil
}

func (ds *dataStore) GetAllStrategyExecutions() map[strategy.StrategyName]*strategy.StrategyExecution {
	return ds.getExecutions()
}

func (ds *dataStore) GetTotalOrderCount() int64 {
	executions := ds.getExecutions()
	var total int64
	for _, execution := range executions {
		if execution != nil {
			total += int64(len(execution.Orders))
		}
	}
	return total
}
