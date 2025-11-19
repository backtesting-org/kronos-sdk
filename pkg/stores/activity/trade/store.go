package trade

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/stores/activity"
)

func NewStore() activity.Trades {
	ds := &dataStore{}
	ds.trades.Store([]connector.Trade{})
	ds.byID.Store(make(TradeMap))
	return ds
}

// Clear removes all trades from the store (for simulation restart)
func (ds *dataStore) Clear() {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.trades.Store([]connector.Trade{})
	ds.byID.Store(make(TradeMap))
}

var _ activity.Trades = (*dataStore)(nil)
