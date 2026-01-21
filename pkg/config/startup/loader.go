package startup

import (
	"fmt"
	"path/filepath"

	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

type startupConfigLoader struct {
	strategySvc   config.StrategyConfig
	connectorSvc  config.ConnectorService
	configuration config.Configuration
	logger        logging.ApplicationLogger
}

// NewStartupConfigLoader creates a new startup config loader
func NewStartupConfigLoader(
	strategySvc config.StrategyConfig,
	connectorSvc config.ConnectorService,
	configuration config.Configuration,
	logger logging.ApplicationLogger,
) config.StartupConfigLoader {
	return &startupConfigLoader{
		strategySvc:   strategySvc,
		connectorSvc:  connectorSvc,
		configuration: configuration,
		logger:        logger,
	}
}

// LoadForStrategy loads ALL configuration needed to run a strategy
func (l *startupConfigLoader) LoadForStrategy(
	strategyDir string,
	kronosPath string,
) (*config.StartupConfig, error) {
	// Load kronos settings first (this sets the path for connector service)
	_, err := l.configuration.LoadSettings(kronosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kronos settings: %w", err)
	}

	l.logger.Info("Loaded kronos settings", "path", kronosPath)

	// Load strategy config
	configPath := filepath.Join(strategyDir, "config.yml")
	stratConfig, err := l.strategySvc.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load strategy config: %w", err)
	}

	l.logger.Info("Loaded strategy config", "name", stratConfig.Name, "exchanges", stratConfig.Exchanges)

	// Get connector configs (now using the loaded settings)
	connectorConfigs, err := l.connectorSvc.GetConnectorConfigsForStrategy(stratConfig.Exchanges)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector configs: %w", err)
	}

	l.logger.Info("Loaded connector configs", "count", len(connectorConfigs))

	// Convert assets to instruments
	assetConfigs := l.convertAssets(stratConfig)

	l.logger.Info("Loaded asset configs", "count", len(assetConfigs))

	// Build plugin path
	strategyName := filepath.Base(strategyDir)
	pluginPath := filepath.Join(strategyDir, strategyName+".so")

	// Extract execution config from strategy config
	execConfig, err := l.extractExecutionConfig(stratConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract execution config: %w", err)
	}

	return &config.StartupConfig{
		Strategy:         stratConfig,
		ConnectorConfigs: connectorConfigs,
		AssetConfigs:     assetConfigs,
		PluginPath:       pluginPath,
		StrategyDir:      strategyDir,
		ExecutionConfig:  execConfig,
	}, nil
}

// convertAssets converts strategy config assets to instrument map
func (l *startupConfigLoader) convertAssets(stratConfig *config.Strategy) map[portfolio.Asset][]connector.Instrument {
	instrumentMap := make(map[portfolio.Asset][]connector.Instrument)

	for _, assets := range stratConfig.Assets {
		for _, asset := range assets {
			instruments := make([]connector.Instrument, 0, len(asset.Instruments))

			for _, instStr := range asset.Instruments {
				switch instStr {
				case "spot":
					instruments = append(instruments, connector.TypeSpot)
				case "perpetual":
					instruments = append(instruments, connector.TypePerpetual)
				default:
					l.logger.Warn("Unknown instrument type", "instrument", instStr)
				}
			}

			if len(instruments) > 0 {
				instrumentMap[portfolio.NewAsset(asset.Symbol)] = instruments
			}
		}
	}

	return instrumentMap
}

// extractExecutionConfig extracts and converts execution config from strategy config
// Returns nil if no execution config is defined (strategy will use global 50ms tick)
func (l *startupConfigLoader) extractExecutionConfig(stratConfig *config.Strategy) (*strategy.ExecutionConfig, error) {
	// Parse execution interval from strategy YAML
	interval, err := stratConfig.ParseExecutionInterval()
	if err != nil {
		return nil, fmt.Errorf("failed to parse execution interval: %w", err)
	}

	// If no execution section in YAML, return nil (use global tick interval)
	if interval == nil {
		return nil, nil
	}

	// Create simple execution config with fixed interval
	return &strategy.ExecutionConfig{
		ExecutionInterval: *interval,
	}, nil
}
