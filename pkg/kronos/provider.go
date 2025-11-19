package kronos

import (
	"github.com/backtesting-org/kronos-sdk/pkg/kronos/analytics"
	"github.com/backtesting-org/kronos-sdk/pkg/kronos/indicators"
	"github.com/backtesting-org/kronos-sdk/pkg/kronos/market"
	"github.com/backtesting-org/kronos-sdk/pkg/kronos/signal"
	"github.com/backtesting-org/kronos-sdk/pkg/kronos/trade"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
	marketstore "github.com/backtesting-org/kronos-sdk/pkg/types/stores/market"
)

// Provider creates Kronos context instances with all services wired up
type Provider interface {
	// NewKronos creates a read-only Kronos context for strategies
	NewKronos() *Kronos

	// NewKronosExecutor creates a Kronos context with trade execution capabilities
	NewKronosExecutor() *KronosExecutor
}

type provider struct {
	store         marketstore.MarketData
	tradingLogger logging.TradingLogger
	timeProvider  temporal.TimeProvider
}

// NewProvider creates a new Kronos provider
func NewProvider(
	store marketstore.MarketData,
	tradingLogger logging.TradingLogger,
	timeProvider temporal.TimeProvider,
) Provider {
	return &provider{
		store:         store,
		tradingLogger: tradingLogger,
		timeProvider:  timeProvider,
	}
}

func (p *provider) NewKronos() *Kronos {
	// Create services
	indicatorService := indicators.NewIndicatorService(p.store)
	marketService := market.NewMarketService(p.store)
	analyticsService := analytics.NewAnalyticsService(p.store)
	signalService := signal.NewService(p.timeProvider)

	// Create Kronos context
	return NewKronos(
		p.store,
		p.tradingLogger,
		indicatorService,
		marketService,
		analyticsService,
		signalService,
	)
}

func (p *provider) NewKronosExecutor() *KronosExecutor {
	// Create base Kronos
	baseKronos := p.NewKronos()

	// Create trade service
	tradeService := trade.NewTradeService(p.tradingLogger)

	// Create executor
	return NewKronosExecutor(baseKronos, tradeService)
}
