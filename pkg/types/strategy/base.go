package strategy

import (
	"errors"
	"sync"
	"time"
)

// ExecutionConfig defines how a strategy should be executed
type ExecutionConfig struct {
	ExecutionInterval time.Duration
}

// BaseStrategyConfig holds configuration for creating a base strategy
type BaseStrategyConfig struct {
	Name        StrategyName
	Description string
	RiskLevel   RiskLevel
	Type        StrategyType

	ExecutionConfig *ExecutionConfig
}

// BaseStrategy provides common functionality for all strategies
type BaseStrategy struct {
	name         StrategyName
	description  string
	riskLevel    RiskLevel
	strategyType StrategyType
	enabled      bool
	lastRunAt    time.Time

	executionConfig *ExecutionConfig

	mu sync.RWMutex
}

// NewBaseStrategy creates a new base strategy with the provided configuration
func NewBaseStrategy(config BaseStrategyConfig) Strategy {
	now := time.Now()
	return &BaseStrategy{
		name:            config.Name,
		description:     config.Description,
		riskLevel:       config.RiskLevel,
		strategyType:    config.Type,
		lastRunAt:       now,
		executionConfig: config.ExecutionConfig,
	}
}

func (s *BaseStrategy) GetSignals(ctx StrategyContext) ([]*Signal, error) {
	return nil, errors.New("GetSignals not implemented")
}

func (s *BaseStrategy) GetLastRunAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRunAt
}

// GetName returns the strategy name
func (s *BaseStrategy) GetName() StrategyName {
	return s.name
}

// GetDescription returns the strategy description
func (s *BaseStrategy) GetDescription() string {
	return s.description
}

// GetRiskLevel returns the strategy risk level
func (s *BaseStrategy) GetRiskLevel() RiskLevel {
	return s.riskLevel
}

// GetStrategyType returns the strategy type
func (s *BaseStrategy) GetStrategyType() StrategyType {
	return s.strategyType
}

func (s *BaseStrategy) ExecutionConfig() *ExecutionConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.executionConfig
}

func (s *BaseStrategy) WithExecutionConfig(cfg *ExecutionConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executionConfig = cfg
}

func RecordExecution(strat Strategy, t time.Time) {
	// Type assert to concrete BaseStrategy
	if base, ok := strat.(*BaseStrategy); ok {
		base.mu.Lock()
		defer base.mu.Unlock()
		base.lastRunAt = t
	}
}
