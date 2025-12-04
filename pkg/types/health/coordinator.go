package health

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
)

// DataFlowErrorReport is the aggregated data flow error report
type DataFlowErrorReport struct {
	Errors map[string]map[string]DataFlowError // connector -> (data type -> error details)
}

// DataFlowError contains error details for a specific data type
type DataFlowError struct {
	Error      error
	ErrorTime  int64 // unix timestamp
	ErrorCount int
}

// CoordinatorHealthStore tracks data flow and availability from coordinators.
type CoordinatorHealthStore interface {
	RecordDataReceived(name connector.ExchangeName, dataType DataType, source DataSourceType, latency time.Duration)
	RecordDataError(name connector.ExchangeName, dataType DataType, err error)
	MarkDataTypeAvailable(name connector.ExchangeName, dataType DataType, available bool)
	GetAvailableDataTypes(name connector.ExchangeName) []DataType
	IsDataTypeHealthy(name connector.ExchangeName, dataType DataType) bool
	HasReceivedData(name connector.ExchangeName, dataType DataType) bool
	WaitForFirstData(name connector.ExchangeName, dataType DataType, timeout time.Duration) error
	GetConnectorDataHealth(name connector.ExchangeName) map[DataType]*DataTypeHealth
	GetDegradedDataTypes() map[connector.ExchangeName][]DataType
	GetErrorReport() *DataFlowErrorReport
}
