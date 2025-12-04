package position_test

import (
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

var _ = Describe("Position Store - Executions", func() {
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

	Describe("StoreStrategyExecution", func() {
		Context("when storing a new strategy execution", func() {
			It("should store the execution correctly", func() {
				execution := &strategy.StrategyExecution{
					Orders: []connector.Order{
						{
							ID:       "order-1",
							Symbol:   "BTC",
							Side:     connector.OrderSideBuy,
							Quantity: numerical.NewFromFloat(1.0),
							Price:    numerical.NewFromFloat(50000),
							Status:   connector.OrderStatusNew,
						},
					},
					Trades: []connector.Trade{},
				}

				store.StoreStrategyExecution(strategyName, execution)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved).NotTo(BeNil())
				Expect(retrieved.Orders).To(HaveLen(1))
				Expect(retrieved.Orders[0].ID).To(Equal("order-1"))
			})

			It("should handle multiple strategies", func() {
				strategy1 := strategy.StrategyName("strategy-1")
				strategy2 := strategy.StrategyName("strategy-2")

				exec1 := &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
				}
				exec2 := &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-2"}},
				}

				store.StoreStrategyExecution(strategy1, exec1)
				store.StoreStrategyExecution(strategy2, exec2)

				retrieved1 := store.GetStrategyExecution(strategy1)
				retrieved2 := store.GetStrategyExecution(strategy2)

				Expect(retrieved1.Orders[0].ID).To(Equal("order-1"))
				Expect(retrieved2.Orders[0].ID).To(Equal("order-2"))
			})

			It("should overwrite existing execution", func() {
				exec1 := &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
				}
				store.StoreStrategyExecution(strategyName, exec1)

				exec2 := &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-2"}},
				}
				store.StoreStrategyExecution(strategyName, exec2)

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders).To(HaveLen(1))
				Expect(retrieved.Orders[0].ID).To(Equal("order-2"))
			})
		})
	})

	Describe("GetStrategyExecution", func() {
		Context("when retrieving a strategy execution", func() {
			It("should return nil for unknown strategy", func() {
				retrieved := store.GetStrategyExecution("unknown-strategy")
				Expect(retrieved).To(BeNil())
			})
		})
	})

	Describe("UpdateStrategyExecution", func() {
		Context("when updating an existing execution", func() {
			It("should apply the update function", func() {
				execution := &strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1", Status: connector.OrderStatusNew},
					},
				}
				store.StoreStrategyExecution(strategyName, execution)

				err := store.UpdateStrategyExecution(strategyName, func(exec *strategy.StrategyExecution) {
					exec.Orders[0].Status = connector.OrderStatusFilled
				})

				Expect(err).NotTo(HaveOccurred())

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders[0].Status).To(Equal(connector.OrderStatusFilled))
			})

			It("should return error for unknown strategy", func() {
				err := store.UpdateStrategyExecution("unknown", func(exec *strategy.StrategyExecution) {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strategy execution not found"))
			})

			It("should allow adding orders via update", func() {
				execution := &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
				}
				store.StoreStrategyExecution(strategyName, execution)

				err := store.UpdateStrategyExecution(strategyName, func(exec *strategy.StrategyExecution) {
					exec.Orders = append(exec.Orders, connector.Order{ID: "order-2"})
				})

				Expect(err).NotTo(HaveOccurred())

				retrieved := store.GetStrategyExecution(strategyName)
				Expect(retrieved.Orders).To(HaveLen(2))
			})
		})
	})

	Describe("GetAllStrategyExecutions", func() {
		Context("when retrieving all executions", func() {
			It("should return all stored executions", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
				})
				store.StoreStrategyExecution("strategy-2", &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-2"}},
				})

				all := store.GetAllStrategyExecutions()

				Expect(all).To(HaveLen(2))
				Expect(all["strategy-1"]).NotTo(BeNil())
				Expect(all["strategy-2"]).NotTo(BeNil())
			})

			It("should return empty map when no executions exist", func() {
				all := store.GetAllStrategyExecutions()
				Expect(all).To(BeEmpty())
			})
		})
	})

	Describe("GetTotalOrderCount", func() {
		Context("when counting orders across strategies", func() {
			It("should count orders from all strategies", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}, {ID: "order-2"}},
				})
				store.StoreStrategyExecution("strategy-2", &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-3"}},
				})

				count := store.GetTotalOrderCount()
				Expect(count).To(Equal(int64(3)))
			})

			It("should return 0 when no orders exist", func() {
				count := store.GetTotalOrderCount()
				Expect(count).To(Equal(int64(0)))
			})

			It("should handle empty executions", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{},
				})

				count := store.GetTotalOrderCount()
				Expect(count).To(Equal(int64(0)))
			})
		})
	})

	Describe("Clear", func() {
		Context("when clearing the store", func() {
			It("should remove all executions", func() {
				store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
				})

				store.Clear()

				all := store.GetAllStrategyExecutions()
				Expect(all).To(BeEmpty())
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("when multiple goroutines access the store", func() {
			It("should handle concurrent writes safely", func() {
				done := make(chan bool)
				iterations := 50

				// Writer 1
				go func() {
					for i := 0; i < iterations; i++ {
						store.StoreStrategyExecution("strategy-1", &strategy.StrategyExecution{
							Orders: []connector.Order{{ID: "order-1"}},
						})
					}
					done <- true
				}()

				// Writer 2
				go func() {
					for i := 0; i < iterations; i++ {
						store.StoreStrategyExecution("strategy-2", &strategy.StrategyExecution{
							Orders: []connector.Order{{ID: "order-2"}},
						})
					}
					done <- true
				}()

				// Reader
				go func() {
					for i := 0; i < iterations; i++ {
						_ = store.GetAllStrategyExecutions()
						_ = store.GetTotalOrderCount()
					}
					done <- true
				}()

				<-done
				<-done
				<-done

				// Verify data consistency
				all := store.GetAllStrategyExecutions()
				Expect(all).To(HaveLen(2))
			})
		})
	})
})
