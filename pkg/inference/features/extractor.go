package features

import (
	"context"

	"go.uber.org/fx"
)

// Extractor computes features from market data for ML inference.
// Each extractor implementation is responsible for computing a specific
// category of features (e.g., orderbook, volatility, technical indicators).
type Extractor interface {
	// Extract computes features and adds them to the provided map.
	// It should not overwrite existing keys unless explicitly intended.
	// Returns an error if feature computation fails.
	Extract(ctx context.Context, features map[string]float64) error
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

// Extract runs all registered extractors and returns the combined feature map.
// If any extractor fails, it continues processing others and returns the first error encountered.
func (a *Aggregator) Extract(ctx context.Context) (map[string]float64, error) {
	features := make(map[string]float64)
	var firstErr error

	for _, extractor := range a.extractors {
		if err := extractor.Extract(ctx, features); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return features, firstErr
}
