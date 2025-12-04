package market

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

func (ds *dataStore) UpdateOrderBook(asset portfolio.Asset, exchangeName connector.ExchangeName, orderBookType connector.Instrument, orderBook connector.OrderBook) {
	ds.mutex.Lock()

	current := ds.getOrderBooks()
	updated := make(assetOrderBooks, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.OrderBookMap)
	}

	assetBooks := make(marketTypes.OrderBookMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetBooks[k] = v
	}

	if assetBooks[exchangeName] == nil {
		assetBooks[exchangeName] = make(map[connector.Instrument]*connector.OrderBook)
	}

	exchangeBooks := make(map[connector.Instrument]*connector.OrderBook, len(assetBooks[exchangeName]))
	for k, v := range assetBooks[exchangeName] {
		exchangeBooks[k] = v
	}
	exchangeBooks[orderBookType] = &orderBook
	assetBooks[exchangeName] = exchangeBooks
	updated[asset] = assetBooks

	ds.orderBooks.Store(updated)

	ds.mutex.Unlock()

	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyOrderBooks,
		Asset:    asset,
		Exchange: exchangeName,
	})
}

func (ds *dataStore) GetOrderBooks(asset portfolio.Asset) marketTypes.OrderBookMap {
	current := ds.getOrderBooks()
	if books, ok := current[asset]; ok {
		return books
	}
	return make(marketTypes.OrderBookMap)
}

func (ds *dataStore) GetOrderBook(asset portfolio.Asset, exchangeName connector.ExchangeName, orderBookType connector.Instrument) *connector.OrderBook {
	current := ds.getOrderBooks()
	if books, ok := current[asset]; ok {
		if exchangeBooks, ok := books[exchangeName]; ok {
			if book, ok := exchangeBooks[orderBookType]; ok {
				return book
			}
		}
	}
	return nil
}

func (ds *dataStore) GetAllAssetsWithOrderBooks() []portfolio.Asset {
	current := ds.getOrderBooks()
	assets := make([]portfolio.Asset, 0, len(current))
	for asset := range current {
		assets = append(assets, asset)
	}
	return assets
}
