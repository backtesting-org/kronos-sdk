package position

import (
	"sync"
	"sync/atomic"
)

type dataStore struct {
	executions  atomic.Value // portfolioTypes.StrategyExecutionMap
	lastUpdated atomic.Value // portfolioTypes.LastUpdatedMap
	mutex       sync.RWMutex
}
