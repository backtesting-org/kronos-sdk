package price

import (
	"go.uber.org/fx"
)

var Module = fx.Module("price-features",
	fx.Provide(
		fx.Annotate(
			NewExtractor,
			fx.ResultTags(`group:"feature_extractors"`),
		),
	),
)
