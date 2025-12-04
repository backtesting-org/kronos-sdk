package trade_test

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/stores/activity/trade"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	activityTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trade Store - Queries", func() {
	var (
		store activityTypes.Trades
	)

	BeforeEach(func() {
		store = trade.NewStore()
	})

	Describe("GetAllTrades", func() {
		Context("when store has trades", func() {
			It("should return all trades", func() {
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC"},
					{ID: "t2", Symbol: "ETH"},
					{ID: "t3", Symbol: "SOL"},
				}
				store.AddTrades(trades)

				all := store.GetAllTrades()

				Expect(all).To(HaveLen(3))
			})
		})

		Context("when store is empty", func() {
			It("should return empty slice", func() {
				all := store.GetAllTrades()
				Expect(all).To(BeEmpty())
			})
		})
	})

	Describe("GetTradesByExchange", func() {
		BeforeEach(func() {
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Exchange: "hyperliquid"},
				{ID: "t2", Symbol: "ETH", Exchange: "hyperliquid"},
				{ID: "t3", Symbol: "BTC", Exchange: "bybit"},
				{ID: "t4", Symbol: "SOL", Exchange: "paradex"},
			}
			store.AddTrades(trades)
		})

		Context("when filtering by exchange", func() {
			It("should return trades for hyperliquid", func() {
				trades := store.GetTradesByExchange("hyperliquid")

				Expect(trades).To(HaveLen(2))
				for _, t := range trades {
					Expect(t.Exchange).To(Equal(connector.ExchangeName("hyperliquid")))
				}
			})

			It("should return trades for bybit", func() {
				trades := store.GetTradesByExchange("bybit")

				Expect(trades).To(HaveLen(1))
				Expect(trades[0].ID).To(Equal("t3"))
			})

			It("should return empty for unknown exchange", func() {
				trades := store.GetTradesByExchange("unknown")
				Expect(trades).To(BeEmpty())
			})
		})
	})

	Describe("GetTradesByAsset", func() {
		BeforeEach(func() {
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Exchange: "hyperliquid"},
				{ID: "t2", Symbol: "BTC", Exchange: "bybit"},
				{ID: "t3", Symbol: "ETH", Exchange: "hyperliquid"},
				{ID: "t4", Symbol: "SOL", Exchange: "paradex"},
			}
			store.AddTrades(trades)
		})

		Context("when filtering by asset", func() {
			It("should return trades for BTC", func() {
				btc := portfolio.NewAsset("BTC")
				trades := store.GetTradesByAsset(btc)

				Expect(trades).To(HaveLen(2))
				for _, t := range trades {
					Expect(t.Symbol).To(Equal("BTC"))
				}
			})

			It("should return trades for ETH", func() {
				eth := portfolio.NewAsset("ETH")
				trades := store.GetTradesByAsset(eth)

				Expect(trades).To(HaveLen(1))
				Expect(trades[0].ID).To(Equal("t3"))
			})

			It("should return empty for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				trades := store.GetTradesByAsset(unknown)
				Expect(trades).To(BeEmpty())
			})
		})
	})

	Describe("GetTradesByExchangeAndAsset", func() {
		BeforeEach(func() {
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Exchange: "hyperliquid"},
				{ID: "t2", Symbol: "BTC", Exchange: "bybit"},
				{ID: "t3", Symbol: "ETH", Exchange: "hyperliquid"},
				{ID: "t4", Symbol: "BTC", Exchange: "hyperliquid"},
			}
			store.AddTrades(trades)
		})

		Context("when filtering by exchange and asset", func() {
			It("should return BTC trades on hyperliquid", func() {
				btc := portfolio.NewAsset("BTC")
				trades := store.GetTradesByExchangeAndAsset("hyperliquid", btc)

				Expect(trades).To(HaveLen(2))
				for _, t := range trades {
					Expect(t.Symbol).To(Equal("BTC"))
					Expect(t.Exchange).To(Equal(connector.ExchangeName("hyperliquid")))
				}
			})

			It("should return BTC trades on bybit", func() {
				btc := portfolio.NewAsset("BTC")
				trades := store.GetTradesByExchangeAndAsset("bybit", btc)

				Expect(trades).To(HaveLen(1))
				Expect(trades[0].ID).To(Equal("t2"))
			})

			It("should return empty for non-matching combination", func() {
				eth := portfolio.NewAsset("ETH")
				trades := store.GetTradesByExchangeAndAsset("bybit", eth)
				Expect(trades).To(BeEmpty())
			})
		})
	})

	Describe("GetTradesSince", func() {
		var baseTime time.Time

		BeforeEach(func() {
			baseTime = time.Now().Add(-time.Hour)
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Timestamp: baseTime},
				{ID: "t2", Symbol: "ETH", Timestamp: baseTime.Add(10 * time.Minute)},
				{ID: "t3", Symbol: "SOL", Timestamp: baseTime.Add(30 * time.Minute)},
				{ID: "t4", Symbol: "AVAX", Timestamp: baseTime.Add(50 * time.Minute)},
			}
			store.AddTrades(trades)
		})

		Context("when filtering by time", func() {
			It("should return trades since specific time", func() {
				since := baseTime.Add(20 * time.Minute)
				trades := store.GetTradesSince(since)

				Expect(trades).To(HaveLen(2))
				Expect(trades[0].ID).To(Equal("t3"))
				Expect(trades[1].ID).To(Equal("t4"))
			})

			It("should include trades at exact time", func() {
				trades := store.GetTradesSince(baseTime)

				Expect(trades).To(HaveLen(4))
			})

			It("should return empty when no trades after time", func() {
				since := baseTime.Add(2 * time.Hour)
				trades := store.GetTradesSince(since)
				Expect(trades).To(BeEmpty())
			})

			It("should return all trades when time is before all trades", func() {
				since := baseTime.Add(-time.Hour)
				trades := store.GetTradesSince(since)
				Expect(trades).To(HaveLen(4))
			})
		})
	})

	Describe("GetTradeByID", func() {
		BeforeEach(func() {
			store.AddTrades([]connector.Trade{
				{ID: "t1", Symbol: "BTC", Price: numerical.NewFromFloat(50000)},
				{ID: "t2", Symbol: "ETH", Price: numerical.NewFromFloat(3000)},
			})
		})

		Context("when trade exists", func() {
			It("should return the trade", func() {
				t := store.GetTradeByID("t1")

				Expect(t).NotTo(BeNil())
				Expect(t.Symbol).To(Equal("BTC"))
				Expect(t.Price.Equal(numerical.NewFromFloat(50000))).To(BeTrue())
			})
		})

		Context("when trade does not exist", func() {
			It("should return nil", func() {
				t := store.GetTradeByID("non-existent")
				Expect(t).To(BeNil())
			})
		})
	})

	Describe("TradeExists", func() {
		BeforeEach(func() {
			store.AddTrade(connector.Trade{ID: "t1", Symbol: "BTC"})
		})

		Context("when trade exists", func() {
			It("should return true", func() {
				Expect(store.TradeExists("t1")).To(BeTrue())
			})
		})

		Context("when trade does not exist", func() {
			It("should return false", func() {
				Expect(store.TradeExists("non-existent")).To(BeFalse())
			})
		})
	})

	Describe("GetTradeCount", func() {
		Context("when store has trades", func() {
			It("should return correct count", func() {
				store.AddTrades([]connector.Trade{
					{ID: "t1"},
					{ID: "t2"},
					{ID: "t3"},
				})

				Expect(store.GetTradeCount()).To(Equal(3))
			})
		})

		Context("when store is empty", func() {
			It("should return 0", func() {
				Expect(store.GetTradeCount()).To(Equal(0))
			})
		})
	})

	Describe("GetTotalVolume", func() {
		BeforeEach(func() {
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Quantity: numerical.NewFromFloat(1.5)},
				{ID: "t2", Symbol: "BTC", Quantity: numerical.NewFromFloat(2.0)},
				{ID: "t3", Symbol: "ETH", Quantity: numerical.NewFromFloat(10.0)},
				{ID: "t4", Symbol: "BTC", Quantity: numerical.NewFromFloat(0.5)},
			}
			store.AddTrades(trades)
		})

		Context("when calculating volume", func() {
			It("should sum volume for BTC", func() {
				btc := portfolio.NewAsset("BTC")
				volume := store.GetTotalVolume(btc)

				// 1.5 + 2.0 + 0.5 = 4.0
				expected := numerical.NewFromFloat(4.0)
				Expect(volume.Equal(expected)).To(BeTrue())
			})

			It("should sum volume for ETH", func() {
				eth := portfolio.NewAsset("ETH")
				volume := store.GetTotalVolume(eth)

				expected := numerical.NewFromFloat(10.0)
				Expect(volume.Equal(expected)).To(BeTrue())
			})

			It("should return 0 for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				volume := store.GetTotalVolume(unknown)

				Expect(volume.IsZero()).To(BeTrue())
			})
		})
	})

	Describe("Combined queries", func() {
		Context("when using multiple query methods together", func() {
			It("should allow complex filtering", func() {
				now := time.Now()
				trades := []connector.Trade{
					{ID: "t1", Symbol: "BTC", Exchange: "hyperliquid", Quantity: numerical.NewFromFloat(1.0), Timestamp: now.Add(-2 * time.Hour)},
					{ID: "t2", Symbol: "BTC", Exchange: "hyperliquid", Quantity: numerical.NewFromFloat(2.0), Timestamp: now.Add(-1 * time.Hour)},
					{ID: "t3", Symbol: "BTC", Exchange: "bybit", Quantity: numerical.NewFromFloat(3.0), Timestamp: now.Add(-30 * time.Minute)},
					{ID: "t4", Symbol: "ETH", Exchange: "hyperliquid", Quantity: numerical.NewFromFloat(5.0), Timestamp: now},
				}
				store.AddTrades(trades)

				// Get BTC trades on hyperliquid
				btc := portfolio.NewAsset("BTC")
				btcOnHyper := store.GetTradesByExchangeAndAsset("hyperliquid", btc)
				Expect(btcOnHyper).To(HaveLen(2))

				// Get recent trades (last hour)
				recentTrades := store.GetTradesSince(now.Add(-1 * time.Hour))
				Expect(recentTrades).To(HaveLen(3)) // t2, t3, t4

				// Total BTC volume
				btcVolume := store.GetTotalVolume(btc)
				expected := numerical.NewFromFloat(6.0) // 1 + 2 + 3
				Expect(btcVolume.Equal(expected)).To(BeTrue())
			})
		})
	})
})
