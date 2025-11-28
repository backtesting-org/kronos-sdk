package technical

import "go.uber.org/fx"

// Module provides the technical indicator feature extractor.
// It registers the extractor into the "feature_extractors" group
// so it's automatically picked up by the feature aggregator.
var Module = fx.Module("technical-features",
	fx.Provide(
		fx.Annotate(
			NewExtractor,
			fx.ResultTags(`group:"feature_extractors"`),
		),
	),
)
