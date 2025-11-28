package features

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"go.uber.org/fx"
)

// Extractor computes features from market data for ML inference.
// Each extractor implementation is responsible for computing a specific
// category of features (e.g., market data, orderbook, technical indicators).
type Extractor interface {
	// Extract computes features for the given asset and adds them to the provided map.
	// It should not overwrite existing keys unless explicitly intended.
	// Returns an error if feature computation fails.
	Extract(asset portfolio.Asset, features map[string]float64) error
}

// AggregatorParams defines the dependencies for the feature aggregator.
type AggregatorParams struct {
	fx.In

	Extractors []Extractor `group:"feature_extractors"`
}

// Aggregator combines multiple extractors into a single feature extraction pipeline.
type Aggregator struct {
	extractors []Extractor
}

// NewAggregator creates a new feature aggregator with the given extractors via fx.
func NewAggregator(params AggregatorParams) *Aggregator {
	return &Aggregator{
		extractors: params.Extractors,
	}
}

// Extract runs all registered extractors for the given asset and returns the combined feature map.
// If any extractor fails, it continues processing others and returns the first error encountered.
func (a *Aggregator) Extract(asset portfolio.Asset) (map[string]float64, error) {
	features := make(map[string]float64)
	var firstErr error

	for _, extractor := range a.extractors {
		if err := extractor.Extract(asset, features); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return features, firstErr
}
