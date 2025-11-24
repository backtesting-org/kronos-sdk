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

// MomentumConfig holds configuration for momentum strategy
type MomentumConfig struct {
	Exchange string `yaml:"exchange"`
	Quantity float64 `yaml:"quantity"`
	Indicators struct {
		RSI struct {
			Period     int `yaml:"period"`
			Oversold   int `yaml:"oversold"`
			Overbought int `yaml:"overbought"`
		} `yaml:"rsi"`
	} `yaml:"indicators"`
}

// MomentumStrategy implements an RSI-based momentum trading strategy
type momentumStrategy struct {
	strategy.BaseStrategy
	k        *sdk.Kronos
	config   MomentumConfig
	exConfig ExchangesConfig
}

// NewMomentum creates a new momentum strategy instance
func NewMomentum(k *sdk.Kronos) strategy.Strategy {
	// Load configuration
	var config MomentumConfig
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
	if config.Indicators.RSI.Period == 0 {
		config.Indicators.RSI.Period = 14
		config.Indicators.RSI.Oversold = 30
		config.Indicators.RSI.Overbought = 70
	}
	if config.Exchange == "" {
		config.Exchange = "binance"
	}
	if config.Quantity == 0 {
		config.Quantity = 0.1
	}

	return &momentumStrategy{k: k, config: config, exConfig: exConfig}
}

// GetSignals generates trading signals based on RSI momentum indicators
func (s *momentumStrategy) GetSignals() ([]*strategy.Signal, error) {
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
		s.k.Log().Info("Momentum", "Exchange %s not enabled in exchanges.yml, using defaults", exchangeStr)
		exchange = connector.Binance
	}

	asset := s.k.Asset(assetSymbol)
	quantity := decimal.NewFromFloat(s.config.Quantity)

	// Get current price
	price, err := s.k.Market.Price(asset)
	if err != nil {
		s.k.Log().Debug("Momentum", assetSymbol, "Failed to get price: %v", err)
		return nil, nil
	}

	// Get RSI indicator using config period
	rsi, err := s.k.Indicators.RSI(asset, s.config.Indicators.RSI.Period)
	if err != nil {
		s.k.Log().Debug("Momentum", assetSymbol, "Failed to get RSI: %v", err)
		return nil, nil
	}

	// RSI oversold threshold from config - buy signal
	oversoldThreshold := decimal.NewFromInt(int64(s.config.Indicators.RSI.Oversold))
	if rsi.LessThan(oversoldThreshold) {
		s.k.Log().Opportunity("Momentum", assetSymbol,
			"RSI oversold at %.2f (threshold: %.2f), price: %.2f - BUYING",
			rsi, oversoldThreshold, price)

		signal := s.k.Signal(s.GetName()).
			Buy(asset, exchange, quantity).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	// RSI overbought threshold from config - sell signal
	overboughtThreshold := decimal.NewFromInt(int64(s.config.Indicators.RSI.Overbought))
	if rsi.GreaterThan(overboughtThreshold) {
		s.k.Log().Opportunity("Momentum", assetSymbol,
			"RSI overbought at %.2f (threshold: %.2f), price: %.2f - SELLING",
			rsi, overboughtThreshold, price)

		signal := s.k.Signal(s.GetName()).
			Sell(asset, exchange, quantity).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	// No signal - RSI in neutral zone
	s.k.Log().Debug("Momentum", assetSymbol, "RSI neutral at %.2f, no signal", rsi)
	return nil, nil
}

// Interface implementation
func (s *momentumStrategy) GetName() strategy.StrategyName {
	return "Momentum"
}

func (s *momentumStrategy) GetDescription() string {
	return "RSI-based momentum trading with overbought/oversold signals"
}

func (s *momentumStrategy) GetRiskLevel() strategy.RiskLevel {
	return strategy.RiskLevelMedium
}

func (s *momentumStrategy) GetStrategyType() strategy.StrategyType {
	return strategy.StrategyTypeMomentum
}
