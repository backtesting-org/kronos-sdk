package price

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

// Feature name constants (must match pkg/inference/features/types.go)
const (
	featureReturn1m = "return_1m"
	featureReturn5m = "return_5m"
	featureReturn1h = "return_1h"
	featureHigh1h   = "high_1h"
	featureLow1h    = "low_1h"
	featureVWAP1h   = "vwap_1h"
)

// Extractor computes price-based features (returns, highs, lows, VWAP).
// It uses historical kline (candle) data to calculate metrics over different time windows.
type Extractor struct {
	marketData   market.MarketData
	timeProvider temporal.TimeProvider
}

// NewExtractor creates a new price feature extractor.
func NewExtractor(marketData market.MarketData, timeProvider temporal.TimeProvider) *Extractor {
	return &Extractor{
		marketData:   marketData,
		timeProvider: timeProvider,
	}
}

// Extract computes price features and adds them to the feature map.
// Currently supports: returns (1m, 5m, 1h), high/low (1h), and VWAP (1h).
func (e *Extractor) Extract(asset portfolio.Asset, featureMap map[string]float64) error {
	now := e.timeProvider.Now()

	// Get current price (use first available exchange)
	prices := e.marketData.GetAssetPrices(asset)
	if len(prices) == 0 {
		return nil // No price data available
	}

	// Use first available exchange
	var exchangeName connector.ExchangeName
	var currentPrice numerical.Decimal
	for exchange, price := range prices {
		exchangeName = exchange
		currentPrice = price.Price
		break
	}

	// Calculate returns using 1-minute klines
	e.calculateReturn(asset, exchangeName, currentPrice, now.Add(-1*time.Minute), featureReturn1m, featureMap)
	e.calculateReturn(asset, exchangeName, currentPrice, now.Add(-5*time.Minute), featureReturn5m, featureMap)
	e.calculateReturn(asset, exchangeName, currentPrice, now.Add(-1*time.Hour), featureReturn1h, featureMap)

	// Calculate 1-hour high/low and VWAP using klines
	e.calculateHighLowVWAP(asset, exchangeName, now.Add(-1*time.Hour), now, featureMap)

	return nil
}

// calculateReturn computes percentage return between a past time and current price
func (e *Extractor) calculateReturn(asset portfolio.Asset, exchange connector.ExchangeName, currentPrice numerical.Decimal, pastTime time.Time, featureName string, featureMap map[string]float64) {
	// Get klines since pastTime
	klines := e.marketData.GetKlinesSince(asset, exchange, "1m", pastTime)
	if len(klines) == 0 {
		return
	}

	// Use the first kline's open price as the historical price
	pastPrice := klines[0].Open
	if pastPrice.IsZero() {
		return
	}

	// Calculate return percentage: (current - past) / past * 100
	returnPct := currentPrice.Sub(pastPrice).Div(pastPrice).Mul(numerical.NewFromInt(100))
	featureMap[featureName], _ = returnPct.Float64()
}

// calculateHighLowVWAP computes high, low, and VWAP over a time window
func (e *Extractor) calculateHighLowVWAP(asset portfolio.Asset, exchange connector.ExchangeName, start, end time.Time, featureMap map[string]float64) {
	// Get all klines in the time window
	klines := e.marketData.GetKlinesSince(asset, exchange, "1m", start)
	if len(klines) == 0 {
		return
	}

	// Filter klines to only include those within the time window
	var filtered []connector.Kline
	for _, kline := range klines {
		if kline.OpenTime.After(start) && kline.CloseTime.Before(end) {
			filtered = append(filtered, kline)
		}
	}

	if len(filtered) == 0 {
		return
	}

	// Initialize with first kline
	high := filtered[0].High
	low := filtered[0].Low
	vwapSum := numerical.Zero()
	volumeSum := numerical.Zero()

	// Calculate high, low, and VWAP components
	for _, kline := range filtered {
		// Track highest high
		if kline.High.GreaterThan(high) {
			high = kline.High
		}

		// Track lowest low
		if kline.Low.LessThan(low) {
			low = kline.Low
		}

		// Accumulate for VWAP: sum(price * volume) / sum(volume)
		// Use close price for VWAP calculation
		vwapSum = vwapSum.Add(kline.Close.Mul(kline.Volume))
		volumeSum = volumeSum.Add(kline.Volume)
	}

	// Set high and low features
	featureMap[featureHigh1h], _ = high.Float64()
	featureMap[featureLow1h], _ = low.Float64()

	// Calculate VWAP if we have volume
	if !volumeSum.IsZero() {
		vwap := vwapSum.Div(volumeSum)
		featureMap[featureVWAP1h], _ = vwap.Float64()
	}
}
