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

var _ = Describe("Market Data Store - Funding Rates", func() {
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

	Describe("UpdateFundingRate", func() {
		Context("when adding a new funding rate", func() {
			It("should store the funding rate correctly", func() {
				now := time.Now()
				fundingRate := connector.FundingRate{
					CurrentRate:     numerical.NewFromFloat(0.0001),
					NextFundingTime: now.Add(8 * time.Hour),
					Timestamp:       now,
					MarkPrice:       numerical.NewFromFloat(50000),
					IndexPrice:      numerical.NewFromFloat(49990),
					Premium:         numerical.NewFromFloat(0.0002),
				}

				store.UpdateFundingRate(btc, "hyperliquid", fundingRate)

				retrieved := store.GetFundingRate(btc, "hyperliquid")
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(retrieved.MarkPrice).To(Equal(numerical.NewFromFloat(50000)))
				Expect(retrieved.IndexPrice).To(Equal(numerical.NewFromFloat(49990)))
			})

			It("should handle multiple exchanges for the same asset", func() {
				now := time.Now()

				hyperRate := connector.FundingRate{
					CurrentRate:     numerical.NewFromFloat(0.0001),
					NextFundingTime: now.Add(8 * time.Hour),
					Timestamp:       now,
					MarkPrice:       numerical.NewFromFloat(50000),
				}

				bybitRate := connector.FundingRate{
					CurrentRate:     numerical.NewFromFloat(0.00015),
					NextFundingTime: now.Add(4 * time.Hour),
					Timestamp:       now,
					MarkPrice:       numerical.NewFromFloat(50010),
				}

				store.UpdateFundingRate(btc, "hyperliquid", hyperRate)
				store.UpdateFundingRate(btc, "bybit", bybitRate)

				hyperRetrieved := store.GetFundingRate(btc, "hyperliquid")
				bybitRetrieved := store.GetFundingRate(btc, "bybit")

				Expect(hyperRetrieved).NotTo(BeNil())
				Expect(bybitRetrieved).NotTo(BeNil())
				Expect(hyperRetrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(bybitRetrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.00015)))
			})

			It("should handle multiple assets for the same exchange", func() {
				now := time.Now()

				btcRate := connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0001),
					Timestamp:   now,
					MarkPrice:   numerical.NewFromFloat(50000),
				}

				ethRate := connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0002),
					Timestamp:   now,
					MarkPrice:   numerical.NewFromFloat(3000),
				}

				store.UpdateFundingRate(btc, "hyperliquid", btcRate)
				store.UpdateFundingRate(eth, "hyperliquid", ethRate)

				btcRetrieved := store.GetFundingRate(btc, "hyperliquid")
				ethRetrieved := store.GetFundingRate(eth, "hyperliquid")

				Expect(btcRetrieved).NotTo(BeNil())
				Expect(ethRetrieved).NotTo(BeNil())
				Expect(btcRetrieved.MarkPrice).To(Equal(numerical.NewFromFloat(50000)))
				Expect(ethRetrieved.MarkPrice).To(Equal(numerical.NewFromFloat(3000)))
			})

			It("should update existing funding rate", func() {
				now := time.Now()

				initialRate := connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0001),
					Timestamp:   now,
					MarkPrice:   numerical.NewFromFloat(50000),
				}

				store.UpdateFundingRate(btc, "hyperliquid", initialRate)

				updatedRate := connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0003),
					Timestamp:   now.Add(time.Hour),
					MarkPrice:   numerical.NewFromFloat(51000),
				}

				store.UpdateFundingRate(btc, "hyperliquid", updatedRate)

				retrieved := store.GetFundingRate(btc, "hyperliquid")
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.0003)))
				Expect(retrieved.MarkPrice).To(Equal(numerical.NewFromFloat(51000)))
			})
		})
	})

	Describe("UpdateFundingRates", func() {
		Context("when adding funding rates in batch", func() {
			It("should store multiple funding rates at once", func() {
				now := time.Now()

				rates := map[portfolio.Asset]connector.FundingRate{
					btc: {
						CurrentRate: numerical.NewFromFloat(0.0001),
						Timestamp:   now,
						MarkPrice:   numerical.NewFromFloat(50000),
					},
					eth: {
						CurrentRate: numerical.NewFromFloat(0.0002),
						Timestamp:   now,
						MarkPrice:   numerical.NewFromFloat(3000),
					},
				}

				store.UpdateFundingRates("hyperliquid", rates)

				btcRetrieved := store.GetFundingRate(btc, "hyperliquid")
				ethRetrieved := store.GetFundingRate(eth, "hyperliquid")

				Expect(btcRetrieved).NotTo(BeNil())
				Expect(ethRetrieved).NotTo(BeNil())
				Expect(btcRetrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(ethRetrieved.CurrentRate).To(Equal(numerical.NewFromFloat(0.0002)))
			})
		})
	})

	Describe("GetFundingRatesForAsset", func() {
		Context("when retrieving all funding rates for an asset", func() {
			It("should return funding rates from all exchanges", func() {
				now := time.Now()

				store.UpdateFundingRate(btc, "hyperliquid", connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0001),
					Timestamp:   now,
				})
				store.UpdateFundingRate(btc, "bybit", connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.00015),
					Timestamp:   now,
				})

				ratesMap := store.GetFundingRatesForAsset(btc)

				Expect(ratesMap).To(HaveLen(2))
				Expect(ratesMap["hyperliquid"].CurrentRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(ratesMap["bybit"].CurrentRate).To(Equal(numerical.NewFromFloat(0.00015)))
			})

			It("should return empty map for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				ratesMap := store.GetFundingRatesForAsset(unknown)
				Expect(ratesMap).To(BeEmpty())
			})
		})
	})

	Describe("GetFundingRate", func() {
		Context("when retrieving a specific funding rate", func() {
			It("should return nil for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				rate := store.GetFundingRate(unknown, "hyperliquid")
				Expect(rate).To(BeNil())
			})

			It("should return nil for unknown exchange", func() {
				store.UpdateFundingRate(btc, "hyperliquid", connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0001),
				})

				rate := store.GetFundingRate(btc, "unknown_exchange")
				Expect(rate).To(BeNil())
			})
		})
	})

	Describe("GetAllAssetsWithFundingRates", func() {
		Context("when retrieving all assets with funding rates", func() {
			It("should return all assets that have funding rates", func() {
				store.UpdateFundingRate(btc, "hyperliquid", connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0001),
				})
				store.UpdateFundingRate(eth, "hyperliquid", connector.FundingRate{
					CurrentRate: numerical.NewFromFloat(0.0002),
				})

				assets := store.GetAllAssetsWithFundingRates()

				Expect(assets).To(HaveLen(2))
				Expect(assets).To(ContainElement(btc))
				Expect(assets).To(ContainElement(eth))
			})

			It("should return empty slice when no funding rates exist", func() {
				assets := store.GetAllAssetsWithFundingRates()
				Expect(assets).To(BeEmpty())
			})
		})
	})
})
