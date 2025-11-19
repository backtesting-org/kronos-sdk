package market

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/stores/market"
)

func (ds *dataStore) UpdateHistoricalFundingRates(asset portfolio.Asset, exchangeName connector.ExchangeName, rates []connector.HistoricalFundingRate) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getHistoricalFunding()
	updated := make(assetHistoricalFunding, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.HistoricalFundingMap)
	}

	assetRates := make(marketTypes.HistoricalFundingMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetRates[k] = v
	}
	assetRates[exchangeName] = rates
	updated[asset] = assetRates

	ds.historicalFundingRates.Store(updated)
	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyHistoricalFunding,
		Asset:    asset,
		Exchange: exchangeName,
	})
	ds.notifyOrchestrator()
}

func (ds *dataStore) GetHistoricalFundingRatesForAsset(asset portfolio.Asset) marketTypes.HistoricalFundingMap {
	current := ds.getHistoricalFunding()
	if rates, ok := current[asset]; ok {
		return rates
	}
	return make(marketTypes.HistoricalFundingMap)
}
