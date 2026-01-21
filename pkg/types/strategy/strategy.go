package strategy

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// StrategyType represents the type of trading strategy
type StrategyType string

const (
	StrategyTypeVolumeMaximizer StrategyType = "volume_maximizer"
	StrategyTypeCashCarry       StrategyType = "cash_carry"
	StrategyTypeArbitrage       StrategyType = "arbitrage"
	StrategyTypeTechnical       StrategyType = "technical"
	StrategyTypeMomentum        StrategyType = "momentum"
	StrategyTypeMeanReversion   StrategyType = "mean_reversion"
)

type StrategyName string

const (
	CashCarry       StrategyName = "Cash Carry"
	VolumeMaximizer StrategyName = "Volume Maximizer"
	Momentum        StrategyName = "Momentum"
)

type Strategy interface {
	GetSignals(ctx StrategyContext) ([]*Signal, error)

	GetName() StrategyName
	GetDescription() string
	GetRiskLevel() RiskLevel
	GetStrategyType() StrategyType
	ExecutionConfig() *ExecutionConfig
	WithExecutionConfig(*ExecutionConfig)
	GetLastRunAt() time.Time
}

// RequiredAsset specifies an asset and which instrument types are needed
type RequiredAsset struct {
	Symbol      portfolio.Asset
	Instruments []connector.Instrument
}

type StrategyConfig struct {
	Name            StrategyName
	Enabled         bool
	MaxPositionSize float64
}

type StrategyMetadata interface {
	GetRequiredFields() []string
	Validate() error
}

// RiskLevel represents the risk classification of a strategy
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type StrategyExecution struct {
	Orders []connector.Order
	Trades []connector.Trade
}
