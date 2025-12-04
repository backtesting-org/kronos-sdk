package market

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

func (ds *dataStore) UpdateAssetPrice(asset portfolio.Asset, exchangeName connector.ExchangeName, price connector.Price) {
	ds.mutex.Lock()

	current := ds.getPrices()
	updated := make(assetPrices, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.PriceMap)
	}

	assetPriceMap := make(marketTypes.PriceMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetPriceMap[k] = v
	}
	assetPriceMap[exchangeName] = price
	updated[asset] = assetPriceMap

	ds.prices.Store(updated)

	ds.mutex.Unlock()

	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyAssetPrice,
		Asset:    asset,
		Exchange: exchangeName,
	})
}

func (ds *dataStore) UpdateAssetPrices(asset portfolio.Asset, prices map[connector.ExchangeName]connector.Price) {
	ds.mutex.Lock()

	current := ds.getPrices()
	updated := make(assetPrices, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.PriceMap)
	}

	assetPriceMap := make(marketTypes.PriceMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetPriceMap[k] = v
	}

	// Collect exchanges to update after releasing the lock
	exchangesToUpdate := make([]connector.ExchangeName, 0, len(prices))

	for exchangeName, price := range prices {
		assetPriceMap[exchangeName] = price
		exchangesToUpdate = append(exchangesToUpdate, exchangeName)
	}

	updated[asset] = assetPriceMap
	ds.prices.Store(updated)
	ds.mutex.Unlock()

	// Update timestamps after releasing the lock to avoid deadlock
	for _, exchangeName := range exchangesToUpdate {
		ds.UpdateLastUpdated(marketTypes.UpdateKey{
			DataType: marketTypes.DataKeyAssetPrice,
			Asset:    asset,
			Exchange: exchangeName,
		})
	}
}

func (ds *dataStore) GetAssetPrice(asset portfolio.Asset, exchangeName connector.ExchangeName) *connector.Price {
	current := ds.getPrices()
	if priceMap, ok := current[asset]; ok {
		if price, ok := priceMap[exchangeName]; ok {
			return &price
		}
	}
	return nil
}

func (ds *dataStore) GetAssetPrices(asset portfolio.Asset) marketTypes.PriceMap {
	current := ds.getPrices()
	if prices, ok := current[asset]; ok {
		return prices
	}
	return make(marketTypes.PriceMap)
}
