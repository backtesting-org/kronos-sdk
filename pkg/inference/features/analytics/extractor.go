package analytics

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	analyticsTypes "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

// Feature name constants (must match pkg/inference/features/types.go)
const (
	// Volatility features
	featureVolatility5m    = "volatility_5m"
	featureVolatility1h    = "volatility_1h"
	featureVolatilityRatio = "volatility_ratio"

	// Volume features
	featureVolume1m       = "volume_1m"
	featureVolume5m       = "volume_5m"
	featureVolumeRatio    = "volume_ratio"
	featureBuyVolumeRatio = "buy_volume_ratio"
	featureTradeCount1m   = "trade_count_1m"

	// Time-based features
	featureHour      = "hour"
	featureDayOfWeek = "day_of_week"
	featureMinute    = "minute"
)

// Extractor computes analytics-derived features (volatility, volume, time).
// It uses the analytics service for higher-level analysis, trades store for trade data,
// and time provider for temporal features.
type Extractor struct {
	analytics    analyticsTypes.Analytics
	trades       activity.Trades
	timeProvider temporal.TimeProvider
}

// NewExtractor creates a new analytics feature extractor.
func NewExtractor(analytics analyticsTypes.Analytics, trades activity.Trades, timeProvider temporal.TimeProvider) *Extractor {
	return &Extractor{
		analytics:    analytics,
		trades:       trades,
		timeProvider: timeProvider,
	}
}

// Extract computes analytics features and adds them to the feature map.
// Currently supports: volatility (5m, 1h, ratio), volume (1m, 5m, ratio, buy ratio, trade count),
// and time-based features (hour, day of week, minute).
func (e *Extractor) Extract(asset portfolio.Asset, featureMap map[string]float64) error {
	// Calculate 5-minute volatility
	vol5m, err := e.analytics.Volatility(asset, 5, analyticsTypes.AnalyticsOptions{
		Interval: "1m", // 5 periods of 1-minute data
	})
	if err == nil {
		featureMap[featureVolatility5m], _ = vol5m.Float64()
	}

	// Calculate 1-hour volatility
	vol1h, err := e.analytics.Volatility(asset, 60, analyticsTypes.AnalyticsOptions{
		Interval: "1m", // 60 periods of 1-minute data
	})
	if err == nil {
		featureMap[featureVolatility1h], _ = vol1h.Float64()
	}

	// Calculate volatility ratio (short/long)
	if vol5m, ok5 := featureMap[featureVolatility5m]; ok5 {
		if vol1h, ok1 := featureMap[featureVolatility1h]; ok1 && vol1h != 0 {
			ratio := vol5m / vol1h
			featureMap[featureVolatilityRatio] = ratio
		}
	}

	// Get volume analysis (20-period lookback)
	volAnalysis, err := e.analytics.VolumeAnalysis(asset, 20, analyticsTypes.AnalyticsOptions{
		Interval: "1m",
	})
	if err == nil {
		// Extract current volume (1-minute)
		featureMap[featureVolume1m], _ = volAnalysis.CurrentVolume.Float64()

		// Extract volume ratio (current volume / average volume)
		featureMap[featureVolumeRatio], _ = volAnalysis.VolumeRatio.Float64()
	}

	// Get recent trades for trade-based volume features
	now := e.timeProvider.Now()
	trades1m := e.getTradesInWindow(asset, now.Add(-1*time.Minute), now)
	trades5m := e.getTradesInWindow(asset, now.Add(-5*time.Minute), now)

	if len(trades1m) > 0 {
		// Trade count in last minute
		featureMap[featureTradeCount1m] = float64(len(trades1m))

		// Buy volume ratio (buy volume / total volume)
		buyVol, totalVol := e.calculateBuySellVolumes(trades1m)
		if totalVol > 0 {
			featureMap[featureBuyVolumeRatio] = buyVol / totalVol
		}
	}

	if len(trades5m) > 0 {
		// Total volume in last 5 minutes
		var vol5m numerical.Decimal
		for _, trade := range trades5m {
			vol5m = vol5m.Add(trade.Quantity)
		}
		featureMap[featureVolume5m], _ = vol5m.Float64()
	}

	// Time-based features
	// Extract hour of day (0-23)
	featureMap[featureHour] = float64(now.Hour())

	// Extract day of week (0=Monday, 6=Sunday)
	// Go's time.Weekday is 0=Sunday, so we need to convert
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 6 // Sunday becomes 6
	} else {
		weekday = weekday - 1 // Monday(1)=0, Tuesday(2)=1, ..., Saturday(6)=5
	}
	featureMap[featureDayOfWeek] = float64(weekday)

	// Extract minute of hour (0-59)
	featureMap[featureMinute] = float64(now.Minute())

	return nil
}

// getTradesInWindow retrieves trades for an asset within a time window
func (e *Extractor) getTradesInWindow(asset portfolio.Asset, start, end time.Time) []connector.Trade {
	allTrades := e.trades.GetTradesByAsset(asset)

	var filtered []connector.Trade
	for _, trade := range allTrades {
		if trade.Timestamp.After(start) && trade.Timestamp.Before(end) {
			filtered = append(filtered, trade)
		}
	}
	return filtered
}

// calculateBuySellVolumes returns (buyVolume, totalVolume) as float64
func (e *Extractor) calculateBuySellVolumes(trades []connector.Trade) (float64, float64) {
	var buyVol, sellVol numerical.Decimal

	for _, trade := range trades {
		if trade.Side == connector.OrderSideBuy {
			buyVol = buyVol.Add(trade.Quantity)
		} else {
			sellVol = sellVol.Add(trade.Quantity)
		}
	}

	totalVol := buyVol.Add(sellVol)
	buyVolFloat, _ := buyVol.Float64()
	totalVolFloat, _ := totalVol.Float64()

	return buyVolFloat, totalVolFloat
}
