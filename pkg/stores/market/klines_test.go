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

var _ = Describe("Market Data Store - Klines", func() {
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

	Describe("UpdateKline", func() {
		Context("when adding a new kline", func() {
			It("should store the kline correctly", func() {
				// Create a kline
				now := time.Now().Truncate(time.Minute)
				kline := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					High:      numerical.NewFromFloat(50100),
					Low:       numerical.NewFromFloat(49900),
					Close:     numerical.NewFromFloat(50050),
					Volume:    numerical.NewFromFloat(100),
				}

				// Update the store
				store.UpdateKline(btc, "hyperliquid", kline)

				// Retrieve and verify
				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)
				Expect(klines).To(HaveLen(1))
				Expect(klines[0].Symbol).To(Equal("BTC"))
				Expect(klines[0].Open).To(Equal(numerical.NewFromFloat(50000)))
				Expect(klines[0].Close).To(Equal(numerical.NewFromFloat(50050)))
			})

			It("should handle multiple intervals for the same asset and exchange", func() {
				now := time.Now().Truncate(time.Minute)

				kline1m := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
				}

				kline5m := connector.Kline{
					Symbol:    "BTC",
					Interval:  "5m",
					OpenTime:  now,
					CloseTime: now.Add(5 * time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50200),
				}

				store.UpdateKline(btc, "hyperliquid", kline1m)
				store.UpdateKline(btc, "hyperliquid", kline5m)

				klines1m := store.GetKlines(btc, "hyperliquid", "1m", 0)
				klines5m := store.GetKlines(btc, "hyperliquid", "5m", 0)

				Expect(klines1m).To(HaveLen(1))
				Expect(klines5m).To(HaveLen(1))
				Expect(klines1m[0].Interval).To(Equal("1m"))
				Expect(klines5m[0].Interval).To(Equal("5m"))
			})

			It("should handle multiple exchanges for the same asset", func() {
				now := time.Now().Truncate(time.Minute)

				klineHyper := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
				}

				klineBybit := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50100),
					Close:     numerical.NewFromFloat(50150),
				}

				store.UpdateKline(btc, "hyperliquid", klineHyper)
				store.UpdateKline(btc, "bybit", klineBybit)

				klinesHyper := store.GetKlines(btc, "hyperliquid", "1m", 0)
				klinesBybit := store.GetKlines(btc, "bybit", "1m", 0)

				Expect(klinesHyper).To(HaveLen(1))
				Expect(klinesBybit).To(HaveLen(1))
				Expect(klinesHyper[0].Open).To(Equal(numerical.NewFromFloat(50000)))
				Expect(klinesBybit[0].Open).To(Equal(numerical.NewFromFloat(50100)))
			})

			It("should handle multiple assets", func() {
				now := time.Now().Truncate(time.Minute)

				klineBTC := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
				}

				klineETH := connector.Kline{
					Symbol:    "ETH",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(3000),
					Close:     numerical.NewFromFloat(3010),
				}

				store.UpdateKline(btc, "hyperliquid", klineBTC)
				store.UpdateKline(eth, "hyperliquid", klineETH)

				klinesBTC := store.GetKlines(btc, "hyperliquid", "1m", 0)
				klinesETH := store.GetKlines(eth, "hyperliquid", "1m", 0)

				Expect(klinesBTC).To(HaveLen(1))
				Expect(klinesETH).To(HaveLen(1))
				Expect(klinesBTC[0].Symbol).To(Equal("BTC"))
				Expect(klinesETH[0].Symbol).To(Equal("ETH"))
			})
		})

		Context("when updating an existing kline", func() {
			It("should replace the kline with the same open time", func() {
				now := time.Now().Truncate(time.Minute)

				// First kline
				kline1 := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
					High:      numerical.NewFromFloat(50100),
					Low:       numerical.NewFromFloat(49900),
				}

				// Updated kline with same open time
				kline2 := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now, // Same open time
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50150), // Different close
					High:      numerical.NewFromFloat(50200), // Different high
					Low:       numerical.NewFromFloat(49800), // Different low
				}

				store.UpdateKline(btc, "hyperliquid", kline1)
				store.UpdateKline(btc, "hyperliquid", kline2)

				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)

				Expect(klines).To(HaveLen(1), "Should only have one kline after update")
				Expect(klines[0].Close).To(Equal(numerical.NewFromFloat(50150)), "Close should be updated")
				Expect(klines[0].High).To(Equal(numerical.NewFromFloat(50200)), "High should be updated")
				Expect(klines[0].Low).To(Equal(numerical.NewFromFloat(49800)), "Low should be updated")
			})
		})

		Context("when adding multiple sequential klines", func() {
			It("should store them in order", func() {
				now := time.Now().Truncate(time.Minute)

				// Add 5 sequential klines
				for i := 0; i < 5; i++ {
					kline := connector.Kline{
						Symbol:    "BTC",
						Interval:  "1m",
						OpenTime:  now.Add(time.Duration(i) * time.Minute),
						CloseTime: now.Add(time.Duration(i+1) * time.Minute),
						Open:      numerical.NewFromFloat(50000 + float64(i*10)),
						Close:     numerical.NewFromFloat(50010 + float64(i*10)),
					}
					store.UpdateKline(btc, "hyperliquid", kline)
				}

				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)
				Expect(klines).To(HaveLen(5))

				// Verify they're in order by checking open prices
				for i := 0; i < 5; i++ {
					expectedOpen := numerical.NewFromFloat(50000 + float64(i*10))
					Expect(klines[i].Open).To(Equal(expectedOpen))
				}
			})
		})

		Context("when dealing with concurrent updates", func() {
			It("should handle concurrent updates without data loss", func() {
				now := time.Now().Truncate(time.Minute)
				done := make(chan bool)

				// Spawn 10 goroutines updating different intervals
				for i := 0; i < 10; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						kline := connector.Kline{
							Symbol:    "BTC",
							Interval:  "1m",
							OpenTime:  now.Add(time.Duration(idx) * time.Minute),
							CloseTime: now.Add(time.Duration(idx+1) * time.Minute),
							Open:      numerical.NewFromFloat(50000 + float64(idx*10)),
							Close:     numerical.NewFromFloat(50010 + float64(idx*10)),
						}
						store.UpdateKline(btc, "hyperliquid", kline)
						done <- true
					}(i)
				}

				// Wait for all goroutines
				for i := 0; i < 10; i++ {
					<-done
				}

				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)
				Expect(klines).To(HaveLen(10), "All klines should be stored")
			})
		})
	})

	Describe("GetKlines", func() {
		Context("with limit parameter", func() {
			It("should return the last N klines when limit is set", func() {
				now := time.Now().Truncate(time.Minute)

				// Add 10 klines
				for i := 0; i < 10; i++ {
					kline := connector.Kline{
						Symbol:    "BTC",
						Interval:  "1m",
						OpenTime:  now.Add(time.Duration(i) * time.Minute),
						CloseTime: now.Add(time.Duration(i+1) * time.Minute),
						Open:      numerical.NewFromFloat(50000 + float64(i)),
						Close:     numerical.NewFromFloat(50001 + float64(i)),
					}
					store.UpdateKline(btc, "hyperliquid", kline)
				}

				// Get last 5
				klines := store.GetKlines(btc, "hyperliquid", "1m", 5)
				Expect(klines).To(HaveLen(5))

				// Should be the last 5 (indices 5-9)
				Expect(klines[0].Open).To(Equal(numerical.NewFromFloat(50005)))
				Expect(klines[4].Open).To(Equal(numerical.NewFromFloat(50009)))
			})

			It("should return all klines when limit is 0", func() {
				now := time.Now().Truncate(time.Minute)

				for i := 0; i < 10; i++ {
					kline := connector.Kline{
						Symbol:    "BTC",
						Interval:  "1m",
						OpenTime:  now.Add(time.Duration(i) * time.Minute),
						CloseTime: now.Add(time.Duration(i+1) * time.Minute),
						Open:      numerical.NewFromFloat(50000 + float64(i)),
						Close:     numerical.NewFromFloat(50001 + float64(i)),
					}
					store.UpdateKline(btc, "hyperliquid", kline)
				}

				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)
				Expect(klines).To(HaveLen(10))
			})
		})

		Context("when no klines exist", func() {
			It("should return empty slice for non-existent asset", func() {
				klines := store.GetKlines(btc, "hyperliquid", "1m", 0)
				Expect(klines).To(BeEmpty())
			})

			It("should return empty slice for non-existent exchange", func() {
				now := time.Now().Truncate(time.Minute)
				kline := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
				}
				store.UpdateKline(btc, "hyperliquid", kline)

				klines := store.GetKlines(btc, "bybit", "1m", 0)
				Expect(klines).To(BeEmpty())
			})

			It("should return empty slice for non-existent interval", func() {
				now := time.Now().Truncate(time.Minute)
				kline := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now,
					CloseTime: now.Add(time.Minute),
					Open:      numerical.NewFromFloat(50000),
					Close:     numerical.NewFromFloat(50050),
				}
				store.UpdateKline(btc, "hyperliquid", kline)

				klines := store.GetKlines(btc, "hyperliquid", "5m", 0)
				Expect(klines).To(BeEmpty())
			})
		})
	})

	Describe("GetKlinesSince", func() {
		var now time.Time

		BeforeEach(func() {
			now = time.Now().Truncate(time.Minute)

			// Add klines at different times
			for i := 0; i < 10; i++ {
				kline := connector.Kline{
					Symbol:    "BTC",
					Interval:  "1m",
					OpenTime:  now.Add(time.Duration(i) * time.Minute),
					CloseTime: now.Add(time.Duration(i+1) * time.Minute),
					Open:      numerical.NewFromFloat(50000 + float64(i)),
					Close:     numerical.NewFromFloat(50001 + float64(i)),
				}
				store.UpdateKline(btc, "hyperliquid", kline)
			}
		})

		It("should return klines after the specified time", func() {
			since := now.Add(5 * time.Minute)
			klines := store.GetKlinesSince(btc, "hyperliquid", "1m", since)

			Expect(klines).To(HaveLen(5), "Should return klines from minute 5-9")
			Expect(klines[0].Open).To(Equal(numerical.NewFromFloat(50005)))
		})

		It("should include klines at exactly the specified time", func() {
			since := now.Add(5 * time.Minute)
			klines := store.GetKlinesSince(btc, "hyperliquid", "1m", since)

			Expect(klines[0].OpenTime).To(Equal(since))
		})

		It("should return empty slice when since is after all klines", func() {
			since := now.Add(20 * time.Minute)
			klines := store.GetKlinesSince(btc, "hyperliquid", "1m", since)

			Expect(klines).To(BeEmpty())
		})

		It("should return all klines when since is before all klines", func() {
			since := now.Add(-1 * time.Minute)
			klines := store.GetKlinesSince(btc, "hyperliquid", "1m", since)

			Expect(klines).To(HaveLen(10))
		})

		Context("when no klines exist", func() {
			It("should return empty slice", func() {
				klines := store.GetKlinesSince(eth, "hyperliquid", "1m", now)
				Expect(klines).To(BeEmpty())
			})
		})
	})
})
