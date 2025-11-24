package main

import (
	sdk "github.com/backtesting-org/kronos-sdk/pkg/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/shopspring/decimal"
)

// GridTradingStrategy implements an automated grid trading strategy
type gridTradingStrategy struct {
	strategy.BaseStrategy
	k          *sdk.Kronos
	gridLevels []decimal.Decimal
	spacing    decimal.Decimal
}

// NewGridTrading creates a new grid trading strategy instance
func NewGridTrading(k *sdk.Kronos) strategy.Strategy {
	return &gridTradingStrategy{
		k:          k,
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
	btc := s.k.Asset("BTC")

	// Get current price
	price, err := s.k.Market.Price(btc)
	if err != nil {
		s.k.Log().Debug("GridTrading", "BTC", "Failed to get price: %v", err)
		return nil, nil
	}

	// Initialize grid if not set (example bounds - should come from config)
	if len(s.gridLevels) == 0 {
		lowerBound := decimal.NewFromInt(40000)
		upperBound := decimal.NewFromInt(50000)
		gridLevels := 10
		s.initializeGrid(lowerBound, upperBound, gridLevels)
		s.k.Log().Info("GridTrading", "BTC",
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
		s.k.Log().Opportunity("GridTrading", "BTC",
			"Price %.2f near buy level %.2f - BUYING (next sell at %.2f)",
			price, buyLevel, sellLevel)

		signal := s.k.Signal(s.GetName()).
			Buy(btc, connector.Binance, decimal.NewFromFloat(0.01)).
			Build()
		signals = append(signals, signal)
	}

	// Check if price is near a sell level (within 0.5% tolerance)
	sellTolerance := sellLevel.Mul(tolerance)

	if price.Sub(sellLevel).Abs().LessThanOrEqual(sellTolerance) && !sellLevel.IsZero() {
		s.k.Log().Opportunity("GridTrading", "BTC",
			"Price %.2f near sell level %.2f - SELLING (next buy at %.2f)",
			price, sellLevel, buyLevel)

		signal := s.k.Signal(s.GetName()).
			Sell(btc, connector.Binance, decimal.NewFromFloat(0.01)).
			Build()
		signals = append(signals, signal)
	}

	if len(signals) == 0 {
		s.k.Log().Debug("GridTrading", "BTC",
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
