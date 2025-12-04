package health

import (
	"fmt"
	"sync"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

type connectorHealthStore struct {
	connectorRegistry registry.ConnectorRegistry
	connectors        map[connector.ExchangeName]*connectorStatus
	timeProvider      temporal.TimeProvider
	mu                sync.RWMutex
	stopChan          chan struct{}
}

type connectorStatus struct {
	State     health.ConnectionState
	LastError error
	ErrorTime time.Time
}

func NewConnectorHealthStore(
	timeProvider temporal.TimeProvider,
	connectorRegistry registry.ConnectorRegistry,
) health.ConnectorErrorStore {
	errorStore := &connectorHealthStore{
		timeProvider:      timeProvider,
		connectorRegistry: connectorRegistry,
		connectors:        make(map[connector.ExchangeName]*connectorStatus),
		stopChan:          make(chan struct{}),
	}

	connectors := connectorRegistry.GetReadyWebSocketConnectors()

	for _, socketConnector := range connectors {
		name := socketConnector.GetConnectorInfo().Name
		errorStore.connectors[name] = &connectorStatus{
			State: health.StateConnected,
		}
	}

	// Start listening to error channels from all connectors
	for _, socketConnector := range connectors {
		go errorStore.listenToConnectorErrors(socketConnector)
	}

	return errorStore
}

// listenToConnectorErrors reads from a connector's error channel and records errors
func (c *connectorHealthStore) listenToConnectorErrors(socketConnector connector.WebSocketConnector) {
	name := socketConnector.GetConnectorInfo().Name
	errChan := socketConnector.ErrorChannel()

	for {
		select {
		case <-c.stopChan:
			return
		case err := <-errChan:
			if err != nil {
				c.RecordConnectorError(name, err)
			}
		}
	}
}

func (c *connectorHealthStore) RecordConnectorError(name connector.ExchangeName, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.connectors[name]; !exists {
		c.connectors[name] = &connectorStatus{
			State: health.StateConnecting,
		}
	}

	fmt.Printf("Warning: Recording error for untracked connector %s\n", name)
	fmt.Printf("Error: %v\n", err)

	c.connectors[name].LastError = err
	c.connectors[name].ErrorTime = c.timeProvider.Now()
	c.connectors[name].State = health.StateDegraded
}

func (c *connectorHealthStore) UpdateConnectionState(name connector.ExchangeName, state health.ConnectionState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.connectors[name]; !exists {
		c.connectors[name] = &connectorStatus{}
	}

	c.connectors[name].State = state
}

func (c *connectorHealthStore) GetConnectorState(name connector.ExchangeName) (health.ConnectionState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status, exists := c.connectors[name]
	if !exists {
		return "", false
	}

	return status.State, true
}

func (c *connectorHealthStore) GetConnectorError(name connector.ExchangeName) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status, exists := c.connectors[name]
	if !exists {
		return nil
	}

	return status.LastError
}

func (c *connectorHealthStore) CountTrackedConnectors() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.connectors)
}

// GetUnhealthyConnectors returns all connectors with connection errors
func (c *connectorHealthStore) GetUnhealthyConnectors() []connector.ExchangeName {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var unhealthy []connector.ExchangeName
	for name, status := range c.connectors {
		if status.State != health.StateConnected {
			unhealthy = append(unhealthy, name)
		}
	}

	return unhealthy
}

func (c *connectorHealthStore) GetErrorReport() *health.ConnectorErrorReport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	report := &health.ConnectorErrorReport{
		Errors: make(map[string]health.ConnectorError),
	}

	for name, status := range c.connectors {
		if status.State != health.StateConnected {
			report.Errors[string(name)] = health.ConnectorError{
				State:     status.State,
				Error:     status.LastError,
				ErrorTime: status.ErrorTime.Unix(),
			}
		}
	}

	return report
}

// Stop stops listening to connector error channels
func (c *connectorHealthStore) Stop() {
	close(c.stopChan)
}
