package kronos

import (
	packages "github.com/backtesting-org/kronos-sdk/pkg"
	"go.uber.org/fx"
)

// Module provides the Kronos SDK with all its services wired up via fx DI.
var Module = fx.Module("kronos",
	// Include all pkg modules
	packages.Module,

	// Provide the universe provider (caches trading universe)
	fx.Provide(NewUniverseProvider),

	// Provide the main Kronos context
	fx.Provide(NewKronos),
)
