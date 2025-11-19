package trade

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewStore,
	),
)
