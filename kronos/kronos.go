package kronos

import (
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	kronosTypes "github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// Kronos is the base context object for strategy GetSignals methods.
// It provides read-only access to market data, indicators, and analytics.
type kronos struct {
	store            market.MarketData
	tradingLogger    logging.TradingLogger
	universeProvider UniverseProvider
	indicators       analytics.Indicators
	analytics        analytics.Analytics
	market           analytics.Market
	signal           strategy.SignalFactory
	featureAggregator features.FeatureAggregator
	activity         activity.Activity
}

// NewKronos creates a new Kronos context with injected services.
// This is injected via fx DI into strategies.
func NewKronos(
	store market.MarketData,
	tradingLogger logging.TradingLogger,
	universeProvider UniverseProvider,
	indicators analytics.Indicators,
	analyticsService analytics.Analytics,
	marketService analytics.Market,
	signal strategy.SignalFactory,
	featureAggregator features.FeatureAggregator,
	activityService activity.Activity,
) kronosTypes.Kronos {
	return &kronos{
		store:            store,
		tradingLogger:    tradingLogger,
		universeProvider: universeProvider,
		indicators:       indicators,
		market:           marketService,
		analytics:        analyticsService,
		signal:           signal,
		featureAggregator: featureAggregator,
		activity:         activityService,
	}
}

func (k *kronos) Indicators() analytics.Indicators {
	return k.indicators
}

func (k *kronos) Analytics() analytics.Analytics {
	return k.analytics
}

func (k *kronos) Market() analytics.Market {
	return k.market
}

func (k *kronos) Activity() activity.Activity {
	return k.activity
}

// Log returns the trading logger for strategy logging.
// Usage: k.Log().Opportunity("MyStrategy", "BTC", "Found signal")
func (k *kronos) Log() logging.TradingLogger {
	return k.tradingLogger
}

// Store returns the underlying store for advanced use cases.
// Most users should use the service methods instead.
func (k *kronos) Store() market.MarketData {
	return k.store
}

// Asset creates a new portfolio.Asset from a symbol string.
// This is a convenience method to avoid importing portfolio everywhere.
// Usage: btc := k.Asset("BTC")
func (k *kronos) Asset(symbol string) portfolio.Asset {
	return portfolio.NewAsset(symbol)
}

// Signal creates a new signal builder for constructing trading signals.
// Usage: k.Signal(strategyName).Buy(asset, exchange, qty).SellShort(asset, exchange, qty).Build()
func (k *kronos) Signal(strategyName strategy.StrategyName) strategy.SignalBuilder {
	return k.signal.New(strategyName)
}

// Universe returns the tradeable assets, instruments, and ready exchanges.
// Provides the complete trading universe available to the strategy.
// Data is cached since it does not change after initialization.
func (k *kronos) Universe() kronosTypes.Universe {
	return k.universeProvider.Universe()
}

// Features returns the ML feature aggregator for extracting market features.
// Provides access to 41+ features including market data, orderbook, technical indicators,
// volatility, volume, price metrics, and time-based features.
// Usage: featureMap, err := k.Features().Extract(asset)
func (k *kronos) Features() features.FeatureAggregator {
	return k.featureAggregator
}
