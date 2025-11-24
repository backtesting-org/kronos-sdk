package main

import (
	sdk "github.com/backtesting-org/kronos-sdk/pkg/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/shopspring/decimal"
)

// ArbitrageStrategy implements a cross-exchange arbitrage strategy
type arbitrageStrategy struct {
	strategy.BaseStrategy
	k *sdk.Kronos
}

// NewArbitrage creates a new arbitrage strategy instance
func NewArbitrage(k *sdk.Kronos) strategy.Strategy {
	return &arbitrageStrategy{k: k}
}

// GetSignals generates trading signals based on price differences between exchanges
func (s *arbitrageStrategy) GetSignals() ([]*strategy.Signal, error) {
	btc := s.k.Asset("BTC")

	// Get prices from both exchanges
	priceBinance, err := s.k.Market.PriceOnExchange(btc, connector.Binance)
	if err != nil {
		s.k.Log().Debug("Arbitrage", "BTC", "Failed to get Binance price: %v", err)
		return nil, nil
	}

	priceBybit, err := s.k.Market.PriceOnExchange(btc, connector.Bybit)
	if err != nil {
		s.k.Log().Debug("Arbitrage", "BTC", "Failed to get Bybit price: %v", err)
		return nil, nil
	}

	// Calculate spread percentage
	priceDiff := priceBinance.Sub(priceBybit).Abs()
	avgPrice := priceBinance.Add(priceBybit).Div(decimal.NewFromInt(2))
	spreadPercent := priceDiff.Div(avgPrice).Mul(decimal.NewFromInt(100))

	minSpread := decimal.NewFromFloat(0.5) // 0.5% minimum spread
	maxSpread := decimal.NewFromFloat(3.0) // 3.0% maximum spread (avoid anomalies)

	// Check if spread is profitable
	if spreadPercent.LessThan(minSpread) {
		s.k.Log().Debug("Arbitrage", "BTC",
			"Spread too low: %.2f%% (min: %.2f%%)", spreadPercent, minSpread)
		return nil, nil
	}

	if spreadPercent.GreaterThan(maxSpread) {
		s.k.Log().Debug("Arbitrage", "BTC",
			"Spread too high: %.2f%% (max: %.2f%% - possible anomaly)", spreadPercent, maxSpread)
		return nil, nil
	}

	var signals []*strategy.Signal

	// Binance price is higher - buy on Bybit, sell on Binance
	if priceBinance.GreaterThan(priceBybit) {
		s.k.Log().Opportunity("Arbitrage", "BTC",
			"Price spread detected: Binance $%.2f > Bybit $%.2f (%.2f%% spread)",
			priceBinance, priceBybit, spreadPercent)

		// Buy on cheaper exchange (Bybit)
		buySignal := s.k.Signal(s.GetName()).
			Buy(btc, connector.Bybit, decimal.NewFromFloat(0.05)).
			Build()
		signals = append(signals, buySignal)

		// Sell on expensive exchange (Binance)
		sellSignal := s.k.Signal(s.GetName()).
			Sell(btc, connector.Binance, decimal.NewFromFloat(0.05)).
			Build()
		signals = append(signals, sellSignal)

		return signals, nil
	}

	// Bybit price is higher - buy on Binance, sell on Bybit
	if priceBybit.GreaterThan(priceBinance) {
		s.k.Log().Opportunity("Arbitrage", "BTC",
			"Price spread detected: Bybit $%.2f > Binance $%.2f (%.2f%% spread)",
			priceBybit, priceBinance, spreadPercent)

		// Buy on cheaper exchange (Binance)
		buySignal := s.k.Signal(s.GetName()).
			Buy(btc, connector.Binance, decimal.NewFromFloat(0.05)).
			Build()
		signals = append(signals, buySignal)

		// Sell on expensive exchange (Bybit)
		sellSignal := s.k.Signal(s.GetName()).
			Sell(btc, connector.Bybit, decimal.NewFromFloat(0.05)).
			Build()
		signals = append(signals, sellSignal)

		return signals, nil
	}

	return nil, nil
}

// Interface implementation
func (s *arbitrageStrategy) GetName() strategy.StrategyName {
	return "Arbitrage"
}

func (s *arbitrageStrategy) GetDescription() string {
	return "Cross-exchange price difference exploitation"
}

func (s *arbitrageStrategy) GetRiskLevel() strategy.RiskLevel {
	return strategy.RiskLevelMedium
}

func (s *arbitrageStrategy) GetStrategyType() strategy.StrategyType {
	return strategy.StrategyTypeArbitrage
}
