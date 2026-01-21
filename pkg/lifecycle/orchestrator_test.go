package lifecycle_test

import (
	"context"
	"errors"
	"time"

	mockStrategy "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	sdkTesting "github.com/backtesting-org/kronos-sdk/pkg/testing"
	"github.com/backtesting-org/kronos-sdk/pkg/types/execution"
	lifecycleTypes "github.com/backtesting-org/kronos-sdk/pkg/types/lifecycle"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("Orchestrator", func() {
	var (
		app              *fxtest.App
		orchestrator     lifecycleTypes.Orchestrator
		strategyRegistry registry.StrategyRegistry
		executor         execution.Executor
		ctx              strategy.StrategyContext
	)

	BeforeEach(func() {
		ctx = strategy.NewStrategyContext(context.Background(), strategy.StrategyName("TestStrategy"))

		app = fxtest.New(GinkgoT(),
			sdkTesting.Module,
			fx.Populate(
				&orchestrator,
				&strategyRegistry,
				&executor,
			),
			fx.NopLogger,
		)
		app.RequireStart()
	})

	AfterEach(func() {
		if orchestrator != nil {
			_ = orchestrator.Stop(ctx)
		}
		app.RequireStop()
	})

	Describe("Starting and Stopping", func() {
		It("should start successfully", func() {
			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not allow starting Maybe", func() {
			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = orchestrator.Start(ctx)
			// Should warn but not error
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop successfully", func() {
			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = orchestrator.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle multiple stop calls gracefully", func() {
			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = orchestrator.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = orchestrator.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Strategy Execution", func() {
		It("should execute single strategy", func() {
			// Setup mock strategy
			mockStrat := mockStrategy.NewStrategy(GinkgoT())
			mockStrat.EXPECT().GetName().Return(strategy.StrategyName("Strategy1")).Maybe()
			mockStrat.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).Return([]*strategy.Signal{}, nil).Maybe()
			mockStrat.EXPECT().ExecutionConfig().Return(nil)

			strategyRegistry.RegisterStrategy(mockStrat)

			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for fixed ticker to trigger (50ms interval)
			time.Sleep(100 * time.Millisecond)
		})

		It("should execute multiple strategies concurrently", func() {
			// Setup mock strategies
			mockStrat1 := mockStrategy.NewStrategy(GinkgoT())
			mockStrat2 := mockStrategy.NewStrategy(GinkgoT())

			// Note if the default tick interval changes, these expectations may need adjustment
			mockStrat1.EXPECT().GetName().Return(strategy.StrategyName("Strategy1")).Maybe()
			mockStrat1.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).Return([]*strategy.Signal{}, nil).Maybe()
			mockStrat1.EXPECT().ExecutionConfig().Return(nil)

			mockStrat2.EXPECT().GetName().Return(strategy.StrategyName("Strategy2")).Maybe()
			mockStrat2.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).Return([]*strategy.Signal{}, nil).Maybe()
			mockStrat2.EXPECT().ExecutionConfig().Return(nil)

			strategyRegistry.RegisterStrategy(mockStrat1)
			strategyRegistry.RegisterStrategy(mockStrat2)

			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for fixed ticker to trigger
			time.Sleep(100 * time.Millisecond)

			err = orchestrator.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

		})

		It("should handle strategy returning error", func() {
			// Setup mock strategy that returns error
			mockStrat := mockStrategy.NewStrategy(GinkgoT())
			mockStrat.EXPECT().GetName().Return(strategy.StrategyName("Strategy1")).Maybe()
			mockStrat.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).Return(nil, errors.New("strategy error")).Maybe()
			mockStrat.EXPECT().ExecutionConfig().Return(nil)

			strategyRegistry.RegisterStrategy(mockStrat)

			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for fixed ticker to trigger
			time.Sleep(100 * time.Millisecond)
		})

		It("should handle no strategies registered", func() {
			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for ticker - should not panic with no strategies
			time.Sleep(100 * time.Millisecond)
		})
	})

	Describe("Panic Recovery", func() {
		It("should recover from strategy panic", func() {
			// Setup panicking strategy and normal strategy
			panicStrat := mockStrategy.NewStrategy(GinkgoT())
			normalStrat := mockStrategy.NewStrategy(GinkgoT())

			panicStrat.EXPECT().GetName().Return(strategy.StrategyName("PanicStrategy")).Maybe()
			panicStrat.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).RunAndReturn(func(ctx strategy.StrategyContext) ([]*strategy.Signal, error) {
				panic("intentional panic for testing")
			}).Maybe()
			panicStrat.EXPECT().ExecutionConfig().Return(nil)

			normalStrat.EXPECT().GetName().Return(strategy.StrategyName("NormalStrategy")).Maybe()
			normalStrat.EXPECT().GetSignals(mock.AnythingOfType("strategy.StrategyContext")).Return([]*strategy.Signal{}, nil).Maybe()
			normalStrat.EXPECT().ExecutionConfig().Return(nil)

			strategyRegistry.RegisterStrategy(panicStrat)
			strategyRegistry.RegisterStrategy(normalStrat)

			err := orchestrator.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for fixed ticker to trigger
			time.Sleep(100 * time.Millisecond)

			// Panic should be recovered - no crash
		})
	})
})
