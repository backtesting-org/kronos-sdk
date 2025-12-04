package activity

// Activity provides read-only access to positions, trades, and PNL
type Activity interface {
	Positions() Positions
	Trades() Trades
	PNL() PNL
}
