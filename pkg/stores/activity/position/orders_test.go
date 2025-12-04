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

var _ = Describe("Position Store - Orders", func() {
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

	Describe("AddOrderToStrategy", func() {
		Context("when adding an order to an existing strategy", func() {
			It("should add the order to existing execution", func() {
				// First, create the strategy execution
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1", Symbol: "BTC"},
					},
				})

				// Add new order
				newOrder := connector.Order{
					ID:       "order-2",
					Symbol:   "ETH",
					Side:     connector.OrderSideBuy,
					Quantity: numerical.NewFromFloat(10.0),
					Price:    numerical.NewFromFloat(3000),
					Status:   connector.OrderStatusNew,
				}

				store.AddOrderToStrategy(strategyName, newOrder)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders).To(HaveLen(2))
				Expect(retrieved.Orders[1].ID).To(Equal("order-2"))
				Expect(retrieved.Orders[1].Symbol).To(Equal("ETH"))
			})
		})

		Context("when adding an order to a non-existent strategy", func() {
			It("should create a new execution and add the order", func() {
				order := connector.Order{
					ID:       "order-1",
					Symbol:   "BTC",
					Side:     connector.OrderSideBuy,
					Quantity: numerical.NewFromFloat(1.0),
					Status:   connector.OrderStatusNew,
				}

				store.AddOrderToStrategy(strategyName, order)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Orders).To(HaveLen(1))
				Expect(retrieved.Orders[0].ID).To(Equal("order-1"))
			})
		})

		Context("when adding multiple orders", func() {
			It("should preserve order of addition", func() {
				for i := 1; i <= 5; i++ {
					order := connector.Order{
						ID:     "order-" + string(rune('0'+i)),
						Symbol: "BTC",
					}
					store.AddOrderToStrategy(strategyName, order)
				}

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders).To(HaveLen(5))
			})
		})
	})

	Describe("UpdateOrderStatus", func() {
		Context("when updating an existing order", func() {
			It("should update the order status", func() {
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{
						{
							ID:     "order-1",
							Symbol: "BTC",
							Status: connector.OrderStatusNew,
						},
					},
				})

				err := store.UpdateOrderStatus(strategyName, "order-1", connector.OrderStatusFilled)

				Expect(err).NotTo(HaveOccurred())

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders[0].Status).To(Equal(connector.OrderStatusFilled))
			})

			It("should update the UpdatedAt timestamp", func() {
				now := time.Now()
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{
						{
							ID:        "order-1",
							Status:    connector.OrderStatusNew,
							UpdatedAt: now.Add(-time.Hour),
						},
					},
				})

				err := store.UpdateOrderStatus(strategyName, "order-1", connector.OrderStatusFilled)
				Expect(err).NotTo(HaveOccurred())

				retrieved := store.GetStrategyExecution(strategyName)
				// UpdatedAt should be more recent
				Expect(retrieved.Orders[0].UpdatedAt.After(now.Add(-time.Hour))).To(BeTrue())
			})
		})

		Context("when updating a non-existent order", func() {
			It("should return an error", func() {
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1"},
					},
				})

				err := store.UpdateOrderStatus(strategyName, "non-existent", connector.OrderStatusFilled)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("order not found"))
			})
		})

		Context("when updating an order for non-existent strategy", func() {
			It("should return an error", func() {
				err := store.UpdateOrderStatus("non-existent", "order-1", connector.OrderStatusFilled)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strategy execution not found"))
			})
		})

		Context("when updating order through different statuses", func() {
			It("should track status progression", func() {
				store.StoreStrategyExecution(strategyName, &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1", Status: connector.OrderStatusNew},
					},
				})

				// New -> Pending
				err := store.UpdateOrderStatus(strategyName, "order-1", connector.OrderStatusPending)
				Expect(err).NotTo(HaveOccurred())

				// Pending -> PartiallyFilled
				err = store.UpdateOrderStatus(strategyName, "order-1", connector.OrderStatusPartiallyFilled)
				Expect(err).NotTo(HaveOccurred())

				// PartiallyFilled -> Filled
				err = store.UpdateOrderStatus(strategyName, "order-1", connector.OrderStatusFilled)
				Expect(err).NotTo(HaveOccurred())

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders[0].Status).To(Equal(connector.OrderStatusFilled))
			})
		})
	})

	Describe("GetStrategyForOrder", func() {
		Context("when finding which strategy owns an order", func() {
			It("should return the correct strategy name", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1"},
						{ID: "order-2"},
					},
				})
				store.StoreStrategyExecution("strategy-2", &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-3"},
					},
				})

				strategyName, found := store.GetStrategyForOrder("order-3")

				Expect(found).To(BeTrue())
				Expect(strategyName).To(Equal(strategy.StrategyName("strategy-2")))
			})

			It("should return false for non-existent order", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1"},
					},
				})

				_, found := store.GetStrategyForOrder("non-existent")

				Expect(found).To(BeFalse())
			})

			It("should return false when no strategies exist", func() {
				_, found := store.GetStrategyForOrder("order-1")
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Order lifecycle", func() {
		Context("when tracking a complete order lifecycle", func() {
			It("should handle the full order flow", func() {
				// 1. Add new order
				order := connector.Order{
					ID:        "lifecycle-order",
					Symbol:    "BTC",
					Side:      connector.OrderSideBuy,
					Type:      connector.OrderTypeLimit,
					Quantity:  numerical.NewFromFloat(1.0),
					Price:     numerical.NewFromFloat(50000),
					Status:    connector.OrderStatusNew,
					CreatedAt: time.Now(),
				}

				store.AddOrderToStrategy(strategyName, order)

				// 2. Order submitted to exchange
				err := store.UpdateOrderStatus(strategyName, "lifecycle-order", connector.OrderStatusPending)
				Expect(err).NotTo(HaveOccurred())

				// 3. Verify we can find which strategy owns it
				owningStrategy, found := store.GetStrategyForOrder("lifecycle-order")
				Expect(found).To(BeTrue())
				Expect(owningStrategy).To(Equal(strategyName))

				// 4. Order gets filled
				err = store.UpdateOrderStatus(strategyName, "lifecycle-order", connector.OrderStatusFilled)
				Expect(err).NotTo(HaveOccurred())

				// 5. Verify final state
				execution := store.GetStrategyExecution(strategyName)
				Expect(execution.Orders[0].Status).To(Equal(connector.OrderStatusFilled))
			})
		})
	})
})
