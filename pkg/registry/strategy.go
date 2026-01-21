package registry

import (
	"sync"

	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

type strategyRegistry struct {
	strategies map[strategy.StrategyName]strategy.Strategy
	mu         sync.RWMutex
}

// NewStrategyRegistry creates a new strategy registry
func NewStrategyRegistry() registry.StrategyRegistry {
	return &strategyRegistry{
		strategies: make(map[strategy.StrategyName]strategy.Strategy),
	}
}

func (sr *strategyRegistry) GetStrategy(name strategy.StrategyName) (strategy.Strategy, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	strat, exists := sr.strategies[name]
	return strat, exists
}

func (sr *strategyRegistry) RegisterStrategy(strat strategy.Strategy) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.strategies[strat.GetName()] = strat
}

func (sr *strategyRegistry) RegisterAllStrategies(strategies []strategy.Strategy) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for _, strat := range strategies {
		sr.strategies[strat.GetName()] = strat
	}
}

func (sr *strategyRegistry) GetAllStrategies() []strategy.Strategy {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	strategies := make([]strategy.Strategy, 0, len(sr.strategies))
	for _, strat := range sr.strategies {
		strategies = append(strategies, strat)
	}

	return strategies
}

func (sr *strategyRegistry) GetStrategyCount() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return len(sr.strategies)
}
