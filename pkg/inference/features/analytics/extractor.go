package analytics

import (
	analyticsTypes "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// Feature name constants (must match pkg/inference/features/types.go)
const (
	// Volatility features
	featureVolatility5m    = "volatility_5m"
	featureVolatility1h    = "volatility_1h"
	featureVolatilityRatio = "volatility_ratio"

	// Volume features
	featureVolume1m    = "volume_1m"
	featureVolumeRatio = "volume_ratio"
)

// Extractor computes analytics-derived features (volatility, volume, etc.).
// It uses the analytics service which provides higher-level market analysis.
type Extractor struct {
	analytics analyticsTypes.Analytics
}

// NewExtractor creates a new analytics feature extractor.
func NewExtractor(analytics analyticsTypes.Analytics) *Extractor {
	return &Extractor{
		analytics: analytics,
	}
}

// Extract computes analytics features and adds them to the feature map.
// Currently supports: volatility (5m, 1h, ratio) and volume (ratio).
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

	// === VOLUME FEATURES ===

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

	return nil
}
