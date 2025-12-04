package activity

import (
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
)

// activity provides read-only access to positions, trades, and PNL
type activity struct {
	positions kronosActivity.Positions
	trades    kronosActivity.Trades
	pnl       kronosActivity.PNL
}

// NewActivity creates a new Activity instance with injected dependencies
func NewActivity(
	positions kronosActivity.Positions,
	trades kronosActivity.Trades,
	pnl kronosActivity.PNL,
) kronosActivity.Activity {
	return &activity{
		positions: positions,
		trades:    trades,
		pnl:       pnl,
	}
}

// Positions returns read-only access to position data
func (a *activity) Positions() kronosActivity.Positions {
	return a.positions
}

// Trades returns read-only access to trade data
func (a *activity) Trades() kronosActivity.Trades {
	return a.trades
}

// PNL returns access to PNL calculations
func (a *activity) PNL() kronosActivity.PNL {
	return a.pnl
}

var _ kronosActivity.Activity = (*activity)(nil)
