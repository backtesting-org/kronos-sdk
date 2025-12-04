package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/data/ingestors"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	lifecycleTypes "github.com/backtesting-org/kronos-sdk/pkg/types/lifecycle"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
)

// controller implements the Controller interface for controlling SDK lifecycle
type controller struct {
	// Component references
	marketCoordinator   ingestors.MarketDataCoordinator
	positionCoordinator ingestors.PositionCoordinator
	connectorRegistry   registry.ConnectorRegistry
	healthStore         health.HealthStore
	orchestrator        lifecycleTypes.Orchestrator
	logger              logging.ApplicationLogger

	// Lifecycle state
	state     lifecycleTypes.State
	stateMu   sync.RWMutex
	readyChan chan struct{}
	readyOnce sync.Once
}

// NewController creates a new SDK lifecycle controller
func NewController(
	marketCoordinator ingestors.MarketDataCoordinator,
	positionCoordinator ingestors.PositionCoordinator,
	connectorRegistry registry.ConnectorRegistry,
	healthStore health.HealthStore,
	orchestrator lifecycleTypes.Orchestrator,
	logger logging.ApplicationLogger,
) lifecycleTypes.Controller {
	return &controller{
		marketCoordinator:   marketCoordinator,
		positionCoordinator: positionCoordinator,
		connectorRegistry:   connectorRegistry,
		healthStore:         healthStore,
		orchestrator:        orchestrator,
		logger:              logger,
		state:               lifecycleTypes.StateCreated,
		readyChan:           make(chan struct{}),
	}
}

// Start starts the SDK and all its components
func (c *controller) Start(ctx context.Context) error {
	c.stateMu.Lock()
	if c.state != lifecycleTypes.StateCreated && c.state != lifecycleTypes.StateStopped {
		currentState := c.state
		c.stateMu.Unlock()
		return fmt.Errorf("cannot start: current state is %v", currentState)
	}
	c.state = lifecycleTypes.StateStarting
	c.stateMu.Unlock()

	c.logger.Info("🚀 Initializing Kronos SDK...")

	// Validate connectors are ready (must be initialized before Start())
	if err := c.validateConnectorsReady(); err != nil {
		c.stateMu.Lock()
		c.state = lifecycleTypes.StateCreated
		c.stateMu.Unlock()
		return err
	}

	// Start coordinators
	c.logger.Info("⚡ Starting data coordinators...")

	// Start position coordinator (if needed)
	c.logger.Info("  📊 Starting position tracking...")
	if err := c.positionCoordinator.Start(ctx); err != nil {
		c.stateMu.Lock()
		c.state = lifecycleTypes.StateCreated
		c.stateMu.Unlock()
		return fmt.Errorf("failed to start position coordinator: %w", err)
	}
	c.logger.Info("  ✓ Position tracking ready")
	
	// Start market data ingestion
	c.logger.Info("  📈 Starting market data ingestion...")
	if err := c.marketCoordinator.StartDataCollection(ctx); err != nil {
		c.stateMu.Lock()
		c.state = lifecycleTypes.StateCreated
		c.stateMu.Unlock()
		return fmt.Errorf("failed to start market data collection: %w", err)
	}

	c.logger.Info("  ✓ Market data ingestion ready")

	// Start orchestrator
	c.logger.Info("  🎯 Starting strategy orchestrator...")
	if err := c.orchestrator.Start(ctx); err != nil {
		c.stateMu.Lock()
		c.state = lifecycleTypes.StateCreated
		c.stateMu.Unlock()
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	c.logger.Info("  ✓ Strategy orchestrator ready")

	// Start runtime health monitoring
	go c.monitorHealth(ctx)

	// Mark as ready
	c.stateMu.Lock()
	c.state = lifecycleTypes.StateReady
	c.stateMu.Unlock()

	c.readyOnce.Do(func() {
		close(c.readyChan)
	})

	c.logger.Info("✅ Kronos SDK ready - strategies can now execute")
	return nil
}

// Stop gracefully shuts down the SDK
func (c *controller) Stop(ctx context.Context) error {
	c.stateMu.Lock()
	if c.state == lifecycleTypes.StateStopped || c.state == lifecycleTypes.StateStopping {
		c.stateMu.Unlock()
		return nil
	}
	c.state = lifecycleTypes.StateStopping
	c.stateMu.Unlock()

	c.logger.Info("Stopping Kronos SDK...")

	// Stop orchestrator first (stop generating new signals)
	if err := c.orchestrator.Stop(ctx); err != nil {
		c.logger.Error("Error stopping orchestrator: %v", err)
	}

	// Stop market data ingestion
	if err := c.marketCoordinator.StopDataCollection(); err != nil {
		c.logger.Error("Error stopping market data collection: %v", err)
	}

	// Stop position coordinator
	if c.positionCoordinator != nil {
		if err := c.positionCoordinator.Stop(); err != nil {
			c.logger.Error("Error stopping position coordinator: %v", err)
		}
	}

	c.stateMu.Lock()
	c.state = lifecycleTypes.StateStopped
	c.stateMu.Unlock()

	c.logger.Info("Kronos SDK stopped")
	return nil
}

// WaitUntilReady blocks until the SDK is ready or context is cancelled
func (c *controller) WaitUntilReady(ctx context.Context) error {
	select {
	case <-c.readyChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// State returns the current lifecycle state
func (c *controller) State() lifecycleTypes.State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

// IsReady returns true if the SDK is ready
func (c *controller) IsReady() bool {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state == lifecycleTypes.StateReady
}

// validateConnectorsReady validates that at least one connector is ready.
// Applications must initialize and mark connectors ready BEFORE calling Start().
// This method does not wait - it only verifies the precondition.
func (c *controller) validateConnectorsReady() error {
	readyConnectors := c.connectorRegistry.GetReadyConnectors()
	if len(readyConnectors) == 0 {
		return fmt.Errorf("no connectors marked as ready - initialize and mark connectors ready before calling Start()")
	}

	c.logger.Info("✓ Validated %d connector(s) ready", len(readyConnectors))

	return nil
}

// monitorHealth continuously monitors system health and reports aggregated errors
func (c *controller) monitorHealth(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report := c.healthStore.GetSystemHealth()

			fmt.Println("Starting system health monitoring...")

			if report.HasErrors {
				c.logger.Warn("⚠️  System health report:")

				// Log connector errors
				if len(report.ConnectorErrors.Errors) > 0 {
					c.logger.Warn("  🔴 Connector errors:")
					for connector, err := range report.ConnectorErrors.Errors {
						c.logger.Warn("    - %s [%s]: %v", connector, err.State, err.Error)
					}
				}

				// Log data flow errors
				if len(report.DataFlowErrors.Errors) > 0 {
					c.logger.Warn("  🟡 Data flow errors:")
					for connector, dataTypeErrors := range report.DataFlowErrors.Errors {
						for dataType, err := range dataTypeErrors {
							c.logger.Warn("    - %s:%s [%d errors]: %v", connector, dataType, err.ErrorCount, err.Error)
						}
					}
				}
			}
		}
	}
}
