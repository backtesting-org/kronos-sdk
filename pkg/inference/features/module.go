package features

import (
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features/market"
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features/orderbook"
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features/price"
	"github.com/backtesting-org/kronos-sdk/pkg/inference/features/technical"
	"go.uber.org/fx"
)

// Module provides the inference feature extraction system.
// This includes the aggregator and all feature extractors.
// Individual feature extractor sub-modules will be added here as they are implemented.
var Module = fx.Module("inference-features",
	fx.Provide(
		fx.Annotate(
			NewAggregator,
			fx.As(new(FeatureAggregator)),
		),
	),
	// Feature extractor sub-modules:
	market.Module,     // Price data, funding rates
	technical.Module,  // Technical indicators (RSI, MACD, BB, etc.)
	analytics.Module,  // Analytics features (volatility, volume, time)
	price.Module,      // Price metrics (returns, high/low, VWAP)
	orderbook.Module,  // Orderbook features (spread, depth, imbalance)
)
