package main

import (
	sdk "github.com/backtesting-org/kronos-sdk/pkg/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/shopspring/decimal"
)

// MomentumStrategy implements an RSI-based momentum trading strategy
type momentumStrategy struct {
	strategy.BaseStrategy
	k *sdk.Kronos
}

// NewMomentum creates a new momentum strategy instance
func NewMomentum(k *sdk.Kronos) strategy.Strategy {
	return &momentumStrategy{k: k}
}

// GetSignals generates trading signals based on RSI momentum indicators
func (s *momentumStrategy) GetSignals() ([]*strategy.Signal, error) {
	btc := s.k.Asset("BTC")

	// Get current price
	price, err := s.k.Market.Price(btc)
	if err != nil {
		s.k.Log().Debug("Momentum", "BTC", "Failed to get price: %v", err)
		return nil, nil
	}

	// Get RSI indicator
	rsi, err := s.k.Indicators.RSI(btc, 14)
	if err != nil {
		s.k.Log().Debug("Momentum", "BTC", "Failed to get RSI: %v", err)
		return nil, nil
	}

	// RSI oversold threshold - buy signal
	oversoldThreshold := decimal.NewFromInt(30)
	if rsi.LessThan(oversoldThreshold) {
		s.k.Log().Opportunity("Momentum", "BTC",
			"RSI oversold at %.2f (threshold: %.2f), price: %.2f - BUYING",
			rsi, oversoldThreshold, price)

		signal := s.k.Signal(s.GetName()).
			Buy(btc, connector.Binance, decimal.NewFromFloat(0.1)).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	// RSI overbought threshold - sell signal
	overboughtThreshold := decimal.NewFromInt(70)
	if rsi.GreaterThan(overboughtThreshold) {
		s.k.Log().Opportunity("Momentum", "BTC",
			"RSI overbought at %.2f (threshold: %.2f), price: %.2f - SELLING",
			rsi, overboughtThreshold, price)

		signal := s.k.Signal(s.GetName()).
			Sell(btc, connector.Binance, decimal.NewFromFloat(0.1)).
			Build()
		return []*strategy.Signal{signal}, nil
	}

	// No signal - RSI in neutral zone
	s.k.Log().Debug("Momentum", "BTC", "RSI neutral at %.2f, no signal", rsi)
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
