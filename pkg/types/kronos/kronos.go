package kronos

import (
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// Kronos is the main SDK context provided to strategies for accessing market data,
// indicators, analytics, and trading functionality. It is injected into strategy
// implementations and provides read-only access to all framework services.
type Kronos interface {
	// Universe returns the tradeable assets, instruments, and exchanges.
	// Provides strategies with the complete trading universe available.
	// Example: u := k.Universe(); for asset, instruments := range u.Assets { ... }
	Universe() Universe

	// Indicators returns the indicators service for technical analysis.
	// Provides methods like RSI, SMA, EMA, MACD, etc.
	Indicators() analytics.Indicators

	// Analytics returns the analytics service for market analysis.
	// Provides methods for volatility, trend analysis, and volume analysis.
	Analytics() analytics.Analytics

	// Market returns the market data service for accessing live and historical prices.
	// Provides methods to query market data across registered exchanges.
	Market() analytics.Market

	// Log returns the trading logger for strategy-specific logging.
	// Use for recording trading decisions and strategy events.
	Log() logging.TradingLogger

	// Store returns the underlying market data store for advanced use cases.
	// Most users should use the Indicators, Analytics, and Market services instead.
	Store() market.MarketData

	// Activity returns read-only access to positions, trades, and PNL data.
	// Provides methods to query strategy executions, orders, and trade history.
	// Example: k.Activity().Positions().GetStrategyExecution(strategyName)
	Activity() activity.Activity

	// Asset creates a new portfolio.Asset from a symbol string.
	// Convenience method to avoid importing portfolio package in strategies.
	// Example: btc := k.Asset("BTC")
	Asset(symbol string) portfolio.Asset

	// Signal creates a new signal builder for constructing trading signals.
	// Returns a fluent API for building buy/sell signals with price targets.
	// Example: k.Signal(strategyName).Buy(asset, exchange, qty).Build()
	Signal(strategyName strategy.StrategyName) strategy.SignalBuilder

	// Features returns the ML feature aggregator for extracting market features.
	// Provides access to 41+ features including market data, orderbook, technical indicators,
	// volatility, volume, price metrics, and time-based features.
	// Example: featureMap, err := k.Features().Extract(asset)
	Features() features.FeatureAggregator
}

// Universe holds the tradeable assets and exchanges available to the strategy.
type Universe struct {
	// Exchanges are the ready/initialized exchanges available for trading
	Exchanges []connector.Exchange

	// Assets maps each tradeable asset to its supported instruments on registered exchanges
	// Example: {BTC: [Spot, Perpetual], ETH: [Spot]}
	Assets map[portfolio.Asset][]connector.Instrument
}
