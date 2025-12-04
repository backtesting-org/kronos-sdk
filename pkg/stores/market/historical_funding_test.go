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

var _ = Describe("Market Data Store - Historical Funding Rates", func() {
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

	Describe("UpdateHistoricalFundingRates", func() {
		Context("when adding historical funding rates", func() {
			It("should store historical funding rates correctly", func() {
				now := time.Now()
				rates := []connector.HistoricalFundingRate{
					{
						FundingRate: numerical.NewFromFloat(0.0001),
						Timestamp:   now.Add(-24 * time.Hour),
					},
					{
						FundingRate: numerical.NewFromFloat(0.00015),
						Timestamp:   now.Add(-16 * time.Hour),
					},
					{
						FundingRate: numerical.NewFromFloat(0.0002),
						Timestamp:   now.Add(-8 * time.Hour),
					},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", rates)

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)
				Expect(ratesMap).To(HaveLen(1))
				Expect(ratesMap["hyperliquid"]).To(HaveLen(3))
				Expect(ratesMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(ratesMap["hyperliquid"][2].FundingRate).To(Equal(numerical.NewFromFloat(0.0002)))
			})

			It("should handle multiple exchanges for the same asset", func() {
				now := time.Now()

				hyperRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0001), Timestamp: now.Add(-8 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.0002), Timestamp: now},
				}

				bybitRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.00012), Timestamp: now.Add(-8 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.00022), Timestamp: now},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", hyperRates)
				store.UpdateHistoricalFundingRates(btc, "bybit", bybitRates)

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)

				Expect(ratesMap).To(HaveLen(2))
				Expect(ratesMap["hyperliquid"]).To(HaveLen(2))
				Expect(ratesMap["bybit"]).To(HaveLen(2))
				Expect(ratesMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(ratesMap["bybit"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.00012)))
			})

			It("should handle multiple assets", func() {
				now := time.Now()

				btcRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0001), Timestamp: now},
				}

				ethRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0003), Timestamp: now},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", btcRates)
				store.UpdateHistoricalFundingRates(eth, "hyperliquid", ethRates)

				btcMap := store.GetHistoricalFundingRatesForAsset(btc)
				ethMap := store.GetHistoricalFundingRatesForAsset(eth)

				Expect(btcMap["hyperliquid"]).To(HaveLen(1))
				Expect(ethMap["hyperliquid"]).To(HaveLen(1))
				Expect(btcMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0001)))
				Expect(ethMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0003)))
			})

			It("should replace existing historical rates for same asset/exchange", func() {
				now := time.Now()

				initialRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0001), Timestamp: now.Add(-8 * time.Hour)},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", initialRates)

				updatedRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0002), Timestamp: now.Add(-8 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.0003), Timestamp: now},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", updatedRates)

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)
				Expect(ratesMap["hyperliquid"]).To(HaveLen(2))
				Expect(ratesMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0002)))
				Expect(ratesMap["hyperliquid"][1].FundingRate).To(Equal(numerical.NewFromFloat(0.0003)))
			})

			It("should handle empty rates slice", func() {
				store.UpdateHistoricalFundingRates(btc, "hyperliquid", []connector.HistoricalFundingRate{})

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)
				Expect(ratesMap).To(HaveLen(1))
				Expect(ratesMap["hyperliquid"]).To(BeEmpty())
			})
		})
	})

	Describe("GetHistoricalFundingRatesForAsset", func() {
		Context("when retrieving historical funding rates", func() {
			It("should return empty map for unknown asset", func() {
				unknown := portfolio.NewAsset("UNKNOWN")
				ratesMap := store.GetHistoricalFundingRatesForAsset(unknown)
				Expect(ratesMap).To(BeEmpty())
			})

			It("should preserve ordering of rates", func() {
				now := time.Now()
				rates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0001), Timestamp: now.Add(-24 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.0002), Timestamp: now.Add(-16 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.0003), Timestamp: now.Add(-8 * time.Hour)},
					{FundingRate: numerical.NewFromFloat(0.0004), Timestamp: now},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", rates)

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)
				retrieved := ratesMap["hyperliquid"]

				Expect(retrieved).To(HaveLen(4))
				// Verify order is preserved
				for i, rate := range retrieved {
					Expect(rate.FundingRate).To(Equal(rates[i].FundingRate))
				}
			})

			It("should not affect other exchanges when updating one", func() {
				now := time.Now()

				hyperRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0001), Timestamp: now},
				}

				bybitRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0002), Timestamp: now},
				}

				store.UpdateHistoricalFundingRates(btc, "hyperliquid", hyperRates)
				store.UpdateHistoricalFundingRates(btc, "bybit", bybitRates)

				// Update hyperliquid rates
				newHyperRates := []connector.HistoricalFundingRate{
					{FundingRate: numerical.NewFromFloat(0.0005), Timestamp: now},
				}
				store.UpdateHistoricalFundingRates(btc, "hyperliquid", newHyperRates)

				ratesMap := store.GetHistoricalFundingRatesForAsset(btc)

				// Bybit rates should be unchanged
				Expect(ratesMap["bybit"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0002)))
				// Hyperliquid rates should be updated
				Expect(ratesMap["hyperliquid"][0].FundingRate).To(Equal(numerical.NewFromFloat(0.0005)))
			})
		})
	})
})
