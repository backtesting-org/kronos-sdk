package activity

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewPositions,
		NewTrades,
		NewPNL,
		NewActivity,
	),
)
