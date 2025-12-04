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

var _ = Describe("Market Data Store - OrderBooks", func() {
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

	Describe("UpdateOrderBook", func() {
		Context("when adding a new orderbook", func() {
			It("should store the orderbook correctly", func() {
				// Create an orderbook
				now := time.Now()
				orderBook := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
						{Price: numerical.NewFromFloat(49990), Quantity: numerical.NewFromFloat(2.0)},
						{Price: numerical.NewFromFloat(49980), Quantity: numerical.NewFromFloat(3.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)},
						{Price: numerical.NewFromFloat(50020), Quantity: numerical.NewFromFloat(2.5)},
						{Price: numerical.NewFromFloat(50030), Quantity: numerical.NewFromFloat(3.5)},
					},
				}

				// Update the store
				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)

				// Retrieve and verify
				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Asset).To(Equal(btc))
				Expect(retrieved.Bids).To(HaveLen(3))
				Expect(retrieved.Asks).To(HaveLen(3))
				Expect(retrieved.Bids[0].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(retrieved.Asks[0].Price).To(Equal(numerical.NewFromFloat(50010)))
			})

			It("should handle multiple instrument types for the same asset and exchange", func() {
				now := time.Now()

				orderBookPerp := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)},
					},
				}

				orderBookSpot := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50005), Quantity: numerical.NewFromFloat(2.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50015), Quantity: numerical.NewFromFloat(1.5)},
					},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBookPerp)
				store.UpdateOrderBook(btc, "hyperliquid", connector.TypeSpot, orderBookSpot)

				perpBook := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				spotBook := store.GetOrderBook(btc, "hyperliquid", connector.TypeSpot)

				Expect(perpBook).NotTo(BeNil())
				Expect(spotBook).NotTo(BeNil())
				Expect(perpBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(spotBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(50005)))
			})

			It("should handle multiple exchanges for the same asset", func() {
				now := time.Now()

				orderBookHyper := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)},
					},
				}

				orderBookBybit := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50100), Quantity: numerical.NewFromFloat(2.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50110), Quantity: numerical.NewFromFloat(1.5)},
					},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBookHyper)
				store.UpdateOrderBook(btc, "bybit", connector.TypePerpetual, orderBookBybit)

				hyperBook := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				bybitBook := store.GetOrderBook(btc, "bybit", connector.TypePerpetual)

				Expect(hyperBook).NotTo(BeNil())
				Expect(bybitBook).NotTo(BeNil())
				Expect(hyperBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(bybitBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(50100)))
			})

			It("should handle multiple assets", func() {
				now := time.Now()

				orderBookBTC := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)},
					},
				}

				orderBookETH := connector.OrderBook{
					Asset:     eth,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(3000), Quantity: numerical.NewFromFloat(5.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(3010), Quantity: numerical.NewFromFloat(4.0)},
					},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBookBTC)
				store.UpdateOrderBook(eth, "hyperliquid", connector.TypePerpetual, orderBookETH)

				btcBook := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				ethBook := store.GetOrderBook(eth, "hyperliquid", connector.TypePerpetual)

				Expect(btcBook).NotTo(BeNil())
				Expect(ethBook).NotTo(BeNil())
				Expect(btcBook.Asset).To(Equal(btc))
				Expect(ethBook.Asset).To(Equal(eth))
				Expect(btcBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(50000)))
				Expect(ethBook.Bids[0].Price).To(Equal(numerical.NewFromFloat(3000)))
			})
		})

		Context("when updating an existing orderbook", func() {
			It("should replace the orderbook with new data", func() {
				now := time.Now()

				// First orderbook
				orderBook1 := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
						{Price: numerical.NewFromFloat(49990), Quantity: numerical.NewFromFloat(2.0)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)},
						{Price: numerical.NewFromFloat(50020), Quantity: numerical.NewFromFloat(2.5)},
					},
				}

				// Updated orderbook with different prices/quantities
				orderBook2 := connector.OrderBook{
					Asset:     btc,
					Timestamp: now.Add(time.Second),
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50005), Quantity: numerical.NewFromFloat(2.0)},
						{Price: numerical.NewFromFloat(49995), Quantity: numerical.NewFromFloat(3.0)},
						{Price: numerical.NewFromFloat(49985), Quantity: numerical.NewFromFloat(1.5)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50015), Quantity: numerical.NewFromFloat(1.5)},
					},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook1)
				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook2)

				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)

				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Bids).To(HaveLen(3), "Should have updated bid levels")
				Expect(retrieved.Asks).To(HaveLen(1), "Should have updated ask levels")
				Expect(retrieved.Bids[0].Price).To(Equal(numerical.NewFromFloat(50005)), "Best bid should be updated")
				Expect(retrieved.Bids[0].Quantity).To(Equal(numerical.NewFromFloat(2.0)), "Best bid quantity should be updated")
				Expect(retrieved.Asks[0].Price).To(Equal(numerical.NewFromFloat(50015)), "Best ask should be updated")
			})
		})

		Context("when dealing with empty orderbooks", func() {
			It("should handle orderbooks with empty bids and asks", func() {
				now := time.Now()

				orderBook := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{},
					Asks:      []connector.PriceLevel{},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)

				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Bids).To(HaveLen(0))
				Expect(retrieved.Asks).To(HaveLen(0))
			})
		})

		Context("when dealing with concurrent updates", func() {
			It("should handle concurrent updates without data loss", func() {
				now := time.Now()
				done := make(chan bool)

				// Spawn 10 goroutines updating different assets/exchanges
				for i := 0; i < 10; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						orderBook := connector.OrderBook{
							Asset:     btc,
							Timestamp: now,
							Bids: []connector.PriceLevel{
								{Price: numerical.NewFromFloat(50000 + float64(idx*10)), Quantity: numerical.NewFromFloat(1.0)},
							},
							Asks: []connector.PriceLevel{
								{Price: numerical.NewFromFloat(50100 + float64(idx*10)), Quantity: numerical.NewFromFloat(1.0)},
							},
						}
						store.UpdateOrderBook(btc, connector.ExchangeName("exchange"+string(rune(idx))), connector.TypePerpetual, orderBook)
						done <- true
					}(i)
				}

				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}

				// Verify we can retrieve orderbooks for all exchanges
				books := store.GetOrderBooks(btc)
				Expect(books).To(HaveLen(10))
			})
		})
	})

	Describe("GetOrderBooks", func() {
		Context("when orderbooks exist for an asset", func() {
			It("should return all orderbooks for that asset", func() {
				now := time.Now()

				orderBook1 := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)}},
				}

				orderBook2 := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50100), Quantity: numerical.NewFromFloat(2.0)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50110), Quantity: numerical.NewFromFloat(1.5)}},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook1)
				store.UpdateOrderBook(btc, "bybit", connector.TypePerpetual, orderBook2)

				books := store.GetOrderBooks(btc)
				Expect(books).To(HaveLen(2))
				Expect(books["hyperliquid"]).NotTo(BeNil())
				Expect(books["bybit"]).NotTo(BeNil())
			})
		})

		Context("when no orderbooks exist for an asset", func() {
			It("should return an empty map", func() {
				books := store.GetOrderBooks(btc)
				Expect(books).To(HaveLen(0))
			})
		})
	})

	Describe("GetOrderBook", func() {
		Context("when orderbook exists", func() {
			It("should return the correct orderbook", func() {
				now := time.Now()

				orderBook := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)}},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)

				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Asset).To(Equal(btc))
			})
		})

		Context("when orderbook does not exist", func() {
			It("should return nil for non-existent asset", func() {
				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
				Expect(retrieved).To(BeNil())
			})

			It("should return nil for non-existent exchange", func() {
				now := time.Now()
				orderBook := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)}},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)

				retrieved := store.GetOrderBook(btc, "bybit", connector.TypePerpetual)
				Expect(retrieved).To(BeNil())
			})

			It("should return nil for non-existent instrument type", func() {
				now := time.Now()
				orderBook := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)}},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)

				retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypeSpot)
				Expect(retrieved).To(BeNil())
			})
		})
	})

	Describe("GetAllAssetsWithOrderBooks", func() {
		Context("when orderbooks exist", func() {
			It("should return all assets with orderbooks", func() {
				now := time.Now()

				orderBookBTC := connector.OrderBook{
					Asset:     btc,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(50010), Quantity: numerical.NewFromFloat(1.0)}},
				}

				orderBookETH := connector.OrderBook{
					Asset:     eth,
					Timestamp: now,
					Bids:      []connector.PriceLevel{{Price: numerical.NewFromFloat(3000), Quantity: numerical.NewFromFloat(5.0)}},
					Asks:      []connector.PriceLevel{{Price: numerical.NewFromFloat(3010), Quantity: numerical.NewFromFloat(4.0)}},
				}

				store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBookBTC)
				store.UpdateOrderBook(eth, "hyperliquid", connector.TypePerpetual, orderBookETH)

				assets := store.GetAllAssetsWithOrderBooks()
				Expect(assets).To(HaveLen(2))
				Expect(assets).To(ContainElement(btc))
				Expect(assets).To(ContainElement(eth))
			})
		})

		Context("when no orderbooks exist", func() {
			It("should return an empty slice", func() {
				assets := store.GetAllAssetsWithOrderBooks()
				Expect(assets).To(HaveLen(0))
			})
		})
	})

	Describe("Thread safety and data integrity", func() {
		It("should maintain data integrity under high concurrent load", func() {
			now := time.Now()
			iterations := 100
			done := make(chan bool, iterations*2)

			// Concurrent writers
			for i := 0; i < iterations; i++ {
				go func(idx int) {
					defer GinkgoRecover()
					orderBook := connector.OrderBook{
						Asset:     btc,
						Timestamp: now,
						Bids: []connector.PriceLevel{
							{Price: numerical.NewFromFloat(50000 + float64(idx)), Quantity: numerical.NewFromFloat(float64(idx))},
						},
						Asks: []connector.PriceLevel{
							{Price: numerical.NewFromFloat(50100 + float64(idx)), Quantity: numerical.NewFromFloat(float64(idx))},
						},
					}
					store.UpdateOrderBook(btc, "hyperliquid", connector.TypePerpetual, orderBook)
					done <- true
				}(i)
			}

			// Concurrent readers
			for i := 0; i < iterations; i++ {
				go func() {
					defer GinkgoRecover()
					_ = store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
					_ = store.GetOrderBooks(btc)
					_ = store.GetAllAssetsWithOrderBooks()
					done <- true
				}()
			}

			// Wait for all operations to complete
			for i := 0; i < iterations*2; i++ {
				<-done
			}

			// Verify final state is valid
			retrieved := store.GetOrderBook(btc, "hyperliquid", connector.TypePerpetual)
			Expect(retrieved).NotTo(BeNil())
			Expect(retrieved.Asset).To(Equal(btc))
		})
	})
})
