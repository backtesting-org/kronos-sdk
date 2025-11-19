package position

import (
	"fmt"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

func (ds *dataStore) AddOrderToStrategy(strategyName strategy.StrategyName, order connector.Order) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getExecutions()
	execution := current[strategyName]

	if execution == nil {
		execution = &strategy.StrategyExecution{
			Orders: []connector.Order{},
			Trades: []connector.Trade{},
		}
	}

	// Add order
	execution.Orders = append(execution.Orders, order)

	// Store updated map
	updated := make(map[strategy.StrategyName]*strategy.StrategyExecution, len(current)+1)
	for k, v := range current {
		updated[k] = v
	}
	updated[strategyName] = execution
	ds.executions.Store(updated)
}

func (ds *dataStore) UpdateOrderStatus(strategyName strategy.StrategyName, orderID string, status connector.OrderStatus) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	current := ds.getExecutions()
	execution := current[strategyName]

	if execution == nil {
		return fmt.Errorf("strategy execution not found: %s", strategyName)
	}

	// Find and update order status
	found := false
	for i := range execution.Orders {
		if execution.Orders[i].ID == orderID {
			execution.Orders[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Store updated map
	updated := make(map[strategy.StrategyName]*strategy.StrategyExecution, len(current))
	for k, v := range current {
		updated[k] = v
	}
	updated[strategyName] = execution
	ds.executions.Store(updated)

	return nil
}
