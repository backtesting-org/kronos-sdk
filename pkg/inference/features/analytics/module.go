package analytics

import "go.uber.org/fx"

// Module provides the analytics feature extractor.
// This includes volatility and volume features from analytics.Analytics service.
// It registers the extractor into the "feature_extractors" group
// so it's automatically picked up by the feature aggregator.
var Module = fx.Module("analytics-features",
	fx.Provide(
		fx.Annotate(
			NewExtractor,
			fx.ResultTags(`group:"feature_extractors"`),
		),
	),
)
