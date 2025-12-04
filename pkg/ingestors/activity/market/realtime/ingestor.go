package realtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/ingestors"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
)

type ingestor struct {
	store            market.MarketData
	exchangeRegistry registry.ConnectorRegistry
	assetRegistry    registry.AssetRegistry
	logger           logging.ApplicationLogger
	healthStore      health.CoordinatorHealthStore
	notifier         ingestors.DataUpdateNotifier

	// WebSocket management
	wsContext context.Context
	wsCancel  context.CancelFunc
	isActive  bool
	mutex     sync.RWMutex

	activeConnections map[connector.ExchangeName]connector.WebSocketConnector

	// Track which instrument types are subscribed per asset
	subscriptions     map[portfolio.Asset][]connector.Instrument
	subscriptionMutex sync.RWMutex
}

func NewIngestor(
	store market.MarketData,
	exchangeRegistry registry.ConnectorRegistry,
	assetRegistry registry.AssetRegistry,
	logger logging.ApplicationLogger,
	healthStore health.CoordinatorHealthStore,
	notifier ingestors.DataUpdateNotifier,
) ingestors.RealtimeIngestor {
	return &ingestor{
		store:             store,
		exchangeRegistry:  exchangeRegistry,
		assetRegistry:     assetRegistry,
		logger:            logger,
		healthStore:       healthStore,
		notifier:          notifier,
		activeConnections: make(map[connector.ExchangeName]connector.WebSocketConnector),
		subscriptions:     make(map[portfolio.Asset][]connector.Instrument),
	}
}

// notifyDataUpdate signals that data was updated
func (ri *ingestor) notifyDataUpdate() {
	ri.notifier.Notify()
}

func (ri *ingestor) Start(ctx context.Context) error {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	if ri.isActive {
		return fmt.Errorf("realtime ingestor already active")
	}

	ri.wsContext, ri.wsCancel = context.WithCancel(ctx)
	ri.isActive = true

	// Get required assets from strategy configs
	tradingAssets := ri.assetRegistry.GetRequiredAssets()
	ri.logger.Info("📋 Required trading assets from strategies: %d assets", len(tradingAssets))
	for _, asset := range tradingAssets {
		ri.logger.Info("  - %s", asset.Symbol())
	}

	connectors := ri.exchangeRegistry.GetReadyWebSocketConnectors()

	if len(tradingAssets) == 0 {
		ri.logger.Info("❌ No assets required by enabled strategies - realtime ingestor won't start")
		return nil
	}

	ri.logger.Info("🔄 About to start WebSocket streams for trading assets...")

	// Start WebSocket streams for each exchange
	ri.logger.Info("🔗 Found %d trading WebSocket connectors", len(connectors))

	for _, conn := range connectors {
		exchangeName := conn.GetConnectorInfo().Name
		ri.logger.Info("🚀 Starting WebSocket stream for %s", exchangeName)
		go ri.startExchangeStream(conn)
	}

	ri.logger.Info("✅ Started WebSocket ingestion for %d assets", len(tradingAssets))
	return nil
}

func (ri *ingestor) startExchangeStream(wsConn connector.WebSocketConnector) {
	exchangeName := wsConn.GetConnectorInfo().Name

	if err := wsConn.StartWebSocket(); err != nil {
		ri.logger.Error("Failed to start WebSocket for %s: %v", exchangeName, err)
		return
	}

	// Get detailed asset requirements (which instrument types per asset)
	assetRequirements := ri.assetRegistry.GetAssetRequirements()

	// Subscribe to data based on actual requirements
	for _, req := range assetRequirements {
		// Subscribe to orderbooks for each required instrument type
		for _, instrumentType := range req.Instruments {
			if err := wsConn.SubscribeOrderBook(req.Asset, instrumentType); err != nil {
				ri.logger.Error("Failed to subscribe to %s orderbook for %s on %s: %v",
					instrumentType, req.Asset.Symbol(), exchangeName, err)
			} else {
				ri.logger.Info("✅ Subscribed to %s orderbook for %s on %s",
					instrumentType, req.Asset.Symbol(), exchangeName)

				// Track this subscription
				ri.subscriptionMutex.Lock()
				ri.subscriptions[req.Asset] = append(ri.subscriptions[req.Asset], instrumentType)
				ri.subscriptionMutex.Unlock()
			}
		}

		// Subscribe to klines for strategy analysis
		klineIntervals := []string{"1m", "5m", "15m", "1h"}
		for _, interval := range klineIntervals {
			if err := wsConn.SubscribeKlines(req.Asset, interval); err != nil {
				ri.logger.Error("Failed to subscribe to %s klines for %s on %s: %v",
					interval, req.Asset.Symbol(), exchangeName, err)
			}
		}
	}

	// Process real-time updates
	go ri.processOrderBookStream(wsConn, exchangeName)
	go ri.processKlineStream(wsConn, exchangeName)
	go ri.processErrorStream(wsConn, exchangeName)
}

func (ri *ingestor) processOrderBookStream(wsConn connector.WebSocketConnector, exchangeName connector.ExchangeName) {
	ri.logger.Info("🔄 Starting orderbook stream processing for %s", exchangeName)

	channels := wsConn.GetOrderBookChannels()
	ri.logger.Info("📊 Got %d dedicated orderbook channels for %s", len(channels), exchangeName)

	// Launch goroutine for each orderbook channel
	for channelKey, orderBookChan := range channels {
		go ri.processOrderBookChannel(channelKey, orderBookChan, exchangeName)
	}
}

// processOrderBookChannel processes orderbooks from a specific asset channel
func (ri *ingestor) processOrderBookChannel(channelKey string, orderBookChan <-chan connector.OrderBook, exchangeName connector.ExchangeName) {
	defer func() {
		if r := recover(); r != nil {
			ri.logger.Error("🔥 PANIC in processOrderBookChannel for %s/%s: %v", exchangeName, channelKey, r)
		}
	}()

	ri.logger.Info("🔄 Starting channel processor for orderbook %s on %s", channelKey, exchangeName)

	updateCount := 0
	for orderBookUpdate := range orderBookChan {
		updateCount++
		ri.logger.Debug("📊 Received orderbook update #%d for %s from channel %s",
			updateCount, orderBookUpdate.Asset.Symbol(), channelKey)

		// Get which instrument types we subscribed to for this asset
		ri.subscriptionMutex.RLock()
		instrumentTypes := ri.subscriptions[orderBookUpdate.Asset]
		ri.subscriptionMutex.RUnlock()

		// Store the orderbook update for each subscribed instrument type
		for _, instrumentType := range instrumentTypes {
			ri.store.UpdateOrderBook(
				orderBookUpdate.Asset,
				exchangeName,
				instrumentType,
				orderBookUpdate,
			)

			ri.logger.Debug("📊 Updated %s orderbook for %s on %s",
				instrumentType, orderBookUpdate.Asset.Symbol(), exchangeName)
		}

		// Notify coordinator that data was updated
		ri.notifyDataUpdate()

		// Report successful data receipt to health monitoring
		ri.healthStore.RecordDataReceived(exchangeName, health.DataTypeOrderbooks, health.SourceWebSocket, 0)
	}

	ri.logger.Info("📪 Orderbook channel %s closed for %s after %d updates", channelKey, exchangeName, updateCount)
}

func (ri *ingestor) processKlineStream(wsConn connector.WebSocketConnector, exchangeName connector.ExchangeName) {
	ri.logger.Info("🔄 Starting kline stream processing for %s", exchangeName)

	channels := wsConn.GetKlineChannels()
	ri.logger.Info("📊 Got %d dedicated kline channels for %s", len(channels), exchangeName)

	// Launch goroutine for each channel
	for channelKey, klineChan := range channels {
		go ri.processKlineChannel(channelKey, klineChan, exchangeName)
	}
}

// processKlineChannel processes klines from a specific channel
func (ri *ingestor) processKlineChannel(channelKey string, klineChan <-chan connector.Kline, exchangeName connector.ExchangeName) {
	defer func() {
		if r := recover(); r != nil {
			ri.logger.Error("🔥 PANIC in processKlineChannel for %s/%s: %v", exchangeName, channelKey, r)
		}
	}()

	ri.logger.Info("🔄 Starting channel processor for %s on %s", channelKey, exchangeName)

	for klineUpdate := range klineChan {
		asset := portfolio.NewAsset(klineUpdate.Symbol)
		ri.logger.Debug("📊 Received %s kline for %s on %s",
			klineUpdate.Interval, klineUpdate.Symbol, exchangeName)

		ri.store.UpdateKline(asset, exchangeName, klineUpdate)

		ri.notifier.Notify()
		ri.healthStore.RecordDataReceived(exchangeName, health.DataTypeKlines, health.SourceWebSocket, 0)
	}

	ri.logger.Info("📪 Kline channel %s closed for %s", channelKey, exchangeName)
}

func (ri *ingestor) processErrorStream(wsConn connector.WebSocketConnector, exchangeName connector.ExchangeName) {
	for err := range wsConn.ErrorChannel() {
		ri.logger.Error("WebSocket error for %s: %v", exchangeName, err)
		// Report error to health monitoring - affects all data types on this websocket
		ri.healthStore.RecordDataError(exchangeName, health.DataTypeKlines, err)
		ri.healthStore.RecordDataError(exchangeName, health.DataTypeOrderbooks, err)
	}
}

func (ri *ingestor) Stop() error {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	if !ri.isActive {
		return nil
	}

	if ri.wsCancel != nil {
		ri.wsCancel()
	}

	// Stop all WebSocket connections
	for _, conn := range ri.exchangeRegistry.GetReadyWebSocketConnectors() {
		conn.StopWebSocket()
	}

	ri.isActive = false
	ri.logger.Info("Stopped WebSocket ingestion")
	return nil
}

func (ri *ingestor) IsActive() bool {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	return ri.isActive
}

func (ri *ingestor) GetActiveConnections() map[connector.ExchangeName]connector.WebSocketConnector {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()

	result := make(map[connector.ExchangeName]connector.WebSocketConnector)
	for name, conn := range ri.activeConnections {
		result[name] = conn
	}
	return result
}
