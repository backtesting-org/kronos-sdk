package features

// Feature name constants for ML inference.
// These define the standard feature names that will be sent to user's inference server.
// Users can choose which features their model actually uses.
const (
	// Raw market data features (from exchange)
	FeatureMidPrice    = "mid_price"    // (bid + ask) / 2
	FeatureBidPrice    = "bid_price"    // best bid price
	FeatureAskPrice    = "ask_price"    // best ask price
	FeatureLastPrice   = "last_price"   // last trade price
	FeatureMarkPrice   = "mark_price"   // exchange mark price
	FeatureIndexPrice  = "index_price"  // underlying index price
	FeatureVolume24h   = "volume_24h"   // 24 hour volume
	FeatureFundingRate = "funding_rate" // current funding rate (perps)

	// Orderbook metric features
	FeatureBidAskSpread       = "bid_ask_spread"       // ask - bid
	FeatureSpreadBps          = "spread_bps"           // spread / mid_price * 10000
	FeatureOrderbookImbalance = "orderbook_imbalance"  // bid_volume / (bid_volume + ask_volume)
	FeatureBidDepth5          = "bid_depth_5"          // sum of bid sizes (5 levels)
	FeatureAskDepth5          = "ask_depth_5"          // sum of ask sizes (5 levels)
	FeatureDepthRatio         = "depth_ratio"          // bid_depth / ask_depth
	FeatureWeightedMid        = "weighted_mid"         // volume-weighted mid price

	// Price metric features
	FeatureReturn1m = "return_1m" // 1-minute return
	FeatureReturn5m = "return_5m" // 5-minute return
	FeatureReturn1h = "return_1h" // 1-hour return
	FeatureHigh1h   = "high_1h"   // 1-hour high
	FeatureLow1h    = "low_1h"    // 1-hour low
	FeatureVWAP1h   = "vwap_1h"   // volume-weighted average price (1h)

	// Volatility features
	FeatureVolatility5m    = "volatility_5m"    // realized volatility (5min window)
	FeatureVolatility1h    = "volatility_1h"    // realized volatility (1h window)
	FeatureVolatilityRatio = "volatility_ratio" // short_vol / long_vol

	// Volume metric features
	FeatureVolume1m       = "volume_1m"        // 1-minute volume
	FeatureVolume5m       = "volume_5m"        // 5-minute volume
	FeatureVolumeRatio    = "volume_ratio"     // current_vol / avg_vol
	FeatureBuyVolumeRatio = "buy_volume_ratio" // buy_vol / total_vol
	FeatureTradeCount1m   = "trade_count_1m"   // number of trades (1min)

	// Technical indicator features
	FeatureRSI14       = "rsi_14"       // RSI (14 periods)
	FeatureMACD        = "macd"         // MACD line
	FeatureMACDSignal  = "macd_signal"  // MACD signal line
	FeatureBBUpper     = "bb_upper"     // Bollinger Band upper
	FeatureBBLower     = "bb_lower"     // Bollinger Band lower
	FeatureBBPosition  = "bb_position"  // (price - lower) / (upper - lower)

	// Time-based features
	FeatureHour      = "hour"        // hour of day (0-23)
	FeatureDayOfWeek = "day_of_week" // day of week (0=Monday, 6=Sunday)
	FeatureMinute    = "minute"      // minute of hour (0-59)

	// Market microstructure features
	FeaturePriceMomentum1m  = "price_momentum_1m"  // directional price movement
	FeatureAggressorRatio   = "aggressor_ratio"    // buy_aggressor / total_trades
	FeatureLargeTradeRatio  = "large_trade_ratio"  // trades > threshold / total_trades
)
