# What SDK MUST Provide (Critical - Blockers)

Live-trading currently has ~800 lines of generic code that SDK already has in `internal/` but we can't access.

**Problem**: Go doesn't allow importing from other modules' `internal/` packages.
**Solution**: SDK must make these implementations public.

---

## 🔴 CRITICAL: Make Public IMMEDIATELY

These block us from using SDK properly. SDK has complete implementations in `internal/`, just needs to move to `pkg/`:

### 1. Market Data Store Implementation
**SDK has**: `internal/data/store/market/store.go` (complete, working)
**Make public as**: `pkg/stores/market/NewStore() market.MarketData`

**Why critical**: We need in-memory market data storage. Currently duplicating 337 lines.

**Current**: We have `internal/services/memory_store.go` (337 lines - duplicate of SDK)
**After SDK fix**: Delete our file, use `sdkmarket.NewStore()`

### 2. Time Provider Implementation
**SDK has**: `internal/time/time.go` (complete, working)
**Make public as**: `pkg/runtime/time/NewTimeProvider() temporal.TimeProvider`

**Why critical**: Need mockable time for testing and live time for production.

**Current**: We have `internal/services/time_provider.go` (50 lines - duplicate of SDK)
**After SDK fix**: Delete our file, use `sdktime.NewTimeProvider()`

### 3. Position/Activity Store Implementation
**SDK has**: `internal/data/store/activity/position/store.go` (complete, working)
**Make public as**: `pkg/stores/activity/position/NewStore() activity.Positions`

**Why critical**: Track strategy positions, orders, trades, P&L.

**Current**: We have `internal/services/position_manager.go` (170 lines - duplicate of SDK)
**After SDK fix**: Delete our file, use `sdkposition.NewStore()`

### 4. Trade Store Implementation
**SDK has**: `internal/data/store/activity/trade/store.go` (complete, working)
**Make public as**: `pkg/stores/activity/trade/NewStore() activity.Trades`

**Why critical**: Store and query trade history.

**Current**: Partially duplicated in position_manager.go
**After SDK fix**: Use `sdktrade.NewStore()`

---

## 🆕 NEW: Add to SDK (Not in internal/)

These are generic components that should be in SDK but aren't yet:

### 5. Event Bus
**Location**: Should be `pkg/events/bus.go`
**Interface**:
```go
type EventBus interface {
    Subscribe(topic string, handler func(event interface{}))
    Publish(topic string, event interface{})
    Close()
}
```

**Why needed**: Decouple components, enable event-driven architecture.

**Current**: We have `internal/services/event_bus.go` (80 lines)
**After SDK adds**: Delete our file, use SDK's

### 6. Kronos Provider/Factory
**Location**: Should be `pkg/kronos/provider.go`
**Purpose**: Create Kronos context instances with all services wired up

**Why needed**: Every deployment needs to create Kronos contexts for strategies.

**Current**: We have `internal/services/kronos_provider.go` (57 lines)
**After SDK adds**: Delete our file, use SDK's

### 7. Logging Adapters
**Location**: Should be `pkg/adapters/logging/zap.go`
**Purpose**: Adapt popular logging libraries (zap, logrus, etc.) to SDK interfaces

**Why needed**: Deployments use different logging libraries.

**Current**: We have `internal/services/trading_logger.go` + `application_logger.go` (100 lines)
**After SDK adds**: Delete our files, use SDK adapters

---

## ✅ What Stays in Live-Trading (Truly Deployment-Specific)

Only 4 files in `internal/services/` (after SDK provides above):

1. **market_feed.go** - Orchestrates which exchange to connect to (deployment choice)
2. **plugin_manager.go** - Loads strategy plugins from disk paths (deployment choice)
3. **strategy_executor.go** - Manages strategy lifecycle for this deployment
4. **trade_executor.go** - Executes trades with deployment-specific risk checks

Plus deployment infrastructure:
- `internal/database/` - Postgres (could be MySQL, SQLite, etc.)
- `internal/api/` - REST API (could be gRPC, GraphQL, etc.)
- `internal/config/` - YAML/ENV config (deployment choice)
- `external/exchanges/` - Which exchanges to support (deployment choice)

---

## 📊 Impact

**Current state**:
- Total: ~2,500 lines in live-trading
- Generic duplications: ~800 lines (blocks SDK usage)
- Deployment-specific: ~1,700 lines

**After SDK provides above**:
- Total: ~1,700 lines in live-trading
- Generic duplications: 0 lines ✅
- Deployment-specific: ~1,700 lines ✅

**Result**: 32% smaller, zero duplication, proper SDK usage

---

## 🎯 Action Items for SDK Team

**Phase 1** (Move existing `internal/` to `pkg/`):
1. ✅ Move `internal/data/store/market/` → `pkg/stores/market/`
2. ✅ Move `internal/data/store/activity/position/` → `pkg/stores/activity/position/`
3. ✅ Move `internal/data/store/activity/trade/` → `pkg/stores/activity/trade/`
4. ✅ Move `internal/time/` → `pkg/runtime/time/`

**Phase 2** (Add new components):
5. ⏳ Add event bus to `pkg/events/`
6. ⏳ Add Kronos provider to `pkg/kronos/`
7. ⏳ Add logging adapters to `pkg/adapters/logging/`

---

## 🚫 Stop Overcomplicating Rule

**Before writing ANY code in live-trading**:
1. Is it generic trading logic? → It belongs in SDK
2. Does SDK have it in `internal/`? → Ask to make public
3. Does SDK have it in `pkg/`? → Use it
4. Is it deployment-specific (DB, API, config, exchange choice)? → OK to implement here

**Never duplicate generic code again!**

---

## 🔴 ADDITIONAL: Connector Registry (Just Discovered)

### 8. Connector Registry Implementation
**We have**: `internal/connectors/registry.go` (47 lines - generic registry pattern)
**SDK has**: 
- Interface: `pkg/types/registry/connector.go` ✅
- Implementation: ❌ Missing (or in `internal/`)

**What it does**: Registry to manage multiple exchange connectors (Paradex, Binance, etc.)

**Why generic**: Every deployment needs to register available exchange connectors. This is just a map-based registry pattern.

**Make public as**: `pkg/registry/connector.go`
```go
package registry

func NewConnectorRegistry() ConnectorRegistry {
    // Implementation with Register(), GetConnector(), ListExchanges()
}
```

**Current**: We have `internal/connectors/registry.go` (47 lines)
**After SDK adds**: Delete our file, use SDK's

**Usage in main.go**:
```go
// Before
registry := connectors.NewRegistry()

// After
registry := sdkregistry.NewConnectorRegistry()
```

---

## Updated Impact

**Files to DELETE** (was 7, now 8):
1. memory_store.go (337 lines)
2. time_provider.go (50 lines)
3. position_manager.go (170 lines)
4. event_bus.go (80 lines)
5. kronos_provider.go (57 lines)
6. application_logger.go (50 lines)
7. trading_logger.go (50 lines)
8. **internal/connectors/registry.go (47 lines)** ← NEW

**Total generic duplication**: ~841 lines (was ~794)

**After SDK provides all 8 components**:
- Delete: 841 lines
- Keep: 9,700 lines (100% deployment-specific)
- Result: 8% code reduction, zero duplication
