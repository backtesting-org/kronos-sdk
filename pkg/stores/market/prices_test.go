package market_test

import (
	"time"

	timeProvider "github.com/backtesting-org/kronos-sdk/pkg/runtime/time"
	"github.com/backtesting-org/kronos-sdk/pkg/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	marketTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/market"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Market Data Store - Prices", func() {
	var (
		store    marketTypes.MarketData
		btc      portfolio.Asset
		eth      portfolio.Asset
		provider temporal.TimeProvider
	)

	BeforeEach(func() {
		provider = timeProvider.NewTimeProvider()
		store = market.NewStore(provider)
		btc = portfolio.NewAsset("BTC")
		eth = portfolio.NewAsset("ETH")
	})

	Describe("UpdateAssetPrice", func() {
		Context("when adding a new price", func() {
			It("should store the price correctly", func() {
				now := time.Now()
				price := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					BidPrice:  numerical.NewFromFloat(49990),
					AskPrice:  numerical.NewFromFloat(50010),
					Volume24h: numerical.NewFromFloat(1000000),
					Change24h: numerical.NewFromFloat(2.5),
					Source:    "hyperliquid",
					Timestamp: now,
				}

				store.UpdateAssetPrice(btc, "hyperliquid", price)

				retrieved := store.GetAssetPrice(btc, "hyperliquid")
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(retrieved.BidPrice).To(Equal(numerical.NewFromFloat(49990)))
				Expect(retrieved.AskPrice).To(Equal(numerical.NewFromFloat(50010)))
				Expect(retrieved.Volume24h).To(Equal(numerical.NewFromFloat(1000000)))
				Expect(retrieved.Change24h).To(Equal(numerical.NewFromFloat(2.5)))
			})

			It("should handle multiple exchanges for the same asset", func() {
				now := time.Now()

				hyperPrice := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Source:    "hyperliquid",
					Timestamp: now,
				}

				bybitPrice := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50020),
					Source:    "bybit",
					Timestamp: now,
				}

				store.UpdateAssetPrice(btc, "hyperliquid", hyperPrice)
				store.UpdateAssetPrice(btc, "bybit", bybitPrice)

				hyperRetrieved := store.GetAssetPrice(btc, "hyperliquid")
				bybitRetrieved := store.GetAssetPrice(btc, "bybit")

				Expect(hyperRetrieved).NotTo(BeNil())
				Expect(bybitRetrieved).NotTo(BeNil())
				Expect(hyperRetrieved.Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(bybitRetrieved.Price).To(Equal(numerical.NewFromFloat(50020)))
			})

			It("should handle multiple assets for the same exchange", func() {
				now := time.Now()

				btcPrice := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Timestamp: now,
				}

				ethPrice := connector.Price{
					Symbol:    "ETH",
					Price:     numerical.NewFromFloat(3000),
					Timestamp: now,
				}

				store.UpdateAssetPrice(btc, "hyperliquid", btcPrice)
				store.UpdateAssetPrice(eth, "hyperliquid", ethPrice)

				btcRetrieved := store.GetAssetPrice(btc, "hyperliquid")
				ethRetrieved := store.GetAssetPrice(eth, "hyperliquid")

				Expect(btcRetrieved).NotTo(BeNil())
				Expect(ethRetrieved).NotTo(BeNil())
				Expect(btcRetrieved.Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(ethRetrieved.Price).To(Equal(numerical.NewFromFloat(3000)))
			})

			It("should update existing price", func() {
				now := time.Now()

				initialPrice := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Timestamp: now,
				}

				store.UpdateAssetPrice(btc, "hyperliquid", initialPrice)

				updatedPrice := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(51000),
					Timestamp: now.Add(time.Minute),
				}

				store.UpdateAssetPrice(btc, "hyperliquid", updatedPrice)

				retrieved := store.GetAssetPrice(btc, "hyperliquid")
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Price).To(Equal(numerical.NewFromFloat(51000)))
			})
		})
	})

	Describe("UpdateAssetPrices", func() {
		Context("when adding prices for multiple exchanges at once", func() {
			It("should store all prices correctly", func() {
				now := time.Now()

				prices := map[connector.ExchangeName]connector.Price{
					"hyperliquid": {
						Symbol:    "BTC",
						Price:     numerical.NewFromFloat(50000),
						Timestamp: now,
					},
					"bybit": {
						Symbol:    "BTC",
						Price:     numerical.NewFromFloat(50020),
						Timestamp: now,
					},
					"paradex": {
						Symbol:    "BTC",
						Price:     numerical.NewFromFloat(50010),
						Timestamp: now,
					},
				}

				store.UpdateAssetPrices(btc, prices)

				priceMap := store.GetAssetPrices(btc)

				Expect(priceMap).To(HaveLen(3))
				Expect(priceMap["hyperliquid"].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(priceMap["bybit"].Price).To(Equal(numerical.NewFromFloat(50020)))
				Expect(priceMap["paradex"].Price).To(Equal(numerical.NewFromFloat(50010)))
			})

			It("should merge with existing prices", func() {
				now := time.Now()

				// Add initial price
				store.UpdateAssetPrice(btc, "hyperliquid", connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Timestamp: now,
				})

				// Add batch prices
				prices := map[connector.ExchangeName]connector.Price{
					"bybit": {
						Symbol:    "BTC",
						Price:     numerical.NewFromFloat(50020),
						Timestamp: now,
					},
				}

				store.UpdateAssetPrices(btc, prices)

				priceMap := store.GetAssetPrices(btc)

				Expect(priceMap).To(HaveLen(2))
				Expect(priceMap["hyperliquid"].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(priceMap["bybit"].Price).To(Equal(numerical.NewFromFloat(50020)))
			})
		})
	})

	Describe("GetAssetPrice", func() {
		Context("when retrieving a specific price", func() {
			It("should return nil for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				price := store.GetAssetPrice(unknown, "hyperliquid")
				Expect(price).To(BeNil())
			})

			It("should return nil for unknown exchange", func() {
				store.UpdateAssetPrice(btc, "hyperliquid", connector.Price{
					Symbol: "BTC",
					Price:  numerical.NewFromFloat(50000),
				})

				price := store.GetAssetPrice(btc, "unknown_exchange")
				Expect(price).To(BeNil())
			})
		})
	})

	Describe("GetAssetPrices", func() {
		Context("when retrieving all prices for an asset", func() {
			It("should return prices from all exchanges", func() {
				now := time.Now()

				store.UpdateAssetPrice(btc, "hyperliquid", connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Timestamp: now,
				})
				store.UpdateAssetPrice(btc, "bybit", connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50020),
					Timestamp: now,
				})

				priceMap := store.GetAssetPrices(btc)

				Expect(priceMap).To(HaveLen(2))
				Expect(priceMap["hyperliquid"].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(priceMap["bybit"].Price).To(Equal(numerical.NewFromFloat(50020)))
			})

			It("should return empty map for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				priceMap := store.GetAssetPrices(unknown)
				Expect(priceMap).To(BeEmpty())
			})
		})
	})

	Describe("Price spread calculations", func() {
		Context("when analyzing bid-ask spread", func() {
			It("should store bid and ask prices correctly", func() {
				now := time.Now()
				price := connector.Price{
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					BidPrice:  numerical.NewFromFloat(49990),
					AskPrice:  numerical.NewFromFloat(50010),
					Timestamp: now,
				}

				store.UpdateAssetPrice(btc, "hyperliquid", price)

				retrieved := store.GetAssetPrice(btc, "hyperliquid")
				Expect(retrieved).NotTo(BeNil())

				// Spread should be 20 (50010 - 49990)
				spread := retrieved.AskPrice.Sub(retrieved.BidPrice)
				Expect(spread).To(Equal(numerical.NewFromFloat(20)))
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("when multiple goroutines update prices", func() {
			It("should handle concurrent writes safely", func() {
				done := make(chan bool)
				iterations := 100

				// Writer 1 - updates BTC on hyperliquid
				go func() {
					for i := 0; i < iterations; i++ {
						store.UpdateAssetPrice(btc, "hyperliquid", connector.Price{
							Symbol: "BTC",
							Price:  numerical.NewFromFloat(float64(50000 + i)),
						})
					}
					done <- true
				}()

				// Writer 2 - updates BTC on bybit
				go func() {
					for i := 0; i < iterations; i++ {
						store.UpdateAssetPrice(btc, "bybit", connector.Price{
							Symbol: "BTC",
							Price:  numerical.NewFromFloat(float64(50100 + i)),
						})
					}
					done <- true
				}()

				// Reader - reads prices continuously
				go func() {
					for i := 0; i < iterations; i++ {
						_ = store.GetAssetPrices(btc)
						_ = store.GetAssetPrice(btc, "hyperliquid")
					}
					done <- true
				}()

				// Wait for all goroutines to complete
				<-done
				<-done
				<-done

				// Verify data is consistent
				priceMap := store.GetAssetPrices(btc)
				Expect(priceMap).To(HaveLen(2))
			})
		})
	})
})
