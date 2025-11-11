package notifications

// This file contains type definitions used by the notification service
// These types mirror the core types but are defined here to avoid circular dependencies

// Trade represents a trade executed on an exchange
type Trade struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"` // "buy" or "sell"
	Price      float64 `json:"price"`
	Quantity   float64 `json:"quantity"`
	BaseAsset  string  `json:"base_asset"`
	QuoteAsset string  `json:"quote_asset"`
	Fee        float64 `json:"fee"`
	FeeCoin    string  `json:"fee_coin"`
	Timestamp  int64   `json:"timestamp"`
}

// Position represents a trading position
type Position struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"` // "buy" or "sell"
	Quantity      float64 `json:"quantity"`
	EntryPrice    float64 `json:"entry_price"`
	ExitPrice     float64 `json:"exit_price"`
	RealizedPnL   float64 `json:"realized_pnl"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	OpenTime      int64   `json:"open_time"`
	CloseTime     int64   `json:"close_time"`
}

// Order represents a trading order
type Order struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"` // "buy" or "sell"
	Type          string  `json:"type"` // "market", "limit", "stop", "take_profit", etc.
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	ExecutedPrice float64 `json:"executed_price"`
	Status        string  `json:"status"`
	Timestamp     int64   `json:"timestamp"`
}
