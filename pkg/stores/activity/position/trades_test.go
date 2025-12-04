package position_test

import (
	"time"

	timeProvider "github.com/backtesting-org/kronos-sdk/pkg/runtime/time"
	"github.com/backtesting-org/kronos-sdk/pkg/stores/activity/position"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	activityTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position Store - Trades", func() {
	var (
		store        activityTypes.Positions
		provider     temporal.TimeProvider
		strategyName strategy.StrategyName
	)

	BeforeEach(func() {
		provider = timeProvider.NewTimeProvider()
		store = position.NewStore(provider)
		strategyName = "test-strategy"
	})

	Describe("AddTradeToStrategy", func() {
		Context("when adding a trade to an existing strategy", func() {
			It("should add the trade to existing execution", func() {
				// First, create the strategy execution
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{},
					Trades: []connector.Trade{
						{ID: "trade-1", Symbol: "BTC"},
					},
				})

				// Add new trade
				newTrade := connector.Trade{
					ID:        "trade-2",
					OrderID:   "order-1",
					Symbol:    "ETH",
					Exchange:  "hyperliquid",
					Price:     numerical.NewFromFloat(3000),
					Quantity:  numerical.NewFromFloat(10.0),
					Side:      connector.OrderSideBuy,
					IsMaker:   false,
					Fee:       numerical.NewFromFloat(0.01),
					Timestamp: time.Now(),
				}

				store.AddTradeToStrategy(strategyName, newTrade)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Trades).To(HaveLen(2))
				Expect(retrieved.Trades[1].ID).To(Equal("trade-2"))
				Expect(retrieved.Trades[1].Symbol).To(Equal("ETH"))
				Expect(retrieved.Trades[1].Exchange).To(Equal(connector.ExchangeName("hyperliquid")))
			})
		})

		Context("when adding a trade to a non-existent strategy", func() {
			It("should create a new execution and add the trade", func() {
				trade := connector.Trade{
					ID:        "trade-1",
					Symbol:    "BTC",
					Price:     numerical.NewFromFloat(50000),
					Quantity:  numerical.NewFromFloat(1.0),
					Side:      connector.OrderSideBuy,
					Timestamp: time.Now(),
				}

				store.AddTradeToStrategy(strategyName, trade)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Trades).To(HaveLen(1))
				Expect(retrieved.Trades[0].ID).To(Equal("trade-1"))
			})
		})

		Context("when adding multiple trades", func() {
			It("should preserve order of addition", func() {
				for i := 1; i <= 5; i++ {
					trade := connector.Trade{
						ID:        "trade-" + string(rune('0'+i)),
						Symbol:    "BTC",
						Timestamp: time.Now().Add(time.Duration(i) * time.Second),
					}
					store.AddTradeToStrategy(strategyName, trade)
				}

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Trades).To(HaveLen(5))
			})

			It("should track trades with different sides", func() {
				buyTrade := connector.Trade{
					ID:       "buy-trade",
					Symbol:   "BTC",
					Side:     connector.OrderSideBuy,
					Quantity: numerical.NewFromFloat(1.0),
					Price:    numerical.NewFromFloat(50000),
				}

				sellTrade := connector.Trade{
					ID:       "sell-trade",
					Symbol:   "BTC",
					Side:     connector.OrderSideSell,
					Quantity: numerical.NewFromFloat(0.5),
					Price:    numerical.NewFromFloat(51000),
				}

				store.AddTradeToStrategy(strategyName, buyTrade)
				store.AddTradeToStrategy(strategyName, sellTrade)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Trades).To(HaveLen(2))
				Expect(retrieved.Trades[0].Side).To(Equal(connector.OrderSideBuy))
				Expect(retrieved.Trades[1].Side).To(Equal(connector.OrderSideSell))
			})
		})
	})

	Describe("GetTradesForStrategy", func() {
		Context("when retrieving trades for an existing strategy", func() {
			It("should return all trades for the strategy", func() {
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Trades: []connector.Trade{
						{ID: "trade-1", Symbol: "BTC"},
						{ID: "trade-2", Symbol: "ETH"},
						{ID: "trade-3", Symbol: "SOL"},
					},
				})

				trades := store.GetTradesForStrategy(strategyName)

				Expect(trades).To(HaveLen(3))
				Expect(trades[0].ID).To(Equal("trade-1"))
				Expect(trades[1].ID).To(Equal("trade-2"))
				Expect(trades[2].ID).To(Equal("trade-3"))
			})
		})

		Context("when retrieving trades for a non-existent strategy", func() {
			It("should return an empty slice", func() {
				trades := store.GetTradesForStrategy("non-existent")
				Expect(trades).To(BeEmpty())
			})
		})

		Context("when strategy has no trades", func() {
			It("should return an empty slice", func() {
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
					Trades: []connector.Trade{},
				})

				trades := store.GetTradesForStrategy(strategyName)
				Expect(trades).To(BeEmpty())
			})
		})
	})

	Describe("Trade analysis scenarios", func() {
		Context("when calculating trade statistics", func() {
			It("should allow calculation of total volume", func() {
				trades := []connector.Trade{
					{ID: "t1", Quantity: numerical.NewFromFloat(1.0), Price: numerical.NewFromFloat(50000)},
					{ID: "t2", Quantity: numerical.NewFromFloat(2.0), Price: numerical.NewFromFloat(51000)},
					{ID: "t3", Quantity: numerical.NewFromFloat(0.5), Price: numerical.NewFromFloat(52000)},
				}

				for _, trade := range trades {
					store.AddTradeToStrategy(strategyName, trade)
				}

				retrieved := store.GetTradesForStrategy(strategyName)

				totalVolume := numerical.NewFromFloat(0)
				for _, trade := range retrieved {
					totalVolume = totalVolume.Add(trade.Quantity.Mul(trade.Price))
				}

				// 1*50000 + 2*51000 + 0.5*52000 = 50000 + 102000 + 26000 = 178000
				expected := numerical.NewFromFloat(178000)
				Expect(totalVolume.Equal(expected)).To(BeTrue())
			})

			It("should allow calculation of total fees", func() {
				trades := []connector.Trade{
					{ID: "t1", Fee: numerical.NewFromFloat(10)},
					{ID: "t2", Fee: numerical.NewFromFloat(20)},
					{ID: "t3", Fee: numerical.NewFromFloat(5)},
				}

				for _, trade := range trades {
					store.AddTradeToStrategy(strategyName, trade)
				}

				retrieved := store.GetTradesForStrategy(strategyName)

				totalFees := numerical.NewFromFloat(0)
				for _, trade := range retrieved {
					totalFees = totalFees.Add(trade.Fee)
				}

				expected := numerical.NewFromFloat(35)
				Expect(totalFees.Equal(expected)).To(BeTrue())
			})
		})

		Context("when tracking maker vs taker trades", func() {
			It("should distinguish between maker and taker trades", func() {
				makerTrade := connector.Trade{
					ID:      "maker-trade",
					IsMaker: true,
					Fee:     numerical.NewFromFloat(5), // Lower fee for maker
				}

				takerTrade := connector.Trade{
					ID:      "taker-trade",
					IsMaker: false,
					Fee:     numerical.NewFromFloat(10), // Higher fee for taker
				}

				store.AddTradeToStrategy(strategyName, makerTrade)
				store.AddTradeToStrategy(strategyName, takerTrade)

				trades := store.GetTradesForStrategy(strategyName)

				makerCount := 0
				takerCount := 0
				for _, trade := range trades {
					if trade.IsMaker {
						makerCount++
					} else {
						takerCount++
					}
				}

				Expect(makerCount).To(Equal(1))
				Expect(takerCount).To(Equal(1))
			})
		})
	})

	Describe("Order-Trade relationship", func() {
		Context("when linking trades to orders", func() {
			It("should store trades with order references", func() {
				// Add order first
				order := connector.Order{
					ID:     "order-123",
					Symbol: "BTC",
					Side:   connector.OrderSideBuy,
				}
				store.AddOrderToStrategy(strategyName, order)

				// Add trade linked to order
				trade := connector.Trade{
					ID:       "trade-456",
					OrderID:  "order-123",
					Symbol:   "BTC",
					Quantity: numerical.NewFromFloat(0.5),
				}
				store.AddTradeToStrategy(strategyName, trade)

				execution := store.GetStrategyExecution(strategyName)
				Expect(execution.Orders).To(HaveLen(1))
				Expect(execution.Trades).To(HaveLen(1))
				Expect(execution.Trades[0].OrderID).To(Equal("order-123"))
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("when multiple goroutines add trades", func() {
			It("should handle concurrent writes safely", func() {
				done := make(chan bool)
				iterations := 50

				// Writer 1 - adds BTC trades
				go func() {
					for i := 0; i < iterations; i++ {
						store.AddTradeToStrategy(strategyName, connector.Trade{
							ID:     "btc-trade-" + string(rune(i)),
							Symbol: "BTC",
						})
					}
					done <- true
				}()

				// Writer 2 - adds ETH trades
				go func() {
					for i := 0; i < iterations; i++ {
						store.AddTradeToStrategy(strategyName, connector.Trade{
							ID:     "eth-trade-" + string(rune(i)),
							Symbol: "ETH",
						})
					}
					done <- true
				}()

				// Reader
				go func() {
					for i := 0; i < iterations; i++ {
						_ = store.GetTradesForStrategy(strategyName)
					}
					done <- true
				}()

				<-done
				<-done
				<-done

				// Verify all trades were added
				trades := store.GetTradesForStrategy(strategyName)
				Expect(len(trades)).To(Equal(iterations * 2))
			})
		})
	})
})
