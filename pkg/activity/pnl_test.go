package activity_test

import (
	"time"

	mockKronosActivity "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	mockAnalytics "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	kronosActivity "github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PNL", func() {
	var (
		mockPositions *mockKronosActivity.Positions
		mockTrades    *mockKronosActivity.Trades
		mockMarket    *mockAnalytics.Market
		pnl           kronosActivity.PNL
	)

	BeforeEach(func() {
		mockPositions = mockKronosActivity.NewPositions(GinkgoT())
		mockTrades = mockKronosActivity.NewTrades(GinkgoT())
		mockMarket = mockAnalytics.NewMarket(GinkgoT())
		pnl = activity.NewPNL(mockPositions, mockTrades, mockMarket)
	})

	Describe("GetFeesByStrategy", func() {
		It("should sum fees for a strategy", func() {
			strategyName := strategy.StrategyName("test-strategy")
			trades := []connector.Trade{
				{ID: "t1", Symbol: "BTC", Fee: numerical.NewFromFloat(10)},
				{ID: "t2", Symbol: "BTC", Fee: numerical.NewFromFloat(15)},
				{ID: "t3", Symbol: "ETH", Fee: numerical.NewFromFloat(5)},
			}

			mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

			result := pnl.GetFeesByStrategy(strategyName)

			Expect(result.Equal(numerical.NewFromFloat(30))).To(BeTrue())
		})

		It("should return zero when no trades", func() {
			strategyName := strategy.StrategyName("empty-strategy")

			mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return([]connector.Trade{})

			result := pnl.GetFeesByStrategy(strategyName)

			Expect(result.IsZero()).To(BeTrue())
		})
	})

	Describe("GetTotalFees", func() {
		It("should sum fees across all trades", func() {
			allTrades := []connector.Trade{
				{ID: "t1", Fee: numerical.NewFromFloat(10)},
				{ID: "t2", Fee: numerical.NewFromFloat(20)},
				{ID: "t3", Fee: numerical.NewFromFloat(30)},
			}

			mockTrades.EXPECT().GetAllTrades().Return(allTrades)

			result := pnl.GetTotalFees()

			Expect(result.Equal(numerical.NewFromFloat(60))).To(BeTrue())
		})
	})

	Describe("GetRealizedPNL", func() {
		Context("long position closed for profit", func() {
			It("should calculate profit correctly", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Buy 1 BTC at 50000, sell 1 BTC at 55000 = profit of 5000
				trades := []connector.Trade{
					{
						ID:        "t1",
						Symbol:    "BTC",
						Side:      connector.OrderSideBuy,
						Quantity:  numerical.NewFromFloat(1),
						Price:     numerical.NewFromFloat(50000),
						Fee:       numerical.NewFromFloat(50),
						Timestamp: time.Now(),
					},
					{
						ID:        "t2",
						Symbol:    "BTC",
						Side:      connector.OrderSideSell,
						Quantity:  numerical.NewFromFloat(1),
						Price:     numerical.NewFromFloat(55000),
						Fee:       numerical.NewFromFloat(55),
						Timestamp: time.Now(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				// PNL = 5000 - fees (50 + 55) = 4895
				Expect(result.Equal(numerical.NewFromFloat(4895))).To(BeTrue())
			})
		})

		Context("long position closed for loss", func() {
			It("should calculate loss correctly", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Buy 1 BTC at 50000, sell 1 BTC at 45000 = loss of 5000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(45000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				Expect(result.Equal(numerical.NewFromFloat(-5000))).To(BeTrue())
			})
		})

		Context("short position closed for profit", func() {
			It("should calculate profit correctly", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Sell 1 BTC at 50000, buy 1 BTC at 45000 = profit of 5000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(45000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})

		Context("short position closed for loss", func() {
			It("should calculate loss correctly", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Sell 1 BTC at 50000, buy 1 BTC at 55000 = loss of 5000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(55000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				Expect(result.Equal(numerical.NewFromFloat(-5000))).To(BeTrue())
			})
		})

		Context("partial close", func() {
			It("should calculate realized PNL for partial close only", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Buy 2 BTC at 50000, sell 1 BTC at 55000 = profit of 5000 on closed portion
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(55000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				// Only 1 BTC closed: (55000 - 50000) * 1 = 5000
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})

		Context("average cost basis", func() {
			It("should use weighted average entry price", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Buy 1 BTC at 50000, buy 1 BTC at 52000, sell 2 BTC at 54000
				// Avg entry = (50000 + 52000) / 2 = 51000
				// PNL = (54000 - 51000) * 2 = 6000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(52000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t3",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(54000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				Expect(result.Equal(numerical.NewFromFloat(6000))).To(BeTrue())
			})
		})

		Context("position flipping long to short", func() {
			It("should realize PNL on closed portion and track new short position", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Buy 1 BTC at 50000, sell 2 BTC at 55000
				// Closes long 1 BTC for profit of 5000, opens short 1 BTC at 55000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(55000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				// Only realized PNL from closing the long: (55000 - 50000) * 1 = 5000
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})

		Context("position flipping short to long", func() {
			It("should realize PNL on closed portion and track new long position", func() {
				strategyName := strategy.StrategyName("test-strategy")
				// Sell 1 BTC at 50000, buy 2 BTC at 45000
				// Closes short 1 BTC for profit of 5000, opens long 1 BTC at 45000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(45000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result := pnl.GetRealizedPNL(strategyName)

				// Only realized PNL from closing the short: (50000 - 45000) * 1 = 5000
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})
	})

	Describe("GetUnrealizedPNL", func() {
		Context("open long position", func() {
			It("should calculate unrealized profit", func() {
				strategyName := strategy.StrategyName("test-strategy")
				btc := portfolio.NewAsset("BTC")

				// Buy 1 BTC at 50000, current price 55000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)
				mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(55000), nil)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})

			It("should calculate unrealized loss", func() {
				strategyName := strategy.StrategyName("test-strategy")
				btc := portfolio.NewAsset("BTC")

				// Buy 1 BTC at 50000, current price 45000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)
				mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(45000), nil)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Equal(numerical.NewFromFloat(-5000))).To(BeTrue())
			})
		})

		Context("open short position", func() {
			It("should calculate unrealized profit", func() {
				strategyName := strategy.StrategyName("test-strategy")
				btc := portfolio.NewAsset("BTC")

				// Sell 1 BTC at 50000, current price 45000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)
				mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(45000), nil)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})

		Context("closed position", func() {
			It("should return zero unrealized PNL", func() {
				strategyName := strategy.StrategyName("test-strategy")

				// Buy 1 BTC, then sell 1 BTC - position is flat
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(55000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.IsZero()).To(BeTrue())
			})
		})

		Context("flipped position unrealized PNL", func() {
			It("should use new entry price for flipped short position", func() {
				strategyName := strategy.StrategyName("test-strategy")
				btc := portfolio.NewAsset("BTC")

				// Buy 1 BTC at 50000, sell 2 BTC at 55000 (now short 1 at 55000)
				// Current price 60000 - short is losing 5000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(55000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)
				mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(60000), nil)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				// Short 1 BTC at 55000, current price 60000: (55000 - 60000) * 1 = -5000
				Expect(result.Equal(numerical.NewFromFloat(-5000))).To(BeTrue())
			})

			It("should use new entry price for flipped long position", func() {
				strategyName := strategy.StrategyName("test-strategy")
				btc := portfolio.NewAsset("BTC")

				// Sell 1 BTC at 50000, buy 2 BTC at 45000 (now long 1 at 45000)
				// Current price 50000 - long is gaining 5000
				trades := []connector.Trade{
					{
						ID:       "t1",
						Symbol:   "BTC",
						Side:     connector.OrderSideSell,
						Quantity: numerical.NewFromFloat(1),
						Price:    numerical.NewFromFloat(50000),
						Fee:      numerical.Zero(),
					},
					{
						ID:       "t2",
						Symbol:   "BTC",
						Side:     connector.OrderSideBuy,
						Quantity: numerical.NewFromFloat(2),
						Price:    numerical.NewFromFloat(45000),
						Fee:      numerical.Zero(),
					},
				}

				mockPositions.EXPECT().GetTradesForStrategy(strategyName).Return(trades)
				mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(50000), nil)

				result, err := pnl.GetUnrealizedPNL(strategyName)

				Expect(err).NotTo(HaveOccurred())
				// Long 1 BTC at 45000, current price 50000: (50000 - 45000) * 1 = 5000
				Expect(result.Equal(numerical.NewFromFloat(5000))).To(BeTrue())
			})
		})
	})

	Describe("GetTotalPNL", func() {
		It("should sum realized and unrealized PNL", func() {
			btc := portfolio.NewAsset("BTC")

			// Trade 1: Closed position with 5000 profit
			// Trade 2: Open position with 3000 unrealized profit
			allTrades := []connector.Trade{
				// Closed BTC trade
				{ID: "t1", Symbol: "BTC", Side: connector.OrderSideBuy, Quantity: numerical.NewFromFloat(1), Price: numerical.NewFromFloat(50000), Fee: numerical.Zero()},
				{ID: "t2", Symbol: "BTC", Side: connector.OrderSideSell, Quantity: numerical.NewFromFloat(1), Price: numerical.NewFromFloat(55000), Fee: numerical.Zero()},
				// Open ETH position
				{ID: "t3", Symbol: "ETH", Side: connector.OrderSideBuy, Quantity: numerical.NewFromFloat(10), Price: numerical.NewFromFloat(3000), Fee: numerical.Zero()},
			}

			eth := portfolio.NewAsset("ETH")

			mockTrades.EXPECT().GetAllTrades().Return(allTrades)
			mockMarket.EXPECT().Price(btc).Return(numerical.NewFromFloat(55000), nil).Maybe()
			mockMarket.EXPECT().Price(eth).Return(numerical.NewFromFloat(3300), nil)

			result, err := pnl.GetTotalPNL()

			Expect(err).NotTo(HaveOccurred())
			// Realized: 5000, Unrealized: (3300-3000)*10 = 3000, Total: 8000
			Expect(result.Equal(numerical.NewFromFloat(8000))).To(BeTrue())
		})
	})
})
