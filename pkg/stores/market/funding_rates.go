package market

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/stores/market"
)

func (ds *dataStore) UpdateFundingRate(asset portfolio.Asset, exchangeName connector.ExchangeName, rate connector.FundingRate) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

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
	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyFundingRates,
		Asset:    asset,
		Exchange: exchangeName,
	})
	ds.notifyOrchestrator()
}

func (ds *dataStore) UpdateFundingRates(exchangeName connector.ExchangeName, rates map[portfolio.Asset]connector.FundingRate) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getFundingRates()
	updated := make(assetFundingRates, len(current))
	for k, v := range current {
		updated[k] = v
	}

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

		ds.UpdateLastUpdated(marketTypes.UpdateKey{
			DataType: marketTypes.DataKeyFundingRates,
			Asset:    asset,
			Exchange: exchangeName,
		})
	}

	ds.fundingRates.Store(updated)
	ds.notifyOrchestrator()
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
