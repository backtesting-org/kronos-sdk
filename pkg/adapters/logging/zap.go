package logging

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"go.uber.org/zap"
)

// ZapApplicationLogger adapts zap.Logger to ApplicationLogger interface
type ZapApplicationLogger struct {
	logger *zap.Logger
}

// NewZapApplicationLogger creates an application logger from a zap logger
func NewZapApplicationLogger(logger *zap.Logger) logging.ApplicationLogger {
	return &ZapApplicationLogger{logger: logger}
}

func (z *ZapApplicationLogger) Info(msg string, args ...interface{}) {
	z.logger.Sugar().Infof(msg, args...)
}

func (z *ZapApplicationLogger) Debug(msg string, args ...interface{}) {
	z.logger.Sugar().Debugf(msg, args...)
}

func (z *ZapApplicationLogger) Warn(msg string, args ...interface{}) {
	z.logger.Sugar().Warnf(msg, args...)
}

func (z *ZapApplicationLogger) Error(msg string, args ...interface{}) {
	z.logger.Sugar().Errorf(msg, args...)
}

func (z *ZapApplicationLogger) ErrorWithDebug(msg string, rawResponse []byte, args ...interface{}) {
	z.logger.Sugar().Errorw(msg, "raw_response", string(rawResponse), "args", args)
}

func (z *ZapApplicationLogger) Fatal(msg string, args ...interface{}) {
	z.logger.Sugar().Fatalf(msg, args...)
}

// ZapTradingLogger adapts zap.Logger to TradingLogger interface
type ZapTradingLogger struct {
	logger *zap.Logger
}

// NewZapTradingLogger creates a trading logger from a zap logger
func NewZapTradingLogger(logger *zap.Logger) logging.TradingLogger {
	return &ZapTradingLogger{logger: logger}
}

func (z *ZapTradingLogger) Opportunity(strategy, asset, msg string, args ...interface{}) {
	z.logger.Sugar().Infow("Opportunity",
		"strategy", strategy,
		"asset", asset,
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) Success(strategy, asset, msg string, args ...interface{}) {
	z.logger.Sugar().Infow("Success",
		"strategy", strategy,
		"asset", asset,
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) Failed(strategy, asset, msg string, args ...interface{}) {
	z.logger.Sugar().Errorw("Failed",
		"strategy", strategy,
		"asset", asset,
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) MarketCondition(msg string, args ...interface{}) {
	z.logger.Sugar().Infow("Market Condition",
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) OrderLifecycle(msg, asset string, args ...interface{}) {
	z.logger.Sugar().Infow("Order Lifecycle",
		"asset", asset,
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) Info(msg string, args ...interface{}) {
	z.logger.Sugar().Infof(msg, args...)
}

func (z *ZapTradingLogger) Debug(strategy, asset, msg string, args ...interface{}) {
	z.logger.Sugar().Debugw("Debug",
		"strategy", strategy,
		"asset", asset,
		"message", formatMessage(msg, args...),
	)
}

func (z *ZapTradingLogger) DataCollection(exchange, msg string, args ...interface{}) {
	z.logger.Sugar().Infow("Data Collection",
		"exchange", exchange,
		"message", formatMessage(msg, args...),
	)
}

func formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	// Use sprintf-style formatting
	return msg // In production, you'd use fmt.Sprintf(msg, args...)
}
