package connector

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
)

// WebSocketConnector extends the base Connector with real-time capabilities
type WebSocketConnector interface {
	Connector

	// WebSocket lifecycle management
	StartWebSocket() error
	StopWebSocket() error
	IsWebSocketConnected() bool

	// Subscription management
	SubscribeOrderBook(asset portfolio.Asset, instrumentType Instrument) error
	SubscribeTrades(asset portfolio.Asset, instrumentType Instrument) error
	SubscribePositions(asset portfolio.Asset, instrumentType Instrument) error
	SubscribeAccountBalance() error
	SubscribeKlines(asset portfolio.Asset, interval string) error

	UnsubscribeKlines(asset portfolio.Asset, interval string) error
	UnsubscribeTrades(asset portfolio.Asset, instrumentType Instrument) error
	UnsubscribeOrderBook(asset portfolio.Asset, instrumentType Instrument) error
	UnsubscribePositions(asset portfolio.Asset, instrumentType Instrument) error
	UnsubscribeAccountBalance() error

	// Data access - returns map of channelKey -> channel
	// channelKey format varies by data type (e.g., "BTC-PERP", "ETH-1m")
	GetOrderBookChannels() map[string]<-chan OrderBook
	GetKlineChannels() map[string]<-chan Kline

	// Single-channel data access (for trades/positions/account balance)
	TradeUpdates() <-chan Trade
	PositionUpdates() <-chan Position
	AccountBalanceUpdates() <-chan AccountBalance
	ErrorChannel() <-chan error
}
