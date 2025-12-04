package trade_test

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/stores/activity/trade"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	activityTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trade Store - Add Trades", func() {
	var (
		store activityTypes.Trades
	)

	BeforeEach(func() {
		store = trade.NewStore()
	})

	Describe("AddTrade", func() {
		Context("when adding a new trade", func() {
			It("should store the trade correctly", func() {
				t := connector.Trade{
					ID:        "trade-1",
					OrderID:   "order-1",
					Symbol:    "BTC",
					Exchange:  "hyperliquid",
					Price:     numerical.NewFromFloat(50000),
					Quantity:  numerical.NewFromFloat(1.0),
					Side:      connector.OrderSideBuy,
					IsMaker:   false,
					Fee:       numerical.NewFromFloat(10),
					Timestamp: time.Now(),
				}

				store.AddTrade(t)

				Expect(store.GetTradeCount()).To(Equal(1))
				retrieved := store.GetTradeByID("trade-1")
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Symbol).To(Equal("BTC"))
				Expect(retrieved.Price.Equal(numerical.NewFromFloat(50000))).To(BeTrue())
			})

			It("should handle multiple trades", func() {
				for i := 1; i <= 5; i++ {
					t := connector.Trade{
						ID:       "trade-" + string(rune('0'+i)),
						Symbol:   "BTC",
						Exchange: "hyperliquid",
					}
					store.AddTrade(t)
				}

				Expect(store.GetTradeCount()).To(Equal(5))
			})

			It("should skip duplicate trades", func() {
				t := connector.Trade{
					ID:     "trade-1",
					Symbol: "BTC",
				}

				store.AddTrade(t)
				store.AddTrade(t) // Same ID

				Expect(store.GetTradeCount()).To(Equal(1))
			})

			It("should preserve insertion order", func() {
				now := time.Now()
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC", Timestamp: now},
					{ID: "t2", Symbol: "ETH", Timestamp: now.Add(time.Second)},
					{ID: "t3", Symbol: "SOL", Timestamp: now.Add(2 * time.Second)},
				}

				for _, t := range trades {
					store.AddTrade(t)
				}

				all := store.GetAllTrades()
				Expect(all).To(HaveLen(3))
				Expect(all[0].ID).To(Equal("t1"))
				Expect(all[1].ID).To(Equal("t2"))
				Expect(all[2].ID).To(Equal("t3"))
			})
		})
	})

	Describe("AddTrades", func() {
		Context("when adding multiple trades at once", func() {
			It("should store all trades correctly", func() {
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC", Exchange: "hyperliquid"},
					{ID: "t2", Symbol: "ETH", Exchange: "hyperliquid"},
					{ID: "t3", Symbol: "SOL", Exchange: "bybit"},
				}

				store.AddTrades(trades)

				Expect(store.GetTradeCount()).To(Equal(3))
				Expect(store.TradeExists("t1")).To(BeTrue())
				Expect(store.TradeExists("t2")).To(BeTrue())
				Expect(store.TradeExists("t3")).To(BeTrue())
			})

			It("should skip duplicate trades in batch against existing trades", func() {
				// Add initial trade
				store.AddTrade(connector.Trade{ID: "t1", Symbol: "BTC"})

				// Add batch with duplicate of existing trade
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC"}, // Duplicate of existing
					{ID: "t2", Symbol: "ETH"},
				}

				store.AddTrades(trades)

				Expect(store.GetTradeCount()).To(Equal(2))
			})

			It("should handle empty slice", func() {
				store.AddTrades([]connector.Trade{})

				Expect(store.GetTradeCount()).To(Equal(0))
			})

			It("should handle all duplicates", func() {
				store.AddTrade(connector.Trade{ID: "t1", Symbol: "BTC"})
				store.AddTrade(connector.Trade{ID: "t2", Symbol: "ETH"})

				// Add batch with all existing trades
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC"},
					{ID: "t2", Symbol: "ETH"},
				}

				store.AddTrades(trades)

				Expect(store.GetTradeCount()).To(Equal(2))
			})
		})
	})

	Describe("Clear", func() {
		Context("when clearing the store", func() {
			It("should remove all trades", func() {
				store.AddTrade(connector.Trade{ID: "t1"})
				store.AddTrade(connector.Trade{ID: "t2"})

				store.Clear()

				Expect(store.GetTradeCount()).To(Equal(0))
				Expect(store.GetAllTrades()).To(BeEmpty())
			})

			It("should reset ID lookup", func() {
				store.AddTrade(connector.Trade{ID: "t1"})

				store.Clear()

				Expect(store.TradeExists("t1")).To(BeFalse())
				Expect(store.GetTradeByID("t1")).To(BeNil())
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("when multiple goroutines add trades", func() {
			It("should handle concurrent writes safely", func() {
				done := make(chan bool)
				iterations := 50

				// Writer 1
				go func() {
					for i := 0; i < iterations; i++ {
						store.AddTrade(connector.Trade{
							ID:     "btc-" + string(rune(i)),
							Symbol: "BTC",
						})
					}
					done <- true
				}()

				// Writer 2
				go func() {
					for i := 0; i < iterations; i++ {
						store.AddTrade(connector.Trade{
							ID:     "eth-" + string(rune(i)),
							Symbol: "ETH",
						})
					}
					done <- true
				}()

				// Reader
				go func() {
					for i := 0; i < iterations; i++ {
						_ = store.GetAllTrades()
						_ = store.GetTradeCount()
					}
					done <- true
				}()

				<-done
				<-done
				<-done

				Expect(store.GetTradeCount()).To(Equal(iterations * 2))
			})
		})
	})
})
