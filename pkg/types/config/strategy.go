package config

import "time"

type StrategyConfig interface {
	Load(path string) (*Strategy, error)
	FindStrategies() ([]Strategy, error)
	Save(path string, config *Strategy) error
}

// Asset represents a trading asset with its required instruments
type Asset struct {
	Symbol      string   `yaml:"symbol"`
	Instruments []string `yaml:"instruments"`
}

// StrategyExecutionConfig defines strategy execution timing
type StrategyExecutionConfig struct {
	// Interval defines how frequently the strategy will be executed (e.g., "1m", "5m", "1h")
	// If not set, the global tick interval (50ms) is used
	Interval string `yaml:"interval,omitempty"`
}

// Strategy represents the parsed config.yml for a strategy
type Strategy struct {
	Name        string                   `yaml:"name"`
	Path        string                   `yaml:"-"`
	Description string                   `yaml:"description"`
	Status      StrategyStatus           `yaml:"-"`
	Error       string                   `yaml:"-"`
	Exchanges   []string                 `yaml:"exchanges"`
	Assets      map[string][]Asset       `yaml:"assets"`
	Parameters  map[string]interface{}   `yaml:"parameters"`
	Execution   *StrategyExecutionConfig `yaml:"execution,omitempty"`
}

type StrategyStatus string

const (
	StatusReady   StrategyStatus = "ready"
	StatusRunning StrategyStatus = "running"
	StatusStopped StrategyStatus = "stopped"
	StatusError   StrategyStatus = "error"
)

type Exchange struct {
	Name    string
	Enabled bool
	Assets  []string
}

// ParseExecutionInterval parses the execution interval from config
// Returns nil if no execution config is set (use global tick interval)
func (s *Strategy) ParseExecutionInterval() (*time.Duration, error) {
	if s.Execution == nil || s.Execution.Interval == "" {
		return nil, nil
	}

	interval, err := time.ParseDuration(s.Execution.Interval)
	if err != nil {
		return nil, err
	}

	return &interval, nil
}
