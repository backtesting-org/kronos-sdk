package market

import (
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

func NewStore(timeProvider temporal.TimeProvider) marketTypes.MarketData {
	ds := &dataStore{
		timeProvider: timeProvider,
	}
	ds.fundingRates.Store(make(assetFundingRates))
	ds.historicalFundingRates.Store(make(assetHistoricalFunding))
	ds.orderBooks.Store(make(assetOrderBooks))
	ds.prices.Store(make(assetPrices))
	ds.klines.Store(make(assetKlines))
	ds.lastUpdated.Store(make(marketTypes.LastUpdatedMap))
	return ds
}

func (ds *dataStore) Clear() {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.fundingRates.Store(make(assetFundingRates))
	ds.historicalFundingRates.Store(make(assetHistoricalFunding))
	ds.orderBooks.Store(make(assetOrderBooks))
	ds.prices.Store(make(assetPrices))
	ds.klines.Store(make(assetKlines))
	ds.lastUpdated.Store(make(marketTypes.LastUpdatedMap))
}

// Helper methods to get typed data from atomic.Value
func (ds *dataStore) getFundingRates() assetFundingRates {
	if v := ds.fundingRates.Load(); v != nil {
		return v.(assetFundingRates)
	}
	return make(assetFundingRates)
}

func (ds *dataStore) getHistoricalFunding() assetHistoricalFunding {
	if v := ds.historicalFundingRates.Load(); v != nil {
		return v.(assetHistoricalFunding)
	}
	return make(assetHistoricalFunding)
}

func (ds *dataStore) getOrderBooks() assetOrderBooks {
	if v := ds.orderBooks.Load(); v != nil {
		return v.(assetOrderBooks)
	}
	return make(assetOrderBooks)
}

func (ds *dataStore) getPrices() assetPrices {
	if v := ds.prices.Load(); v != nil {
		return v.(assetPrices)
	}
	return make(assetPrices)
}

func (ds *dataStore) getKlines() assetKlines {
	if v := ds.klines.Load(); v != nil {
		return v.(assetKlines)
	}
	return make(assetKlines)
}

func (ds *dataStore) getLastUpdated() marketTypes.LastUpdatedMap {
	if v := ds.lastUpdated.Load(); v != nil {
		return v.(marketTypes.LastUpdatedMap)
	}
	return make(marketTypes.LastUpdatedMap)
}

var _ marketTypes.MarketData = (*dataStore)(nil)
