package activity

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	storeActivity "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// trades wraps the internal trade store with read-only access
type trades struct {
	store storeActivity.Trades
}

// NewTrades creates a new read-only trades accessor
func NewTrades(store storeActivity.Trades) kronosActivity.Trades {
	return &trades{store: store}
}

// GetAllTrades retrieves all trades
func (t *trades) GetAllTrades() []connector.Trade {
	return t.store.GetAllTrades()
}

// GetTradesByExchange retrieves trades for a specific exchange
func (t *trades) GetTradesByExchange(exchange connector.ExchangeName) []connector.Trade {
	return t.store.GetTradesByExchange(exchange)
}

// GetTradesByAsset retrieves trades for a specific asset
func (t *trades) GetTradesByAsset(asset portfolio.Asset) []connector.Trade {
	return t.store.GetTradesByAsset(asset)
}

// GetTradesSince retrieves trades since a specific time
func (t *trades) GetTradesSince(since time.Time) []connector.Trade {
	return t.store.GetTradesSince(since)
}

// GetTradeByID retrieves a trade by ID
func (t *trades) GetTradeByID(tradeID string) *connector.Trade {
	return t.store.GetTradeByID(tradeID)
}

// GetTradeCount returns the total number of trades
func (t *trades) GetTradeCount() int {
	return t.store.GetTradeCount()
}

// GetTotalVolume calculates total volume for a specific asset
func (t *trades) GetTotalVolume(asset portfolio.Asset) numerical.Decimal {
	return t.store.GetTotalVolume(asset)
}

var _ kronosActivity.Trades = (*trades)(nil)
