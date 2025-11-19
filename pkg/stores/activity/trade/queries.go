package trade

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/shopspring/decimal"
)

// GetAllTrades retrieves all trades
func (ds *dataStore) GetAllTrades() []connector.Trade {
	return ds.getTrades()
}

// GetTradesByExchange retrieves trades for a specific exchange
func (ds *dataStore) GetTradesByExchange(exchange connector.ExchangeName) []connector.Trade {
	trades := ds.getTrades()
	result := make([]connector.Trade, 0)

	for _, trade := range trades {
		if trade.Exchange == exchange {
			result = append(result, trade)
		}
	}

	return result
}

// GetTradesByAsset retrieves trades for a specific asset
func (ds *dataStore) GetTradesByAsset(asset portfolio.Asset) []connector.Trade {
	trades := ds.getTrades()
	result := make([]connector.Trade, 0)

	for _, trade := range trades {
		if trade.Symbol == asset.Symbol() {
			result = append(result, trade)
		}
	}

	return result
}

// GetTradesByExchangeAndAsset retrieves trades for a specific exchange and asset
func (ds *dataStore) GetTradesByExchangeAndAsset(exchange connector.ExchangeName, asset portfolio.Asset) []connector.Trade {
	trades := ds.getTrades()
	result := make([]connector.Trade, 0)

	for _, trade := range trades {
		if trade.Exchange == exchange && trade.Symbol == asset.Symbol() {
			result = append(result, trade)
		}
	}

	return result
}

// GetTradesSince retrieves trades since a specific time
func (ds *dataStore) GetTradesSince(since time.Time) []connector.Trade {
	trades := ds.getTrades()
	result := make([]connector.Trade, 0)

	for _, trade := range trades {
		if trade.Timestamp.After(since) || trade.Timestamp.Equal(since) {
			result = append(result, trade)
		}
	}

	return result
}

// GetTradeByID retrieves a trade by ID
func (ds *dataStore) GetTradeByID(tradeID string) *connector.Trade {
	tradeMap := ds.getTradeMap()

	if trade, exists := tradeMap[tradeID]; exists {
		return &trade
	}

	return nil
}

// TradeExists checks if a trade exists by ID
func (ds *dataStore) TradeExists(tradeID string) bool {
	tradeMap := ds.getTradeMap()
	_, exists := tradeMap[tradeID]
	return exists
}

// GetTradeCount returns the total number of trades
func (ds *dataStore) GetTradeCount() int {
	trades := ds.getTrades()
	return len(trades)
}

// GetTotalVolume calculates total volume for a specific asset
func (ds *dataStore) GetTotalVolume(asset portfolio.Asset) decimal.Decimal {
	trades := ds.getTrades()
	totalVolume := decimal.Zero

	for _, trade := range trades {
		if trade.Symbol == asset.Symbol() {
			totalVolume = totalVolume.Add(trade.Quantity)
		}
	}

	return totalVolume
}
