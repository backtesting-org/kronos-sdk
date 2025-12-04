package activity

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// positionTracker tracks open position size and average entry price for an asset
type positionTracker struct {
	asset    portfolio.Asset
	size     numerical.Decimal // Positive = long, negative = short
	avgEntry numerical.Decimal
}

// addTrade updates the position based on a trade
// Returns realized PNL if the trade closes or reduces a position
func (pt *positionTracker) addTrade(trade connector.Trade) numerical.Decimal {
	qty := trade.Quantity
	price := trade.Price

	// Determine signed quantity (positive for buys, negative for sells)
	signedQty := qty
	if trade.Side == connector.OrderSideSell {
		signedQty = qty.Neg()
	}

	// Check if this trade is reducing or reversing the position
	if pt.size.IsZero() {
		// Opening new position
		pt.size = signedQty
		pt.avgEntry = price
		return numerical.Zero()
	}

	// Same direction - adding to position
	sameDirection := (pt.size.IsPositive() && signedQty.IsPositive()) ||
		(pt.size.IsNegative() && signedQty.IsNegative())

	if sameDirection {
		// Calculate new weighted average entry
		totalValue := pt.avgEntry.Mul(pt.size.Abs()).Add(price.Mul(qty))
		pt.size = pt.size.Add(signedQty)
		if !pt.size.IsZero() {
			pt.avgEntry = totalValue.Div(pt.size.Abs())
		}
		return numerical.Zero()
	}

	// Opposite direction - closing/reducing position
	closeQty := qty
	if closeQty.GreaterThan(pt.size.Abs()) {
		closeQty = pt.size.Abs()
	}

	// Calculate realized PNL on the closed portion
	var realizedPnl numerical.Decimal
	wasPositive := pt.size.IsPositive()
	if wasPositive {
		// Closing long: PNL = (exit_price - entry_price) * quantity
		realizedPnl = price.Sub(pt.avgEntry).Mul(closeQty)
	} else {
		// Closing short: PNL = (entry_price - exit_price) * quantity
		realizedPnl = pt.avgEntry.Sub(price).Mul(closeQty)
	}

	// Calculate new position size
	newSize := pt.size.Add(signedQty)

	// Check if position flipped direction
	positionFlipped := !newSize.IsZero() &&
		((wasPositive && newSize.IsNegative()) || (!wasPositive && newSize.IsPositive()))

	if positionFlipped {
		// New position starts at the trade price
		pt.avgEntry = price
	}

	pt.size = newSize
	return realizedPnl
}

// getUnrealizedPNL calculates unrealized PNL given current market price
func (pt *positionTracker) getUnrealizedPNL(currentPrice numerical.Decimal) numerical.Decimal {
	if pt.size.IsZero() {
		return numerical.Zero()
	}

	if pt.size.IsPositive() {
		// Long position: PNL = (current_price - entry_price) * size
		return currentPrice.Sub(pt.avgEntry).Mul(pt.size)
	}

	// Short position: PNL = (entry_price - current_price) * |size|
	return pt.avgEntry.Sub(currentPrice).Mul(pt.size.Abs())
}

// pnl provides PNL calculation functionality
type pnl struct {
	positions kronosActivity.Positions
	trades    kronosActivity.Trades
	market    analytics.Market
}

// NewPNL creates a new PNL calculator
func NewPNL(positions kronosActivity.Positions, trades kronosActivity.Trades, market analytics.Market) kronosActivity.PNL {
	return &pnl{
		positions: positions,
		trades:    trades,
		market:    market,
	}
}

// calculateFromTrades processes trades and returns realized PNL and open positions
func calculateFromTrades(trades []connector.Trade) (numerical.Decimal, map[string]*positionTracker) {
	positions := make(map[string]*positionTracker)
	realizedPnl := numerical.Zero()

	for _, trade := range trades {
		symbol := trade.Symbol
		tracker, exists := positions[symbol]
		if !exists {
			tracker = &positionTracker{
				asset:    portfolio.NewAsset(symbol),
				size:     numerical.Zero(),
				avgEntry: numerical.Zero(),
			}
			positions[symbol] = tracker
		}

		pnl := tracker.addTrade(trade)
		realizedPnl = realizedPnl.Add(pnl)
	}

	return realizedPnl, positions
}

// GetRealizedPNL returns the realized PNL for a strategy (net of fees)
func (p *pnl) GetRealizedPNL(strategyName strategy.StrategyName) numerical.Decimal {
	trades := p.positions.GetTradesForStrategy(strategyName)
	realizedPnl, _ := calculateFromTrades(trades)

	// Calculate fees inline to avoid extra call
	fees := numerical.Zero()
	for _, trade := range trades {
		fees = fees.Add(trade.Fee)
	}

	return realizedPnl.Sub(fees)
}

// GetRealizedPNLByAsset returns the realized PNL for a specific asset across all strategies
func (p *pnl) GetRealizedPNLByAsset(asset portfolio.Asset) numerical.Decimal {
	trades := p.trades.GetTradesByAsset(asset)
	realizedPnl, _ := calculateFromTrades(trades)

	// Subtract fees for this asset
	fees := numerical.Zero()
	for _, trade := range trades {
		fees = fees.Add(trade.Fee)
	}

	return realizedPnl.Sub(fees)
}

// GetTotalRealizedPNL returns the total realized PNL across all strategies
func (p *pnl) GetTotalRealizedPNL() numerical.Decimal {
	allTrades := p.trades.GetAllTrades()
	realizedPnl, _ := calculateFromTrades(allTrades)

	// Calculate fees inline to avoid extra call
	fees := numerical.Zero()
	for _, trade := range allTrades {
		fees = fees.Add(trade.Fee)
	}

	return realizedPnl.Sub(fees)
}

// GetUnrealizedPNL returns the unrealized PNL for a strategy
func (p *pnl) GetUnrealizedPNL(strategyName strategy.StrategyName) (numerical.Decimal, error) {
	trades := p.positions.GetTradesForStrategy(strategyName)
	_, openPositions := calculateFromTrades(trades)

	unrealizedPnl := numerical.Zero()
	for _, tracker := range openPositions {
		if tracker.size.IsZero() {
			continue
		}

		currentPrice, err := p.market.Price(tracker.asset)
		if err != nil {
			return numerical.Zero(), err
		}

		unrealizedPnl = unrealizedPnl.Add(tracker.getUnrealizedPNL(currentPrice))
	}

	return unrealizedPnl, nil
}

// GetTotalUnrealizedPNL returns the total unrealized PNL across all strategies
func (p *pnl) GetTotalUnrealizedPNL() (numerical.Decimal, error) {
	allTrades := p.trades.GetAllTrades()
	_, openPositions := calculateFromTrades(allTrades)

	unrealizedPnl := numerical.Zero()
	for _, tracker := range openPositions {
		if tracker.size.IsZero() {
			continue
		}

		currentPrice, err := p.market.Price(tracker.asset)
		if err != nil {
			return numerical.Zero(), err
		}

		unrealizedPnl = unrealizedPnl.Add(tracker.getUnrealizedPNL(currentPrice))
	}

	return unrealizedPnl, nil
}

// GetTotalPNL returns the total PNL (realized + unrealized)
func (p *pnl) GetTotalPNL() (numerical.Decimal, error) {
	// Fetch trades once and calculate everything from cached data
	allTrades := p.trades.GetAllTrades()
	realizedPnl, openPositions := calculateFromTrades(allTrades)

	// Calculate fees inline
	fees := numerical.Zero()
	for _, trade := range allTrades {
		fees = fees.Add(trade.Fee)
	}
	realized := realizedPnl.Sub(fees)

	// Calculate unrealized PNL from open positions
	unrealized := numerical.Zero()
	for _, tracker := range openPositions {
		if tracker.size.IsZero() {
			continue
		}

		currentPrice, err := p.market.Price(tracker.asset)
		if err != nil {
			return numerical.Zero(), err
		}

		unrealized = unrealized.Add(tracker.getUnrealizedPNL(currentPrice))
	}

	return realized.Add(unrealized), nil
}

// GetTotalFees returns the total fees paid across all trades
func (p *pnl) GetTotalFees() numerical.Decimal {
	allTrades := p.trades.GetAllTrades()
	totalFees := numerical.Zero()
	for _, trade := range allTrades {
		totalFees = totalFees.Add(trade.Fee)
	}
	return totalFees
}

// GetFeesByStrategy returns the total fees paid for a strategy
func (p *pnl) GetFeesByStrategy(strategyName strategy.StrategyName) numerical.Decimal {
	trades := p.positions.GetTradesForStrategy(strategyName)
	totalFees := numerical.Zero()
	for _, trade := range trades {
		totalFees = totalFees.Add(trade.Fee)
	}
	return totalFees
}

var _ kronosActivity.PNL = (*pnl)(nil)
