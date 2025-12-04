package activity

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// Trades provides read-only access to trade history
type Trades interface {
	// Query trades
	GetAllTrades() []connector.Trade
	GetTradesByExchange(exchange connector.ExchangeName) []connector.Trade
	GetTradesByAsset(asset portfolio.Asset) []connector.Trade
	GetTradesSince(since time.Time) []connector.Trade
	GetTradeByID(tradeID string) *connector.Trade

	// Stats
	GetTradeCount() int
	GetTotalVolume(asset portfolio.Asset) numerical.Decimal
}
