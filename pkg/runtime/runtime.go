package runtime

import (
	"context"
	"fmt"

	configTypes "github.com/backtesting-org/kronos-sdk/pkg/types/config"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/lifecycle"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/plugin"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/runtime"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

type rt struct {
	pluginManager     plugin.Manager
	connectorRegistry registry.ConnectorRegistry
	assetRegistry     registry.AssetRegistry
	strategyRegistry  registry.StrategyRegistry
	configLoader      configTypes.StartupConfigLoader
	controller        lifecycle.Controller
	logger            logging.ApplicationLogger
	loadedStrategy    strategy.Strategy
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewRuntime(
	pluginManager plugin.Manager,
	connectorRegistry registry.ConnectorRegistry,
	assetRegistry registry.AssetRegistry,
	strategyRegistry registry.StrategyRegistry,
	configLoader configTypes.StartupConfigLoader,
	controller lifecycle.Controller,
	logger logging.ApplicationLogger,
) runtime.Runtime {
	return &rt{
		pluginManager:     pluginManager,
		connectorRegistry: connectorRegistry,
		assetRegistry:     assetRegistry,
		strategyRegistry:  strategyRegistry,
		configLoader:      configLoader,
		controller:        controller,
		logger:            logger,
	}
}

// Start runs a strategy in plugin mode
func (r *rt) Start(configPath string, kronosPath string) error {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	// Load all config
	cfg, err := r.configLoader.LoadForStrategy(configPath, kronosPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize connectors
	connectorNames, err := r.initializeConnectors(cfg.ConnectorConfigs)
	if err != nil {
		return err
	}

	// Register assets
	r.registerAssets(cfg.AssetConfigs)

	// Boot in plugin mode with execution config from StartupConfig
	return r.boot(r.ctx, runtime.BootConfig{
		Mode:            runtime.BootModePlugin,
		StrategyPath:    cfg.PluginPath,
		ConnectorNames:  connectorNames,
		ExecutionConfig: cfg.ExecutionConfig,
	})
}

// StartStandalone runs a strategy in standalone mode (debuggable)
func (r *rt) StartStandalone(
	strat strategy.Strategy,
	configPath string,
	kronosPath string,
) error {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	// Load all config
	cfg, err := r.configLoader.LoadForStrategy(configPath, kronosPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize connectors
	connectorNames, err := r.initializeConnectors(cfg.ConnectorConfigs)
	if err != nil {
		return err
	}

	// Register assets
	r.registerAssets(cfg.AssetConfigs)

	// Boot in standalone mode with execution config from StartupConfig
	return r.boot(r.ctx, runtime.BootConfig{
		Mode:            runtime.BootModeStandalone,
		Strategy:        strat,
		ConnectorNames:  connectorNames,
		ExecutionConfig: cfg.ExecutionConfig,
	})
}

// Stop gracefully shuts down
func (r *rt) Stop() error {
	r.logger.Info("🛑 Stopping runtime...")

	if r.cancel != nil {
		r.cancel()
	}

	if err := r.controller.Stop(r.ctx); err != nil {
		r.logger.Error(fmt.Sprintf("Failed to stop controller: %v", err))
		return err
	}

	r.logger.Info("✅ Runtime stopped")
	return nil
}

// initializeConnectors initializes all connectors and returns their names
func (r *rt) initializeConnectors(connectors map[connector.ExchangeName]connector.Config) ([]connector.ExchangeName, error) {
	names := make([]connector.ExchangeName, 0, len(connectors))

	for name, cfg := range connectors {
		conn, exists := r.connectorRegistry.GetConnector(name)
		if !exists {
			r.logger.Warn(fmt.Sprintf("connector %s not registered", name))
			continue
		}

		if err := conn.Initialize(cfg); err != nil {
			r.logger.Error(fmt.Sprintf("connector %s init failed: %s", name, err.Error()))
			return nil, fmt.Errorf("failed to initialize connector %s: %w", name, err)
		}

		names = append(names, name)
		if err := r.connectorRegistry.MarkConnectorReady(name); err != nil {
			return nil, err
		}
	}

	r.logger.Info("Initialized connectors", "count", len(names))
	return names, nil
}

// registerAssets registers all assets with their instruments
func (r *rt) registerAssets(assets map[portfolio.Asset][]connector.Instrument) {
	for asset, instruments := range assets {
		for _, instr := range instruments {
			r.assetRegistry.RegisterAsset(asset, instr)
		}
	}
	r.logger.Info("Registered assets", "count", len(assets))
}

// boot executes the core boot sequence
func (r *rt) boot(ctx context.Context, cfg runtime.BootConfig) error {
	r.logger.Info("🔧 Booting...", "mode", cfg.Mode)

	var strat strategy.Strategy
	var err error

	switch cfg.Mode {
	case runtime.BootModeStandalone:
		r.logger.Info("Using provided strategy instance")
		if cfg.Strategy == nil {
			return fmt.Errorf("no strategy provided in standalone mode")
		}
		strat = cfg.Strategy

	default:
		r.logger.Info("Loading strategy plugin...", "path", cfg.StrategyPath)
		strat, err = r.pluginManager.LoadStrategyPlugin(cfg.StrategyPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}
	}

	// Set execution config on strategy (before registration)
	// Nil is fine - orchestrator will use defaults
	if cfg.ExecutionConfig != nil {
		strat.WithExecutionConfig(cfg.ExecutionConfig)
	}

	r.loadedStrategy = strat
	r.logger.Info("Strategy loaded", "name", strat.GetName())

	// Register strategy (orchestrator can now read execution config from strategy)
	r.strategyRegistry.RegisterStrategy(strat)

	if err := r.controller.Start(ctx, strat.GetName()); err != nil {
		return fmt.Errorf("failed to start lifecycle: %w", err)
	}

	r.logger.Info("✅ Runtime started")
	return nil
}
