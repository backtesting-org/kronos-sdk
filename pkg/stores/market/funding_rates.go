package market

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

func (ds *dataStore) UpdateFundingRate(asset portfolio.Asset, exchangeName connector.ExchangeName, rate connector.FundingRate) {
	ds.mutex.Lock()

	current := ds.getFundingRates()
	updated := make(assetFundingRates, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.FundingRateMap)
	}

	assetRates := make(marketTypes.FundingRateMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetRates[k] = v
	}
	assetRates[exchangeName] = rate
	updated[asset] = assetRates

	ds.fundingRates.Store(updated)

	ds.mutex.Unlock()

	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyFundingRates,
		Asset:    asset,
		Exchange: exchangeName,
	})
}

func (ds *dataStore) UpdateFundingRates(exchangeName connector.ExchangeName, rates map[portfolio.Asset]connector.FundingRate) {
	ds.mutex.Lock()

	current := ds.getFundingRates()
	updated := make(assetFundingRates, len(current))
	for k, v := range current {
		updated[k] = v
	}

	// Collect assets to update after releasing the lock
	assetsToUpdate := make([]portfolio.Asset, 0, len(rates))

	for asset, rate := range rates {
		if updated[asset] == nil {
			updated[asset] = make(marketTypes.FundingRateMap)
		}

		assetRates := make(marketTypes.FundingRateMap, len(updated[asset]))
		for k, v := range updated[asset] {
			assetRates[k] = v
		}
		assetRates[exchangeName] = rate
		updated[asset] = assetRates

		assetsToUpdate = append(assetsToUpdate, asset)
	}

	ds.fundingRates.Store(updated)
	ds.mutex.Unlock()

	// Update timestamps after releasing the lock to avoid deadlock
	for _, asset := range assetsToUpdate {
		ds.UpdateLastUpdated(marketTypes.UpdateKey{
			DataType: marketTypes.DataKeyFundingRates,
			Asset:    asset,
			Exchange: exchangeName,
		})
	}
}

func (ds *dataStore) GetFundingRatesForAsset(asset portfolio.Asset) marketTypes.FundingRateMap {
	current := ds.getFundingRates()
	if rates, ok := current[asset]; ok {
		return rates
	}
	return make(marketTypes.FundingRateMap)
}

func (ds *dataStore) GetFundingRate(asset portfolio.Asset, exchangeName connector.ExchangeName) *connector.FundingRate {
	current := ds.getFundingRates()
	if rates, ok := current[asset]; ok {
		if rate, ok := rates[exchangeName]; ok {
			return &rate
		}
	}
	return nil
}

func (ds *dataStore) GetAllAssetsWithFundingRates() []portfolio.Asset {
	current := ds.getFundingRates()
	assets := make([]portfolio.Asset, 0, len(current))
	for asset := range current {
		assets = append(assets, asset)
	}
	return assets
}
