package core

import "time"

// Tick is the base event used across ingestion, replay, and strategies.
type Tick struct {
	Timestamp     time.Time     `json:"timestamp"`
	Open          float64       `json:"open"`
	High          float64       `json:"high"`
	Low           float64       `json:"low"`
	Close         float64       `json:"close"`
	Volume        int64         `json:"volume"`
	BidAskDelta   int64         `json:"bid_ask_delta"`
	VolumeProfile []PriceLevel  `json:"volume_profile"`
	Session       string        `json:"session"`
	Symbol        string        `json:"symbol"`
}

// PriceLevel is a single level in a volume profile.
type PriceLevel struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

// FeatureSet is a flat collection of scalar features for a tick.
type FeatureSet struct {
	Timestamp time.Time
	Values    map[string]float64
}

// Signal is a directional decision from a strategy.
type Signal struct {
	Timestamp  time.Time
	Direction  Direction
	Confidence float64
	Reason     string
}

// Direction is the trade side.
type Direction int

const (
	Flat Direction = iota
	Long
	Short
)

// Order is a request to the broker.
type Order struct {
	ID        string
	Timestamp time.Time
	Direction Direction
	Size      int64
	Type      OrderType
	Price     float64
}

// OrderType defines execution style.
type OrderType int

const (
	Market OrderType = iota
	Limit
)

// Fill represents an executed order.
type Fill struct {
	OrderID   string
	Timestamp time.Time
	Price     float64
	Size      int64
	Direction Direction
}

// Trade is a closed position record.
type Trade struct {
	EntryTime time.Time
	ExitTime  time.Time
	Entry     float64
	Exit      float64
	Size      int64
	Direction Direction
	PnL       float64
	Reason    string
}

// Position tracks an open trade.
type Position struct {
	Open       bool
	Direction  Direction
	EntryTime  time.Time
	EntryPrice float64
	Size       int64
	StopPrice  float64
	MaxFavorableTicks int64
}
