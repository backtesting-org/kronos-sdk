package position

import (
	portfolioTypes "github.com/backtesting-org/kronos-sdk/pkg/types/stores/activity"
)

func NewStore() portfolioTypes.Positions {
	ds := &dataStore{}
	ds.executions.Store(make(portfolioTypes.StrategyExecutionMap))
	ds.lastUpdated.Store(make(portfolioTypes.LastUpdatedMap))
	return ds
}

func (ds *dataStore) Clear() {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.executions.Store(make(portfolioTypes.StrategyExecutionMap))
	ds.lastUpdated.Store(make(portfolioTypes.LastUpdatedMap))
}

// Helper methods to get typed data from atomic.Value
func (ds *dataStore) getExecutions() portfolioTypes.StrategyExecutionMap {
	if v := ds.executions.Load(); v != nil {
		return v.(portfolioTypes.StrategyExecutionMap)
	}
	return make(portfolioTypes.StrategyExecutionMap)
}

var _ portfolioTypes.Positions = (*dataStore)(nil)
