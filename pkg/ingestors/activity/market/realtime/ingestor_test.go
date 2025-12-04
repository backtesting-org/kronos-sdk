package realtime_test

import (
	"context"
	"testing"
	"time"

	mockConnector "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	mockIngestors "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/data/ingestors"
	mockMarket "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	mockHealth "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/health"
	mockRegistry "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/ingestors/activity/market/realtime"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	ingestorTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/ingestors"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	loggingTypes "github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestRealtime(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Realtime Ingestor Suite")
}

// noopLogger is a no-op logger implementation for testing
type noopLogger struct{}

func (n *noopLogger) Fatal(_ string, _ ...interface{})                    {}
func (n *noopLogger) ErrorWithDebug(_ string, _ []byte, _ ...interface{}) {}
func (n *noopLogger) Info(_ string, _ ...interface{})                     {}
func (n *noopLogger) Debug(_ string, _ ...interface{})                    {}
func (n *noopLogger) Warn(_ string, _ ...interface{})                     {}
func (n *noopLogger) Error(_ string, _ ...interface{})                    {}

var _ = Describe("Ingestor", func() {
	var (
		ingestor             ingestorTypes.RealtimeIngestor
		mockStore            *mockMarket.MarketData
		mockExchangeRegistry *mockRegistry.ConnectorRegistry
		mockAssetRegistry    *mockRegistry.AssetRegistry
		logger               loggingTypes.ApplicationLogger
		mockHealthStore      *mockHealth.CoordinatorHealthStore
		mockNotifier         *mockIngestors.DataUpdateNotifier
		mockWSConn           *mockConnector.WebSocketConnector
		ctx                  context.Context
		cancel               context.CancelFunc
	)

	BeforeEach(func() {
		mockStore = mockMarket.NewMarketData(GinkgoT())
		mockExchangeRegistry = mockRegistry.NewConnectorRegistry(GinkgoT())
		mockAssetRegistry = mockRegistry.NewAssetRegistry(GinkgoT())
		logger = &noopLogger{}
		mockHealthStore = mockHealth.NewCoordinatorHealthStore(GinkgoT())
		mockNotifier = mockIngestors.NewDataUpdateNotifier(GinkgoT())
		mockWSConn = mockConnector.NewWebSocketConnector(GinkgoT())

		ctx, cancel = context.WithCancel(context.Background())

		ingestor = realtime.NewIngestor(
			mockStore,
			mockExchangeRegistry,
			mockAssetRegistry,
			logger,
			mockHealthStore,
			mockNotifier,
		)
	})

	AfterEach(func() {
		cancel()
		// Give goroutines time to stop
		time.Sleep(50 * time.Millisecond)
	})

	Describe("WebSocket data ingestion", func() {
		Context("when orderbook channel closes after initial snapshot", func() {
			It("should process orderbook updates before channel closure", func() {
				orderbookChan := make(chan connector.OrderBook, 10)
				errorChan := make(chan error, 10)

				orderbookChannels := map[string]<-chan connector.OrderBook{
					"BTC-PERP": orderbookChan,
				}
				klineChannels := map[string]<-chan connector.Kline{}

				// Setup all expectations using EXPECT()
				mockWSConn.EXPECT().GetOrderBookChannels().Return(orderbookChannels).Maybe()
				mockWSConn.EXPECT().GetKlineChannels().Return(klineChannels).Maybe()
				mockWSConn.EXPECT().ErrorChannel().Return((<-chan error)(errorChan)).Maybe()
				mockWSConn.EXPECT().GetConnectorInfo().Return(&connector.Info{Name: "hyperliquid"}).Maybe()
				mockWSConn.EXPECT().StartWebSocket().Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeOrderBook(mock.Anything, mock.Anything).Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeKlines(mock.Anything, mock.Anything).Return(nil).Maybe()

				mockAssetRegistry.EXPECT().GetRequiredAssets().Return([]portfolio.Asset{
					portfolio.NewAsset("BTC"),
				}).Maybe()

				mockAssetRegistry.EXPECT().GetAssetRequirements().Return([]registry.AssetRequirement{
					{
						Asset:       portfolio.NewAsset("BTC"),
						Instruments: []connector.Instrument{connector.TypePerpetual},
					},
				}).Maybe()

				mockExchangeRegistry.EXPECT().GetReadyWebSocketConnectors().Return([]connector.WebSocketConnector{mockWSConn}).Maybe()

				orderbook1 := connector.OrderBook{
					Asset: portfolio.NewAsset("BTC"),
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50100), Quantity: numerical.NewFromFloat(1.0)},
					},
				}

				mockStore.EXPECT().UpdateOrderBook(
					portfolio.NewAsset("BTC"),
					connector.ExchangeName("hyperliquid"),
					connector.TypePerpetual,
					orderbook1).Once()

				mockHealthStore.EXPECT().RecordDataReceived(connector.ExchangeName("hyperliquid"), mock.Anything, mock.Anything, mock.Anything).Once()
				mockNotifier.EXPECT().Notify().Once()

				_ = ingestor.Start(ctx)
				time.Sleep(200 * time.Millisecond)

				orderbookChan <- orderbook1
				time.Sleep(200 * time.Millisecond)

				mockStore.AssertExpectations(GinkgoT())
				mockNotifier.AssertExpectations(GinkgoT())

				close(orderbookChan)
				time.Sleep(50 * time.Millisecond)
			})
		})

		Context("when receiving multiple kline updates", func() {
			It("should process all updates continuously", func() {
				klineChan := make(chan connector.Kline, 10)
				errorChan := make(chan error, 10)

				orderbookChannels := map[string]<-chan connector.OrderBook{}
				klineChannels := map[string]<-chan connector.Kline{
					"BTC-1m": klineChan,
				}

				mockWSConn.EXPECT().GetKlineChannels().Return(klineChannels).Maybe()
				mockWSConn.EXPECT().GetOrderBookChannels().Return(orderbookChannels).Maybe()
				mockWSConn.EXPECT().ErrorChannel().Return((<-chan error)(errorChan)).Maybe()
				mockWSConn.EXPECT().GetConnectorInfo().Return(&connector.Info{Name: "hyperliquid"}).Maybe()
				mockWSConn.EXPECT().StartWebSocket().Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeOrderBook(mock.Anything, mock.Anything).Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeKlines(mock.Anything, mock.Anything).Return(nil).Maybe()

				mockAssetRegistry.EXPECT().GetRequiredAssets().Return([]portfolio.Asset{
					portfolio.NewAsset("BTC"),
				}).Maybe()

				mockAssetRegistry.EXPECT().GetAssetRequirements().Return([]registry.AssetRequirement{
					{
						Asset:       portfolio.NewAsset("BTC"),
						Instruments: []connector.Instrument{connector.TypePerpetual},
					},
				}).Maybe()

				mockExchangeRegistry.EXPECT().GetReadyWebSocketConnectors().Return([]connector.WebSocketConnector{mockWSConn}).Maybe()

				kline1 := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					Open:      numerical.NewFromFloat(50000),
					High:      numerical.NewFromFloat(50100),
					Low:       numerical.NewFromFloat(49900),
					Close:     numerical.NewFromFloat(50050),
					CloseTime: time.Now(),
				}

				kline2 := connector.Kline{
					Symbol:    "ETH",
					Interval:  "1m",
					Open:      numerical.NewFromFloat(3000),
					High:      numerical.NewFromFloat(3010),
					Low:       numerical.NewFromFloat(2990),
					Close:     numerical.NewFromFloat(3005),
					CloseTime: time.Now(),
				}

				mockStore.EXPECT().UpdateKline(portfolio.NewAsset("BTC"), connector.ExchangeName("hyperliquid"), kline1).Once()
				mockStore.EXPECT().UpdateKline(portfolio.NewAsset("ETH"), connector.ExchangeName("hyperliquid"), kline2).Once()
				mockHealthStore.EXPECT().RecordDataReceived(connector.ExchangeName("hyperliquid"), mock.Anything, mock.Anything, mock.Anything).Times(2)
				mockNotifier.EXPECT().Notify().Times(2)

				_ = ingestor.Start(ctx)
				time.Sleep(200 * time.Millisecond)

				klineChan <- kline1
				time.Sleep(150 * time.Millisecond)

				klineChan <- kline2
				time.Sleep(150 * time.Millisecond)

				mockStore.AssertExpectations(GinkgoT())
				mockHealthStore.AssertExpectations(GinkgoT())
				mockNotifier.AssertExpectations(GinkgoT())
			})
		})

		Context("when kline channel closes unexpectedly", func() {
			It("should handle channel closure gracefully", func() {
				klineChan := make(chan connector.Kline, 10)
				errorChan := make(chan error, 10)

				orderbookChannels := map[string]<-chan connector.OrderBook{}
				klineChannels := map[string]<-chan connector.Kline{
					"BTC-1m": klineChan,
				}

				mockWSConn.EXPECT().GetKlineChannels().Return(klineChannels).Maybe()
				mockWSConn.EXPECT().GetOrderBookChannels().Return(orderbookChannels).Maybe()
				mockWSConn.EXPECT().ErrorChannel().Return((<-chan error)(errorChan)).Maybe()
				mockWSConn.EXPECT().GetConnectorInfo().Return(&connector.Info{Name: "hyperliquid"}).Maybe()
				mockWSConn.EXPECT().StartWebSocket().Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeOrderBook(mock.Anything, mock.Anything).Return(nil).Maybe()
				mockWSConn.EXPECT().SubscribeKlines(mock.Anything, mock.Anything).Return(nil).Maybe()

				mockAssetRegistry.EXPECT().GetRequiredAssets().Return([]portfolio.Asset{
					portfolio.NewAsset("BTC"),
				}).Maybe()

				mockAssetRegistry.EXPECT().GetAssetRequirements().Return([]registry.AssetRequirement{
					{
						Asset:       portfolio.NewAsset("BTC"),
						Instruments: []connector.Instrument{connector.TypePerpetual},
					},
				}).Maybe()

				mockExchangeRegistry.EXPECT().GetReadyWebSocketConnectors().Return([]connector.WebSocketConnector{mockWSConn}).Maybe()

				_ = ingestor.Start(ctx)
				time.Sleep(100 * time.Millisecond)

				close(klineChan)
				time.Sleep(100 * time.Millisecond)
			})
		})
	})
})
