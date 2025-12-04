package health

import "github.com/backtesting-org/kronos-sdk/pkg/types/connector"

// ConnectionState represents the connection status
type ConnectionState string

const (
	StateConnected    ConnectionState = "connected"
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateDegraded     ConnectionState = "degraded"
)

// ConnectorErrorReport is the aggregated error report from all connectors
type ConnectorErrorReport struct {
	Errors map[string]ConnectorError // connector name -> error details
}

// ConnectorError contains error details for a single connector
type ConnectorError struct {
	State     ConnectionState
	Error     error
	ErrorTime int64 // unix timestamp
}

// ConnectorErrorStore tracks WebSocket connector-level errors and connection state
type ConnectorErrorStore interface {
	RecordConnectorError(name connector.ExchangeName, err error)
	UpdateConnectionState(name connector.ExchangeName, state ConnectionState)
	GetConnectorState(name connector.ExchangeName) (ConnectionState, bool)
	GetConnectorError(name connector.ExchangeName) error
	CountTrackedConnectors() int
	GetUnhealthyConnectors() []connector.ExchangeName
	GetErrorReport() *ConnectorErrorReport
}
