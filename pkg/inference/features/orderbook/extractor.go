package orderbook

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// Feature name constants (must match pkg/inference/features/types.go)
const (
	featureBidAskSpread       = "bid_ask_spread"
	featureSpreadBps          = "spread_bps"
	featureOrderbookImbalance = "orderbook_imbalance"
	featureBidDepth5          = "bid_depth_5"
	featureAskDepth5          = "ask_depth_5"
	featureDepthRatio         = "depth_ratio"
	featureWeightedMid        = "weighted_mid"
)

// Extractor computes orderbook-based features (spread, depth, imbalance, etc.).
// It uses the market analytics service to access orderbook data.
type Extractor struct {
	market analytics.Market
}

// NewExtractor creates a new orderbook feature extractor.
func NewExtractor(market analytics.Market) *Extractor {
	return &Extractor{
		market: market,
	}
}

// Extract computes orderbook features and adds them to the feature map.
// Currently supports: spread, spread_bps, imbalance, bid/ask depth, depth ratio, weighted mid.
func (e *Extractor) Extract(asset portfolio.Asset, featureMap map[string]float64) error {
	// Get order book
	orderBook, err := e.market.OrderBook(asset)
	if err != nil || orderBook == nil {
		return err
	}

	// Need at least one bid and ask
	if len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
		return nil
	}

	bestBid := orderBook.Bids[0].Price
	bestAsk := orderBook.Asks[0].Price

	// Calculate bid-ask spread
	spread := bestAsk.Sub(bestBid)
	featureMap[featureBidAskSpread], _ = spread.Float64()

	// Calculate mid price
	midPrice := bestBid.Add(bestAsk).Div(numerical.NewFromInt(2))

	// Calculate spread in basis points (bps): (spread / mid_price) * 10000
	if !midPrice.IsZero() {
		spreadBps := spread.Div(midPrice).Mul(numerical.NewFromInt(10000))
		featureMap[featureSpreadBps], _ = spreadBps.Float64()
	}

	// Calculate bid and ask depth (sum of sizes for top 5 levels)
	bidDepth := numerical.Zero()
	askDepth := numerical.Zero()

	// Sum bid volumes (up to 5 levels)
	for i := 0; i < len(orderBook.Bids) && i < 5; i++ {
		bidDepth = bidDepth.Add(orderBook.Bids[i].Quantity)
	}

	// Sum ask volumes (up to 5 levels)
	for i := 0; i < len(orderBook.Asks) && i < 5; i++ {
		askDepth = askDepth.Add(orderBook.Asks[i].Quantity)
	}

	featureMap[featureBidDepth5], _ = bidDepth.Float64()
	featureMap[featureAskDepth5], _ = askDepth.Float64()

	// Calculate orderbook imbalance: bid_volume / (bid_volume + ask_volume)
	totalDepth := bidDepth.Add(askDepth)
	if !totalDepth.IsZero() {
		imbalance := bidDepth.Div(totalDepth)
		featureMap[featureOrderbookImbalance], _ = imbalance.Float64()
	}

	// Calculate depth ratio: bid_depth / ask_depth
	if !askDepth.IsZero() {
		depthRatio := bidDepth.Div(askDepth)
		featureMap[featureDepthRatio], _ = depthRatio.Float64()
	}

	// Calculate volume-weighted mid price
	// weighted_mid = (bid_price * ask_volume + ask_price * bid_volume) / (bid_volume + ask_volume)
	bestBidQty := orderBook.Bids[0].Quantity
	bestAskQty := orderBook.Asks[0].Quantity
	totalQty := bestBidQty.Add(bestAskQty)

	if !totalQty.IsZero() {
		weightedMid := bestBid.Mul(bestAskQty).Add(bestAsk.Mul(bestBidQty)).Div(totalQty)
		featureMap[featureWeightedMid], _ = weightedMid.Float64()
	}

	return nil
}
