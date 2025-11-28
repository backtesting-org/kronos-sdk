package technical

import (
	"context"

	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// Feature name constants (must match pkg/inference/features/types.go)
const (
	featureRSI14       = "rsi_14"
	featureMACD        = "macd"
	featureMACDSignal  = "macd_signal"
	featureBBUpper     = "bb_upper"
	featureBBLower     = "bb_lower"
	featureBBPosition  = "bb_position"
	featureATR14       = "atr_14"
	featureEMA20       = "ema_20"
	featureEMA50       = "ema_50"
	featureStochasticK = "stochastic_k"
	featureStochasticD = "stochastic_d"
)

// Extractor computes technical indicator features using the analytics service.
// It leverages existing indicator implementations (RSI, MACD, Bollinger Bands, etc.)
// and extracts them as ML features.
type Extractor struct {
	indicators analytics.Indicators
}

// NewExtractor creates a new technical indicator feature extractor.
func NewExtractor(indicators analytics.Indicators) *Extractor {
	return &Extractor{
		indicators: indicators,
	}
}

// Extract computes technical indicator features and adds them to the feature map.
// Currently supports: RSI, MACD, Bollinger Bands, ATR, SMA, EMA, Stochastic.
//
// Note: This requires an asset to be available in the context.
// TODO: Add context key for asset once orchestration is wired up.
func (e *Extractor) Extract(ctx context.Context, featureMap map[string]float64) error {
	// For now, this is a placeholder implementation that returns early
	asset, ok := e.getAssetFromContext(ctx)
	if !ok {
		// No asset available - skip extraction
		return nil
	}

	// Extract RSI (14-period)
	if rsi, err := e.indicators.RSI(asset, 14); err == nil {
		featureMap[featureRSI14], _ = rsi.Float64()
	}

	// Extract MACD (12, 26, 9)
	if macd, err := e.indicators.MACD(asset, 12, 26, 9); err == nil {
		featureMap[featureMACD], _ = macd.MACD.Float64()
		featureMap[featureMACDSignal], _ = macd.Signal.Float64()
	}

	// Extract Bollinger Bands (20-period, 2 std dev)
	if bb, err := e.indicators.BollingerBands(asset, 20, 2.0); err == nil {
		featureMap[featureBBUpper], _ = bb.Upper.Float64()
		featureMap[featureBBLower], _ = bb.Lower.Float64()

		// Calculate BB position: (price - lower) / (upper - lower)
		// This shows where price is relative to the bands (0-1 range)
		upper, _ := bb.Upper.Float64()
		lower, _ := bb.Lower.Float64()
		middle, _ := bb.Middle.Float64()

		if upper > lower {
			position := (middle - lower) / (upper - lower)
			featureMap[featureBBPosition] = position
		}
	}

	// Extract ATR (14-period) - measures volatility
	if atr, err := e.indicators.ATR(asset, 14); err == nil {
		featureMap[featureATR14], _ = atr.Float64()
	}

	// Extract EMA (20-period and 50-period) - trend indicators
	if ema20, err := e.indicators.EMA(asset, 20); err == nil {
		featureMap[featureEMA20], _ = ema20.Float64()
	}

	if ema50, err := e.indicators.EMA(asset, 50); err == nil {
		featureMap[featureEMA50], _ = ema50.Float64()
	}

	// Extract Stochastic (14, 3) - momentum oscillator
	if stoch, err := e.indicators.Stochastic(asset, 14, 3); err == nil {
		featureMap[featureStochasticK], _ = stoch.K.Float64()
		featureMap[featureStochasticD], _ = stoch.D.Float64()
	}

	return nil
}

// getAssetFromContext retrieves the asset from context.
// This is a placeholder until we define the context key structure.
func (e *Extractor) getAssetFromContext(ctx context.Context) (portfolio.Asset, bool) {
	// Example:
	// asset, ok := ctx.Value(contextKeyAsset).(portfolio.Asset)
	// return asset, ok
	return portfolio.Asset{}, false
}
