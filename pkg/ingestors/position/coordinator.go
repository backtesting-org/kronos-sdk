package position

import (
	"context"
	"fmt"
	"sync"

	"github.com/backtesting-org/kronos-sdk/pkg/types/data/ingestors"
	"github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// coordinator handles trade backfill on startup
type coordinator struct {
	positionStore     activity.Positions
	tradeStore        activity.Trades
	connectorRegistry registry.ConnectorRegistry
	logger            logging.ApplicationLogger

	isActive           bool
	mutex              sync.RWMutex
	tradeBackfillLimit int
	backfillCompleted  bool
}

func NewCoordinator(
	positionStore activity.Positions,
	tradeStore activity.Trades,
	connectorRegistry registry.ConnectorRegistry,
	logger logging.ApplicationLogger,
) ingestors.PositionCoordinator {
	return &coordinator{
		positionStore:      positionStore,
		tradeStore:         tradeStore,
		connectorRegistry:  connectorRegistry,
		logger:             logger,
		tradeBackfillLimit: 100, // Fetch last 100 trades on startup
		backfillCompleted:  false,
	}
}

func (pc *coordinator) Start(_ context.Context) error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if pc.isActive {
		return fmt.Errorf("position coordinator already active")
	}

	pc.isActive = true

	// Backfill trades on startup (one-time operation)
	if !pc.backfillCompleted {
		pc.logger.Info("🔄 Starting trade backfill from exchanges...")
		if err := pc.backfillTrades(); err != nil {
			pc.logger.Error("❌ Trade backfill failed: %v", err)
			return err
		}
		pc.backfillCompleted = true
		pc.logger.Info("✅ Trade backfill completed")
	}

	pc.logger.Info("✅ Position coordinator started")
	return nil
}

func (pc *coordinator) Stop() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if !pc.isActive {
		return nil
	}

	pc.isActive = false
	pc.logger.Info("🛑 Position coordinator stopped")
	return nil
}

func (pc *coordinator) IsActive() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.isActive
}

// backfillTrades fetches recent trade history from all exchanges on startup
func (pc *coordinator) backfillTrades() error {
	totalBackfilled := 0

	// Get all available connectors from registry
	connectors := pc.connectorRegistry.GetReadyWebSocketConnectors()
	if len(connectors) == 0 {
		pc.logger.Warn("⚠️  No connectors available for trade backfill")
		return nil
	}

	for _, conn := range connectors {
		exchangeName := conn.GetConnectorInfo().Name
		pc.logger.Info("📥 Backfilling trades from %s...", exchangeName)

		// Get all strategy executions to know which symbols to fetch
		executions := pc.positionStore.GetAllStrategyExecutions()
		symbols := pc.getUniqueSymbols(executions)

		if len(symbols) == 0 {
			pc.logger.Warn("⚠️  No symbols found for trade backfill on %s", exchangeName)
			continue
		}

		for _, symbol := range symbols {
			trades, err := conn.GetTradingHistory(symbol.Symbol(), pc.tradeBackfillLimit)
			if err != nil {
				pc.logger.Warn("⚠️  Failed to fetch trades for %s on %s: %v", symbol.Symbol(), exchangeName, err)
				continue
			}

			if len(trades) > 0 {
				// Ensure trades have exchange field set
				for i := range trades {
					trades[i].Exchange = exchangeName
				}

				// Add trades to global trade store
				pc.tradeStore.AddTrades(trades)
				totalBackfilled += len(trades)

				// Add trades to strategy
				for _, t := range trades {
					strategyName := pc.findStrategyForSymbol(executions, symbol)
					if strategyName != "" {
						pc.positionStore.AddTradeToStrategy(strategyName, t)
					}
				}

				pc.logger.Info("✅ Backfilled %d trades for %s on %s", len(trades), symbol.Symbol(), exchangeName)
			}
		}
	}

	pc.logger.Info("📊 Total trades backfilled: %d", totalBackfilled)
	return nil
}

// getUniqueSymbols extracts all unique symbols from strategy executions
func (pc *coordinator) getUniqueSymbols(executions map[strategy.StrategyName]*strategy.StrategyExecution) []portfolio.Asset {
	symbolMap := make(map[string]portfolio.Asset)

	for _, execution := range executions {
		if execution == nil {
			continue
		}

		for _, order := range execution.Orders {
			symbolMap[order.Symbol] = portfolio.NewAsset(order.Symbol)
		}
	}

	symbols := make([]portfolio.Asset, 0, len(symbolMap))
	for _, asset := range symbolMap {
		symbols = append(symbols, asset)
	}

	return symbols
}

// findStrategyForSymbol determines which strategy is trading a given symbol
func (pc *coordinator) findStrategyForSymbol(executions map[strategy.StrategyName]*strategy.StrategyExecution, asset portfolio.Asset) strategy.StrategyName {
	for strategyName, execution := range executions {
		if execution == nil {
			continue
		}

		// Check if this strategy has orders for this symbol
		for _, order := range execution.Orders {
			if order.Symbol == asset.Symbol() {
				return strategyName
			}
		}
	}

	return ""
}

func (pc *coordinator) GetStatus() map[string]interface{} {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	status := map[string]interface{}{
		"active":               pc.isActive,
		"backfill_completed":   pc.backfillCompleted,
		"trade_backfill_limit": pc.tradeBackfillLimit,
	}

	if pc.isActive {
		status["total_orders"] = pc.positionStore.GetTotalOrderCount()
		status["trade_store_count"] = pc.tradeStore.GetTradeCount()
	}

	return status
}
