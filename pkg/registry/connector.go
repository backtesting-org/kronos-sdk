package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
)

type connectorRegistry struct {
	connectors map[connector.ExchangeName]connector.Connector
	enabled    map[connector.ExchangeName]bool
	mu         sync.RWMutex
}

// NewConnectorRegistry creates a new connector registry
func NewConnectorRegistry() registry.ConnectorRegistry {
	return &connectorRegistry{
		connectors: make(map[connector.ExchangeName]connector.Connector),
		enabled:    make(map[connector.ExchangeName]bool),
	}
}

func (cr *connectorRegistry) GetConnector(name connector.ExchangeName) (connector.Connector, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	conn, exists := cr.connectors[name]
	return conn, exists
}

func (cr *connectorRegistry) RegisterConnector(name connector.ExchangeName, conn connector.Connector) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.connectors[name] = conn
	cr.enabled[name] = true // Enabled by default
}

func (cr *connectorRegistry) RegisterAllConnectors(connectors []connector.Connector) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for _, conn := range connectors {
		name := conn.GetConnectorInfo().Name
		cr.connectors[name] = conn
		cr.enabled[name] = true
	}
}

func (cr *connectorRegistry) GetAvailableConnectors() []connector.Connector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	connectors := make([]connector.Connector, 0, len(cr.connectors))
	for _, conn := range cr.connectors {
		connectors = append(connectors, conn)
	}

	return connectors
}

func (cr *connectorRegistry) GetWebSocketConnectors() []connector.WebSocketConnector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var wsConnectors []connector.WebSocketConnector
	for _, conn := range cr.connectors {
		if wsConn, ok := conn.(connector.WebSocketConnector); ok {
			wsConnectors = append(wsConnectors, wsConn)
		}
	}

	return wsConnectors
}

func (cr *connectorRegistry) GetTradingWebSocketConnectors() []connector.WebSocketConnector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var tradingWSConnectors []connector.WebSocketConnector
	for _, conn := range cr.connectors {
		// Check if connector supports trading via WebSocket
		if wsConn, ok := conn.(connector.WebSocketConnector); ok {
			// Additional check could be added here for trading-specific websocket
			tradingWSConnectors = append(tradingWSConnectors, wsConn)
		}
	}

	return tradingWSConnectors
}

func (cr *connectorRegistry) EnableConnector(name connector.ExchangeName) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.connectors[name]; !exists {
		return fmt.Errorf("connector %s not found", name)
	}

	cr.enabled[name] = true
	return nil
}

func (cr *connectorRegistry) DisableConnector(name connector.ExchangeName) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.connectors[name]; !exists {
		return fmt.Errorf("connector %s not found", name)
	}

	cr.enabled[name] = false
	return nil
}

func (cr *connectorRegistry) IsConnectorEnabled(name connector.ExchangeName) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	enabled, exists := cr.enabled[name]
	return exists && enabled
}

func (cr *connectorRegistry) GetEnabledConnectors() []connector.Connector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	connectors := make([]connector.Connector, 0)
	for name, conn := range cr.connectors {
		if cr.enabled[name] {
			connectors = append(connectors, conn)
		}
	}

	return connectors
}

func (cr *connectorRegistry) GetDataTimeRange() (start, end time.Time, err error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	// This is a placeholder implementation
	// In practice, this would query the earliest and latest data timestamps
	// across all enabled connectors
	return time.Time{}, time.Time{}, fmt.Errorf("not implemented")
}
