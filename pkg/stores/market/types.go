package market

import (
	"sync"
	"sync/atomic"

	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/stores/market"
)

// Data types for atomic storage
type assetFundingRates map[portfolio.Asset]marketTypes.FundingRateMap
type assetHistoricalFunding map[portfolio.Asset]marketTypes.HistoricalFundingMap
type assetOrderBooks map[portfolio.Asset]marketTypes.OrderBookMap
type assetPrices map[portfolio.Asset]marketTypes.PriceMap
type assetKlines map[portfolio.Asset]marketTypes.KlineMap

type dataStore struct {
	fundingRates           atomic.Value // assetFundingRates
	historicalFundingRates atomic.Value // assetHistoricalFunding
	orderBooks             atomic.Value // assetOrderBooks
	prices                 atomic.Value // assetPrices
	klines                 atomic.Value // assetKlines
	lastUpdated            atomic.Value // marketTypes.LastUpdatedMap

	mutex                sync.RWMutex
	orchestratorNotifier func()
}
