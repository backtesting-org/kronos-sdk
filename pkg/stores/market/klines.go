package market

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

func (ds *dataStore) UpdateKline(asset portfolio.Asset, exchangeName connector.ExchangeName, kline connector.Kline) {
	ds.mutex.Lock()
	
	current := ds.getKlines()
	updated := make(assetKlines, len(current))
	for k, v := range current {
		updated[k] = v
	}

	if updated[asset] == nil {
		updated[asset] = make(marketTypes.KlineMap)
	}

	assetKlineMap := make(marketTypes.KlineMap, len(updated[asset]))
	for k, v := range updated[asset] {
		assetKlineMap[k] = v
	}

	if assetKlineMap[exchangeName] == nil {
		assetKlineMap[exchangeName] = make(map[string][]connector.Kline)
	}

	exchangeKlines := make(map[string][]connector.Kline, len(assetKlineMap[exchangeName]))
	for k, v := range assetKlineMap[exchangeName] {
		exchangeKlines[k] = v
	}

	// Add or update kline for this interval
	interval := kline.Interval
	existingKlines := exchangeKlines[interval]

	// Check if kline already exists (by open time)
	found := false
	for i, existing := range existingKlines {
		if existing.OpenTime.Equal(kline.OpenTime) {
			existingKlines[i] = kline
			found = true
			break
		}
	}

	if !found {
		existingKlines = append(existingKlines, kline)
	}

	exchangeKlines[interval] = existingKlines
	assetKlineMap[exchangeName] = exchangeKlines
	updated[asset] = assetKlineMap

	ds.klines.Store(updated)
	ds.mutex.Unlock()

	ds.UpdateLastUpdated(marketTypes.UpdateKey{
		DataType: marketTypes.DataKeyKlines,
		Asset:    asset,
		Exchange: exchangeName,
	})
}

func (ds *dataStore) GetKlines(asset portfolio.Asset, exchangeName connector.ExchangeName, interval string, limit int) []connector.Kline {
	current := ds.getKlines()
	if klineMap, ok := current[asset]; ok {
		if exchangeKlines, ok := klineMap[exchangeName]; ok {
			if klines, ok := exchangeKlines[interval]; ok {
				if limit > 0 && len(klines) > limit {
					return klines[len(klines)-limit:]
				}
				return klines
			}
		}
	}
	return []connector.Kline{}
}

func (ds *dataStore) GetKlinesSince(asset portfolio.Asset, exchangeName connector.ExchangeName, interval string, since time.Time) []connector.Kline {
	current := ds.getKlines()
	if klineMap, ok := current[asset]; ok {
		if exchangeKlines, ok := klineMap[exchangeName]; ok {
			if klines, ok := exchangeKlines[interval]; ok {
				result := make([]connector.Kline, 0)
				for _, kline := range klines {
					if kline.OpenTime.After(since) || kline.OpenTime.Equal(since) {
						result = append(result, kline)
					}
				}
				return result
			}
		}
	}
	return []connector.Kline{}
}
