package trade

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
)

// AddTrade adds a single trade to the store
func (ds *dataStore) AddTrade(trade connector.Trade) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	// Check if trade already exists
	tradeMap := ds.getTradeMap()
	if _, exists := tradeMap[trade.ID]; exists {
		return // Skip duplicates
	}

	// Update trade list
	currentTrades := ds.getTrades()
	updatedTrades := make([]connector.Trade, len(currentTrades)+1)
	copy(updatedTrades, currentTrades)
	updatedTrades[len(currentTrades)] = trade
	ds.trades.Store(updatedTrades)

	// Update trade map
	updatedMap := make(TradeMap, len(tradeMap)+1)
	for k, v := range tradeMap {
		updatedMap[k] = v
	}
	updatedMap[trade.ID] = trade
	ds.byID.Store(updatedMap)
}

// AddTrades adds multiple trades to the store
func (ds *dataStore) AddTrades(trades []connector.Trade) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	tradeMap := ds.getTradeMap()
	currentTrades := ds.getTrades()

	// Filter out duplicates
	newTrades := make([]connector.Trade, 0)
	for _, trade := range trades {
		if _, exists := tradeMap[trade.ID]; !exists {
			newTrades = append(newTrades, trade)
		}
	}

	if len(newTrades) == 0 {
		return // No new trades
	}

	// Update trade list
	updatedTrades := make([]connector.Trade, len(currentTrades)+len(newTrades))
	copy(updatedTrades, currentTrades)
	copy(updatedTrades[len(currentTrades):], newTrades)
	ds.trades.Store(updatedTrades)

	// Update trade map
	updatedMap := make(TradeMap, len(tradeMap)+len(newTrades))
	for k, v := range tradeMap {
		updatedMap[k] = v
	}
	for _, trade := range newTrades {
		updatedMap[trade.ID] = trade
	}
	ds.byID.Store(updatedMap)
}

// getTrades is a helper to get the current trade slice
func (ds *dataStore) getTrades() []connector.Trade {
	if v := ds.trades.Load(); v != nil {
		return v.([]connector.Trade)
	}
	return []connector.Trade{}
}

// getTradeMap is a helper to get the current trade map
func (ds *dataStore) getTradeMap() TradeMap {
	if v := ds.byID.Load(); v != nil {
		return v.(TradeMap)
	}
	return make(TradeMap)
}
