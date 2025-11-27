package features

import "go.uber.org/fx"

// Module provides the inference feature extraction system.
// This includes the aggregator and all feature extractors (market, orderbook, etc.).
// Individual feature extractor sub-modules will be added here as they are implemented.
var Module = fx.Module("inference-features",
	fx.Provide(NewAggregator),
)
