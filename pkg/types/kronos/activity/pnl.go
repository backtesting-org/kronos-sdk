package activity

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// PNL provides profit and loss calculations
type PNL interface {
	// Realized PNL (from executed trades)
	GetRealizedPNL(strategy strategy.StrategyName) numerical.Decimal
	GetRealizedPNLByAsset(asset portfolio.Asset) numerical.Decimal
	GetTotalRealizedPNL() numerical.Decimal

	// Unrealized PNL (requires current market prices)
	GetUnrealizedPNL(strategy strategy.StrategyName) (numerical.Decimal, error)
	GetTotalUnrealizedPNL() (numerical.Decimal, error)

	// Combined
	GetTotalPNL() (numerical.Decimal, error)

	// Fee tracking
	GetTotalFees() numerical.Decimal
	GetFeesByStrategy(strategy strategy.StrategyName) numerical.Decimal
}
