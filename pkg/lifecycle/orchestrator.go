package lifecycle

import (
	"context"
	"sync"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/monitoring/profiling"
	"github.com/backtesting-org/kronos-sdk/pkg/types/execution"
	lifecycleTypes "github.com/backtesting-org/kronos-sdk/pkg/types/lifecycle"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	profileTypes "github.com/backtesting-org/kronos-sdk/pkg/types/monitoring/profiling"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

const (
	// defaultTickInterval is how often the orchestrator checks if strategies should execute
	// This is intentionally fast - per-strategy ExecutionConfig handles the actual timing
	defaultTickInterval = 50 * time.Millisecond
)

type orchestrator struct {
	executor         execution.Executor
	strategyRegistry registry.StrategyRegistry
	logger           logging.ApplicationLogger
	timeProvider     temporal.TimeProvider
	profilingStore   profileTypes.ProfilingStore  // Optional profiling
	anomalyDetector  profileTypes.AnomalyDetector // Optional anomaly detection

	// Execution control
	ctx    context.Context
	cancel context.CancelFunc

	ticker *time.Ticker

	// Track in-flight strategy executions for graceful shutdown
	executionWg sync.WaitGroup

	// Prevent concurrent executions of same strategy
	strategyMutexes map[strategy.StrategyName]*sync.Mutex
	mutexMapLock    sync.RWMutex
}

// NewOrchestrator creates a new strategy orchestrator
func NewOrchestrator(
	executor execution.Executor,
	strategyRegistry registry.StrategyRegistry,
	logger logging.ApplicationLogger,
	timeProvider temporal.TimeProvider,
	profilingStore profileTypes.ProfilingStore, // Optional: can be nil
	anomalyDetector profileTypes.AnomalyDetector, // Optional: can be nil
) lifecycleTypes.Orchestrator {
	return &orchestrator{
		executor:         executor,
		strategyRegistry: strategyRegistry,
		logger:           logger,
		timeProvider:     timeProvider,
		profilingStore:   profilingStore,
		anomalyDetector:  anomalyDetector,
		strategyMutexes:  make(map[strategy.StrategyName]*sync.Mutex),
	}
}

// Start begins orchestration
func (o *orchestrator) Start(ctx context.Context) error {
	if o.cancel != nil {
		o.logger.Warn("Orchestrator already started")
		return nil
	}

	o.ctx, o.cancel = context.WithCancel(ctx)
	o.ticker = time.NewTicker(defaultTickInterval)

	o.logger.Info("🎯 Starting strategy orchestrator")

	go o.executionLoop()

	o.logger.Info("✅ Strategy orchestrator started")
	return nil
}

// Stop gracefully stops orchestration
func (o *orchestrator) Stop(_ context.Context) error {
	if o.cancel == nil {
		return nil
	}

	o.logger.Info("🛑 Stopping strategy orchestrator")

	// Stop the ticker
	if o.ticker != nil {
		o.ticker.Stop()
	}

	// Cancel the orchestrator's context
	o.cancel()
	o.cancel = nil

	// Wait for all in-flight strategy executions to complete
	o.executionWg.Wait()

	o.logger.Info("✅ Strategy orchestrator stopped")
	return nil
}

// executionLoop checks strategies at fixed intervals
func (o *orchestrator) executionLoop() {
	for {
		select {
		case <-o.ctx.Done():
			return
		case <-o.ticker.C:
			o.executeEnabledStrategies()
		}
	}
}

// executeEnabledStrategies runs all registered strategies
func (o *orchestrator) executeEnabledStrategies() {
	strategies := o.strategyRegistry.GetAllStrategies()

	if len(strategies) == 0 {
		return
	}

	// Execute each strategy concurrently without blocking
	// Per-strategy mutex prevents same strategy running twice
	// executionWg tracks in-flight executions for graceful shutdown
	for _, strat := range strategies {
		go o.executeStrategy(strat)
	}
}

// getStrategyMutex returns a mutex for the given strategy, creating one if needed
func (o *orchestrator) getStrategyMutex(strategyName strategy.StrategyName) *sync.Mutex {
	o.mutexMapLock.RLock()
	mutex, exists := o.strategyMutexes[strategyName]
	o.mutexMapLock.RUnlock()

	if exists {
		return mutex
	}

	o.mutexMapLock.Lock()
	defer o.mutexMapLock.Unlock()

	// Double-check after acquiring write lock
	if mutex, exists := o.strategyMutexes[strategyName]; exists {
		return mutex
	}

	mutex = &sync.Mutex{}
	o.strategyMutexes[strategyName] = mutex
	return mutex
}

// shouldExecuteStrategy checks if a strategy should execute
func (o *orchestrator) shouldExecuteStrategy(strat strategy.Strategy) bool {
	if strat.ExecutionConfig() == nil {
		return true
	}

	config := strat.ExecutionConfig()

	lastRun := strat.GetLastRunAt()
	if o.timeProvider.Now().Sub(lastRun) < config.ExecutionInterval {
		return false
	}

	return true
}

// executeStrategy runs a single strategy
func (o *orchestrator) executeStrategy(strat strategy.Strategy) {
	// Check if strategy should execute based on its config
	if !o.shouldExecuteStrategy(strat) {
		return
	}

	// Track this execution for graceful shutdown
	o.executionWg.Add(1)
	defer o.executionWg.Done()

	defer func() {
		if r := recover(); r != nil {
			o.logger.Error("Strategy %s panicked: %v", strat.GetName(), r)
		}
	}()

	// Prevent concurrent executions of the same strategy
	strategyMutex := o.getStrategyMutex(strat.GetName())
	strategyMutex.Lock()
	defer strategyMutex.Unlock()

	// Create context for this strategy execution
	ctx := strategy.NewStrategyContext(context.Background(), strat.GetName())

	// Add profiling context if profiling is enabled
	var profilingCtx profileTypes.Context
	if o.profilingStore != nil {
		profilingCtx = o.profilingStore.NewContext(string(strat.GetName()))
		ctx = strategy.NewStrategyContext(
			profiling.WithContext(ctx, profilingCtx),
			strat.GetName(),
		)
	}

	startTime := o.timeProvider.Now()
	signals, err := strat.GetSignals(ctx)
	duration := o.timeProvider.Since(startTime)

	//strategy.RecordExecution(strat, startTime)

	// Finalize profiling if enabled
	if profilingCtx != nil {
		metrics := profilingCtx.Finalize(err == nil, err)
		o.profilingStore.RecordExecution(metrics)

		// Check for performance anomalies
		if o.anomalyDetector != nil {
			alert := o.anomalyDetector.CheckExecution(string(strat.GetName()), duration)
			if alert.Severity != profileTypes.OK {
				o.logger.Warn("Performance alert for %s: %s", strat.GetName(), alert.Message)
			}
			// Update baseline for future comparisons
			o.anomalyDetector.UpdateBaseline(string(strat.GetName()), duration)
		}
	}

	if err != nil {
		o.logger.Error("Strategy %s failed: %v (duration: %v)", strat.GetName(), err, duration)
		return
	}

	// Execute signals
	for _, signal := range signals {
		if err := o.executor.ExecuteSignal(signal); err != nil {
			o.logger.Error("Signal execution failed for %s: %v", strat.GetName(), err)
		}
	}
}
