package execution

import (
	"fmt"
	"sync"
	"time"

	"trading-algo-generator/internal/core"
)

// Broker defines order execution for the engine.
type Broker interface {
	PlaceOrder(order core.Order) (core.Fill, error)
	ClosePosition(position core.Position, price float64, reason string) (core.Fill, error)
}

// MockBroker fills orders at the provided price immediately.
type MockBroker struct {
	mu       sync.Mutex
	counter  int64
	LastFill *core.Fill
}

func (b *MockBroker) PlaceOrder(order core.Order) (core.Fill, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counter++
	fill := core.Fill{
		OrderID:   fmt.Sprintf("MOCK-%d", b.counter),
		Timestamp: time.Now().UTC(),
		Price:     order.Price,
		Size:      order.Size,
		Direction: order.Direction,
	}
	b.LastFill = &fill
	return fill, nil
}

func (b *MockBroker) ClosePosition(position core.Position, price float64, reason string) (core.Fill, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counter++
	fill := core.Fill{
		OrderID:   fmt.Sprintf("MOCK-CLOSE-%d", b.counter),
		Timestamp: time.Now().UTC(),
		Price:     price,
		Size:      position.Size,
		Direction: opposite(position.Direction),
	}
	b.LastFill = &fill
	return fill, nil
}

func opposite(direction core.Direction) core.Direction {
	if direction == core.Long {
		return core.Short
	}
	if direction == core.Short {
		return core.Long
	}
	return core.Flat
}
