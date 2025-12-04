package market

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// MarketData Defines the interface for asset market data operations
type MarketData interface {
	UpdateFundingRate(asset portfolio.Asset, exchangeName connector.ExchangeName, rate connector.FundingRate)
	UpdateFundingRates(exchangeName connector.ExchangeName, rates map[portfolio.Asset]connector.FundingRate)
	GetFundingRatesForAsset(asset portfolio.Asset) FundingRateMap
	GetFundingRate(asset portfolio.Asset, exchangeName connector.ExchangeName) *connector.FundingRate
	GetAllAssetsWithFundingRates() []portfolio.Asset

	UpdateHistoricalFundingRates(asset portfolio.Asset, exchangeName connector.ExchangeName, rates []connector.HistoricalFundingRate)
	GetHistoricalFundingRatesForAsset(asset portfolio.Asset) HistoricalFundingMap

	UpdateOrderBook(asset portfolio.Asset, exchangeName connector.ExchangeName, orderBookType connector.Instrument, orderBook connector.OrderBook)
	GetOrderBooks(asset portfolio.Asset) OrderBookMap
	GetOrderBook(asset portfolio.Asset, exchangeName connector.ExchangeName, orderBookType connector.Instrument) *connector.OrderBook
	GetAllAssetsWithOrderBooks() []portfolio.Asset

	UpdateAssetPrice(asset portfolio.Asset, exchangeName connector.ExchangeName, price connector.Price)
	UpdateAssetPrices(asset portfolio.Asset, prices map[connector.ExchangeName]connector.Price)
	GetAssetPrice(asset portfolio.Asset, exchangeName connector.ExchangeName) *connector.Price
	GetAssetPrices(asset portfolio.Asset) PriceMap

	UpdateKline(asset portfolio.Asset, exchangeName connector.ExchangeName, kline connector.Kline)
	GetKlines(asset portfolio.Asset, exchangeName connector.ExchangeName, interval string, limit int) []connector.Kline
	GetKlinesSince(asset portfolio.Asset, exchangeName connector.ExchangeName, interval string, since time.Time) []connector.Kline

	GetLastUpdated() LastUpdatedMap
	UpdateLastUpdated(key UpdateKey)

	// Clear all data for simulation restart
	Clear()
}

type DataKey string

const (
	DataKeyOrderBooks        DataKey = "order_books"
	DataKeyFundingRates      DataKey = "funding_rates"
	DataKeyHistoricalFunding DataKey = "historical_funding"
	DataKeyAssetPrice        DataKey = "asset_price"
	DataKeyKlines            DataKey = "klines"
)

type UpdateKey struct {
	DataType DataKey
	Asset    portfolio.Asset
	Exchange connector.ExchangeName
}

type LastUpdatedMap map[UpdateKey]time.Time
type OrderBookMap map[connector.ExchangeName]map[connector.Instrument]*connector.OrderBook
type FundingRateMap map[connector.ExchangeName]connector.FundingRate
type HistoricalFundingMap map[connector.ExchangeName][]connector.HistoricalFundingRate
type PriceMap map[connector.ExchangeName]connector.Price
type KlineMap map[connector.ExchangeName]map[string][]connector.Kline
