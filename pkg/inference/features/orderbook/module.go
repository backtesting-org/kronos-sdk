package orderbook

import (
	"go.uber.org/fx"
)

var Module = fx.Module("inference-features-orderbook",
	fx.Provide(
		fx.Annotate(
			NewExtractor,
			fx.ResultTags(`group:"feature_extractors"`),
		),
	),
)
