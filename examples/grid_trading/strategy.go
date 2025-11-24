package main

import (
	"os"

	sdk "github.com/backtesting-org/kronos-sdk/pkg/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// ExchangesConfig holds the global exchanges configuration
type ExchangesConfig struct {
	Exchanges []struct {
		Name    string   `yaml:"name"`
		Enabled bool     `yaml:"enabled"`
		Assets  []string `yaml:"assets"`
	} `yaml:"exchanges"`
}

// GridTradingConfig holds configuration for grid trading strategy
type GridTradingConfig struct {
	Exchange string `yaml:"exchange"`
	Quantity float64 `yaml:"quantity"`
	Parameters struct {
		GridLevels         int     `yaml:"grid_levels"`
		GridSpacingPercent float64 `yaml:"grid_spacing_percent"`
		PriceUpperBound    float64 `yaml:"price_upper_bound"`
		PriceLowerBound    float64 `yaml:"price_lower_bound"`
	} `yaml:"parameters"`
}

// GridTradingStrategy implements an automated grid trading strategy
type gridTradingStrategy struct {
	strategy.BaseStrategy
	k          *sdk.Kronos
	config     GridTradingConfig
	exConfig   ExchangesConfig
	gridLevels []decimal.Decimal
	spacing    decimal.Decimal
}

// NewGridTrading creates a new grid trading strategy instance
func NewGridTrading(k *sdk.Kronos) strategy.Strategy {
	// Load configuration
	var config GridTradingConfig
	data, err := os.ReadFile("config.yml")
	if err == nil {
		yaml.Unmarshal(data, &config)
	}

	// Load exchanges configuration
	var exConfig ExchangesConfig
	exData, err := os.ReadFile("exchanges.yml")
	if err == nil {
		yaml.Unmarshal(exData, &exConfig)
	}

	// Set defaults if not loaded
	if config.Parameters.GridLevels == 0 {
		config.Parameters.GridLevels = 10
		config.Parameters.GridSpacingPercent = 1.0
		config.Parameters.PriceUpperBound = 50000
		config.Parameters.PriceLowerBound = 40000
	}
	if config.Exchange == "" {
		config.Exchange = "binance"
	}
	if config.Quantity == 0 {
		config.Quantity = 0.01
	}

	return &gridTradingStrategy{
		k:          k,
		config:     config,
		exConfig:   exConfig,
		gridLevels: make([]decimal.Decimal, 0),
	}
}

// initializeGrid sets up the grid levels based on price bounds
func (s *gridTradingStrategy) initializeGrid(lowerBound, upperBound decimal.Decimal, levels int) {
	// Calculate spacing between grid levels
	priceRange := upperBound.Sub(lowerBound)
	s.spacing = priceRange.Div(decimal.NewFromInt(int64(levels)))

	// Create grid levels
	s.gridLevels = make([]decimal.Decimal, levels+1)
	for i := 0; i <= levels; i++ {
		level := lowerBound.Add(s.spacing.Mul(decimal.NewFromInt(int64(i))))
		s.gridLevels[i] = level
	}
}

// GetSignals generates trading signals based on grid levels
func (s *gridTradingStrategy) GetSignals() ([]*strategy.Signal, error) {
	// Determine asset symbol and exchange from exchanges.yml based on config
	exchangeStr := s.config.Exchange
	assetSymbol := "BTC" // default
	exchange := connector.ExchangeName(exchangeStr)

	// Check if exchange is enabled and get assets
	exFound := false
	for _, ex := range s.exConfig.Exchanges {
		if ex.Name == exchangeStr && ex.Enabled {
			exFound = true
			if len(ex.Assets) > 0 {
				assetSymbol = ex.Assets[0]
			}
			break
		}
	}

	if !exFound {
		s.k.Log().Info("GridTrading", "Exchange %s not enabled in exchanges.yml, using defaults", exchangeStr)
		exchange = connector.Binance
	}

	asset := s.k.Asset(assetSymbol)
	quantity := decimal.NewFromFloat(s.config.Quantity)

	// Get current price
	price, err := s.k.Market.Price(asset)
	if err != nil {
		s.k.Log().Debug("GridTrading", assetSymbol, "Failed to get price: %v", err)
		return nil, nil
	}

	// Initialize grid from config if not set
	if len(s.gridLevels) == 0 {
		lowerBound := decimal.NewFromFloat(s.config.Parameters.PriceLowerBound)
		upperBound := decimal.NewFromFloat(s.config.Parameters.PriceUpperBound)
		gridLevels := s.config.Parameters.GridLevels
		s.initializeGrid(lowerBound, upperBound, gridLevels)
		s.k.Log().Info("GridTrading", assetSymbol,
			"Initialized grid: %.2f - %.2f with %d levels (spacing: %.2f)",
			lowerBound, upperBound, gridLevels, s.spacing)
	}

	var signals []*strategy.Signal

	// Find the closest grid level below and above current price
	var buyLevel, sellLevel decimal.Decimal
	for i := 0; i < len(s.gridLevels)-1; i++ {
		if price.GreaterThanOrEqual(s.gridLevels[i]) && price.LessThan(s.gridLevels[i+1]) {
			buyLevel = s.gridLevels[i]
			sellLevel = s.gridLevels[i+1]
			break
		}
	}

	// Check if price is near a buy level (within 0.5% tolerance)
	tolerance := decimal.NewFromFloat(0.005) // 0.5%
	buyTolerance := buyLevel.Mul(tolerance)

	if price.Sub(buyLevel).Abs().LessThanOrEqual(buyTolerance) && !buyLevel.IsZero() {
		s.k.Log().Opportunity("GridTrading", assetSymbol,
			"Price %.2f near buy level %.2f - BUYING (next sell at %.2f)",
			price, buyLevel, sellLevel)

		signal := s.k.Signal(s.GetName()).
			Buy(asset, exchange, quantity).
			Build()
		signals = append(signals, signal)
	}

	// Check if price is near a sell level (within 0.5% tolerance)
	sellTolerance := sellLevel.Mul(tolerance)

	if price.Sub(sellLevel).Abs().LessThanOrEqual(sellTolerance) && !sellLevel.IsZero() {
		s.k.Log().Opportunity("GridTrading", assetSymbol,
			"Price %.2f near sell level %.2f - SELLING (next buy at %.2f)",
			price, sellLevel, buyLevel)

		signal := s.k.Signal(s.GetName()).
			Sell(asset, exchange, quantity).
			Build()
		signals = append(signals, signal)
	}

	if len(signals) == 0 {
		s.k.Log().Debug("GridTrading", assetSymbol,
			"Price %.2f between grid levels %.2f and %.2f - no action",
			price, buyLevel, sellLevel)
	}

	return signals, nil
}

// Interface implementation
func (s *gridTradingStrategy) GetName() strategy.StrategyName {
	return "Grid Trading"
}

func (s *gridTradingStrategy) GetDescription() string {
	return "Automated grid-based buy/sell orders in ranging markets"
}

func (s *gridTradingStrategy) GetRiskLevel() strategy.RiskLevel {
	return strategy.RiskLevelMedium
}

func (s *gridTradingStrategy) GetStrategyType() strategy.StrategyType {
	return strategy.StrategyTypeTechnical
}
