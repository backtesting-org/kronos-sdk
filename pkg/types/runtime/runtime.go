package runtime

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// BootMode defines how strategies are loaded into the runtime
type BootMode string

const (
	BootModePlugin     BootMode = "plugin"
	BootModeStandalone BootMode = "standalone"
)

// BootConfig holds internal configuration for booting
type BootConfig struct {
	Mode            BootMode
	StrategyPath    string
	Strategy        strategy.Strategy
	ConnectorNames  []connector.ExchangeName
	ExecutionConfig *strategy.ExecutionConfig // Execution timing config (nil = use defaults)
}

// Runtime is the main entry point for running strategies
type Runtime interface {
	// Start runs a strategy in plugin mode
	// Loads config from configPath (strategy dir) and kronosPath (kronos.yml)
	Start(configPath string, kronosPath string) error

	// StartStandalone runs a strategy in standalone mode (debuggable)
	StartStandalone(strategy strategy.Strategy, configPath string, kronosPath string) error

	// Stop gracefully shuts down
	Stop() error
}
