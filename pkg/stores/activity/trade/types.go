package trade

import (
	"sync"
	"sync/atomic"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
)

type TradeMap map[string]connector.Trade // tradeID -> Trade

type dataStore struct {
	mutex  sync.RWMutex
	trades atomic.Value // []connector.Trade (ordered list)
	byID   atomic.Value // TradeMap (for fast lookup)
}
