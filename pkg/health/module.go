package health

import (
	"go.uber.org/fx"
)

var Module = fx.Module("health",
	fx.Provide(NewConnectorHealthStore),
	fx.Provide(NewCoordinatorHealthStore),
	fx.Provide(NewHealthStore),
)
