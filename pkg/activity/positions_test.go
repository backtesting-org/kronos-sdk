package activity_test

import (
	mockStoreActivity "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Positions", func() {
	var (
		mockStore *mockStoreActivity.Positions
		positions kronosActivity.Positions
	)

	BeforeEach(func() {
		mockStore = mockStoreActivity.NewPositions(GinkgoT())
		positions = activity.NewPositions(mockStore)
	})

	Describe("GetStrategyExecution", func() {
		It("should delegate to underlying store", func() {
			strategyName := strategy.StrategyName("test-strategy")
			expectedExecution := &strategy.StrategyExecution{
				Orders: []connector.Order{{ID: "order-1"}},
				Trades: []connector.Trade{{ID: "trade-1"}},
			}

			mockStore.EXPECT().GetStrategyExecution(strategyName).Return(expectedExecution)

			result := positions.GetStrategyExecution(strategyName)

			Expect(result).To(Equal(expectedExecution))
		})

		It("should return nil for unknown strategy", func() {
			strategyName := strategy.StrategyName("unknown")

			mockStore.EXPECT().GetStrategyExecution(strategyName).Return(nil)

			result := positions.GetStrategyExecution(strategyName)

			Expect(result).To(BeNil())
		})
	})

	Describe("GetAllStrategyExecutions", func() {
		It("should return all executions from underlying store", func() {
			expectedExecutions := map[strategy.StrategyName]*strategy.StrategyExecution{
				"strategy-1": {Orders: []connector.Order{{ID: "order-1"}}},
				"strategy-2": {Orders: []connector.Order{{ID: "order-2"}}},
			}

			mockStore.EXPECT().GetAllStrategyExecutions().Return(expectedExecutions)

			result := positions.GetAllStrategyExecutions()

			Expect(result).To(HaveLen(2))
			Expect(result["strategy-1"].Orders[0].ID).To(Equal("order-1"))
			Expect(result["strategy-2"].Orders[0].ID).To(Equal("order-2"))
		})

		It("should return empty map when no executions exist", func() {
			mockStore.EXPECT().GetAllStrategyExecutions().Return(map[strategy.StrategyName]*strategy.StrategyExecution{})

			result := positions.GetAllStrategyExecutions()

			Expect(result).To(BeEmpty())
		})
	})

	Describe("GetStrategyForOrder", func() {
		It("should find strategy for existing order", func() {
			orderID := "order-123"
			expectedStrategy := strategy.StrategyName("test-strategy")

			mockStore.EXPECT().GetStrategyForOrder(orderID).Return(expectedStrategy, true)

			name, found := positions.GetStrategyForOrder(orderID)

			Expect(found).To(BeTrue())
			Expect(name).To(Equal(expectedStrategy))
		})

		It("should return false for unknown order", func() {
			orderID := "unknown-order"

			mockStore.EXPECT().GetStrategyForOrder(orderID).Return(strategy.StrategyName(""), false)

			_, found := positions.GetStrategyForOrder(orderID)

			Expect(found).To(BeFalse())
		})
	})

	Describe("GetTotalOrderCount", func() {
		It("should return count from underlying store", func() {
			mockStore.EXPECT().GetTotalOrderCount().Return(int64(42))

			result := positions.GetTotalOrderCount()

			Expect(result).To(Equal(int64(42)))
		})

		It("should return zero when no orders exist", func() {
			mockStore.EXPECT().GetTotalOrderCount().Return(int64(0))

			result := positions.GetTotalOrderCount()

			Expect(result).To(Equal(int64(0)))
		})
	})

	Describe("GetTradesForStrategy", func() {
		It("should return trades from underlying store", func() {
			strategyName := strategy.StrategyName("test-strategy")
			expectedTrades := []connector.Trade{
				{ID: "trade-1", Price: numerical.NewFromFloat(100)},
				{ID: "trade-2", Price: numerical.NewFromFloat(200)},
			}

			mockStore.EXPECT().GetTradesForStrategy(strategyName).Return(expectedTrades)

			result := positions.GetTradesForStrategy(strategyName)

			Expect(result).To(HaveLen(2))
			Expect(result[0].ID).To(Equal("trade-1"))
			Expect(result[1].ID).To(Equal("trade-2"))
		})

		It("should return empty slice when no trades exist", func() {
			strategyName := strategy.StrategyName("test-strategy")

			mockStore.EXPECT().GetTradesForStrategy(strategyName).Return([]connector.Trade{})

			result := positions.GetTradesForStrategy(strategyName)

			Expect(result).To(BeEmpty())
		})
	})
})
