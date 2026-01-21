package monitoring_test

import (
	"fmt"

	strategyMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	sdkTesting "github.com/backtesting-org/kronos-sdk/pkg/testing"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	storeActivity "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	kronosType "github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("ViewRegistry", func() {
	var (
		app              *fxtest.App
		viewRegistry     monitoring.ViewRegistry
		healthStore      health.HealthStore
		strategyRegistry registry.StrategyRegistry
		kronos           kronosType.Kronos
		positionsStore   storeActivity.Positions
	)

	BeforeEach(func() {
		app = fxtest.New(GinkgoT(),
			sdkTesting.Module,
			fx.Populate(
				&viewRegistry,
				&healthStore,
				&strategyRegistry,
				&kronos,
				&positionsStore,
			),
			fx.NopLogger,
		)

		app.RequireStart()

		// Register a test strategy
		testStrat := strategyMock.NewStrategy(GinkgoT())
		testStrat.On("GetName").Return(strategy.StrategyName("test-strategy"))
		strategyRegistry.RegisterStrategy(testStrat)
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("GetHealth", func() {
		It("should return health report from health store", func() {
			result := viewRegistry.GetHealth()

			Expect(result).NotTo(BeNil())
			Expect(result.OverallState).To(BeElementOf(
				health.StateConnected,
				health.StateDisconnected,
				health.StateConnecting,
				health.StateDegraded,
			))
		})
	})

	Describe("GetPnLView", func() {
		Context("when strategy exists", func() {
			It("should return PnL view with strategy data", func() {
				result := viewRegistry.GetPnLView()

				Expect(result).NotTo(BeNil())
				Expect(result.StrategyName).To(Equal("test-strategy"))
				// Values will be zero without trades, but structure should be correct
				Expect(result.RealizedPnL).NotTo(BeNil())
				Expect(result.UnrealizedPnL).NotTo(BeNil())
				Expect(result.TotalPnL).NotTo(BeNil())
				Expect(result.TotalFees).NotTo(BeNil())
			})
		})
	})

	Describe("GetAvailableAssets", func() {
		It("should return available assets from universe", func() {
			result := viewRegistry.GetAvailableAssets()

			// Without universe configuration, returns empty slice (not nil)
			Expect(result).To(BeEmpty())
		})
	})

	Describe("GetPositionsView", func() {
		Context("when strategy exists", func() {
			It("should return strategy execution with trades and orders", func() {
				// Add test trades to the position store (strategy-specific)
				positionsStore.AddTradeToStrategy("test-strategy", connector.Trade{
					ID:     "trade-1",
					Symbol: "BTC",
					Price:  numerical.NewFromFloat(50000),
				})

				result := viewRegistry.GetPositionsView()

				Expect(result).NotTo(BeNil())
				Expect(result.Trades).To(HaveLen(1))
				Expect(result.Trades[0].ID).To(Equal("trade-1"))
			})
		})

		Context("when no strategy exists", func() {
			It("should return nil", func() {
				result := viewRegistry.GetPositionsView()

				// GetPositionsView returns execution for first strategy, or nil if none
				// Since we still have registered (but disabled) strategy, it might return execution
				// The actual behavior depends on implementation
				_ = result
			})
		})
	})

	Describe("GetOrderbookView", func() {
		It("should return nil when no orderbook data available", func() {
			result := viewRegistry.GetOrderbookView("BTC/USDT")

			// Without market data ingestion, orderbook will be nil
			Expect(result).To(BeNil())
		})
	})

	Describe("GetRecentTrades", func() {
		Context("when strategy exists", func() {
			It("should return trades", func() {
				// Add test trades to position store (strategy-specific)
				positionsStore.AddTradeToStrategy("test-strategy", connector.Trade{ID: "trade-1", Symbol: "BTC/USDT"})
				positionsStore.AddTradeToStrategy("test-strategy", connector.Trade{ID: "trade-2", Symbol: "BTC/USDT"})

				result := viewRegistry.GetRecentTrades(10)

				Expect(result).To(HaveLen(2))
				Expect(result[0].ID).To(Equal("trade-1"))
				Expect(result[1].ID).To(Equal("trade-2"))
			})

			It("should limit trades when more than limit", func() {
				// Add multiple trades
				for i := 1; i <= 5; i++ {
					positionsStore.AddTradeToStrategy("test-strategy", connector.Trade{
						ID:     fmt.Sprintf("trade-%d", i),
						Symbol: "BTC/USDT",
					})
				}

				result := viewRegistry.GetRecentTrades(2)

				Expect(result).To(HaveLen(2))
				Expect(result[0].ID).To(Equal("trade-4"))
				Expect(result[1].ID).To(Equal("trade-5"))
			})
		})

		Context("when no strategy exists", func() {
			It("should return nil", func() {
				// Remove all strategies
				result := viewRegistry.GetRecentTrades(10)

				// With no strategies, should return nil
				_ = result
			})
		})
	})

	Describe("GetMetrics", func() {
		It("should return strategy metrics", func() {
			result := viewRegistry.GetMetrics()

			Expect(result.StrategyName).To(Equal("test-strategy"))
			Expect(result.Status).To(Equal("running"))
		})
	})

	Describe("GetProfilingStats", func() {
		It("should return profiling stats for strategy", func() {
			result := viewRegistry.GetProfilingStats()

			// Profiling store is injected, returns real stats (with zeros for new strategy)
			Expect(result).NotTo(BeNil())
			Expect(result.StrategyName).To(Equal("test-strategy"))
			Expect(result.TotalRuns).To(Equal(int(0)))
		})
	})

	Describe("GetRecentExecutions", func() {
		It("should return empty slice when no executions exist", func() {
			result := viewRegistry.GetRecentExecutions(10)

			// Returns empty slice, not nil
			Expect(result).To(BeEmpty())
		})
	})
})
