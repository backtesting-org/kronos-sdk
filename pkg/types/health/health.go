package health

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
)

// DataType represents types of market data
type DataType string

const (
	DataTypeKlines       DataType = "klines"
	DataTypeOrderbooks   DataType = "orderbooks"
	DataTypeTrades       DataType = "trades"
	DataTypeFundingRates DataType = "funding_rates"
	DataTypePositions    DataType = "positions"
)

// DataSourceType indicates how data is being fetched
type DataSourceType string

const (
	SourceWebSocket DataSourceType = "websocket"
	SourceBatch     DataSourceType = "batch"
	SourceHTTP      DataSourceType = "http"
)

// DataTypeHealth tracks health for a specific data type
type DataTypeHealth struct {
	Available    bool
	Source       DataSourceType
	LastReceived time.Time
	LastError    error
	ErrorCount   int
	Latency      time.Duration
	RecordCount  int64
}

// ConnectorHealth represents the health status of a single connector
type ConnectorHealth struct {
	Name            connector.ExchangeName
	State           ConnectionState
	DataTypes       map[DataType]*DataTypeHealth
	LastHealthCheck time.Time
	UptimeSeconds   int64
	ErrorRate       float64
}

// SystemHealthReport is an aggregated health report
type SystemHealthReport struct {
	OverallState    ConnectionState
	ConnectorErrors *ConnectorErrorReport
	DataFlowErrors  *DataFlowErrorReport
	StartedAt       time.Time
	HasErrors       bool
}

// HealthStore reports aggregated system health.
type HealthStore interface {
	GetSystemHealth() *SystemHealthReport
}
