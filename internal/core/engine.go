package core

import (
	"fmt"

	"trading-algo-generator/internal/eval"
	"trading-algo-generator/internal/execution"
	"trading-algo-generator/internal/features"
	"trading-algo-generator/internal/risk"
	"trading-algo-generator/internal/strategy"
)

// Engine wires together features, strategy, risk, and execution.
type Engine struct {
	Strategy   strategy.Strategy
	Features   features.Engine
	Risk       *risk.Manager
	Broker     execution.Broker
	Evaluator  *eval.Evaluator
	TickSize   float64
	TradeSize  int64
	Symbol     string
	Position   Position
}

func (e *Engine) OnTick(tick Tick) error {
	e.Risk.ResetIfNewSession(tick)
	features := e.Features.Build(tick)

	if e.Position.Open {
		e.Risk.UpdateStops(&e.Position, tick)
		if e.stopTriggered(tick) {
			return e.closePosition(tick, "stop")
		}
	}

	signal := e.Strategy.OnTick(tick, features, e.Position)
	if signal == nil {
		return nil
	}
	if signal.Direction == Flat {
		return nil
	}
	if !e.Risk.AllowEntry() {
		return nil
	}
	if e.Position.Open {
		return nil
	}
	return e.openPosition(tick, signal)
}

func (e *Engine) openPosition(tick Tick, signal *Signal) error {
	order := Order{
		Timestamp: tick.Timestamp,
		Direction: signal.Direction,
		Size:      e.TradeSize,
		Type:      Market,
		Price:     tick.Close,
	}
	fill, err := e.Broker.PlaceOrder(order)
	if err != nil {
		return err
	}
	e.Position = Position{
		Open:       true,
		Direction:  signal.Direction,
		EntryTime:  tick.Timestamp,
		EntryPrice: fill.Price,
		Size:       fill.Size,
		StopPrice:  0,
	}
	e.Risk.DailyTrades++
	return nil
}

func (e *Engine) closePosition(tick Tick, reason string) error {
	if !e.Position.Open {
		return nil
	}
	fill, err := e.Broker.ClosePosition(e.Position, tick.Close, reason)
	if err != nil {
		return err
	}
	pnl := e.realizedPnL(fill.Price)
	trade := Trade{
		EntryTime: e.Position.EntryTime,
		ExitTime:  tick.Timestamp,
		Entry:     e.Position.EntryPrice,
		Exit:      fill.Price,
		Size:      e.Position.Size,
		Direction: e.Position.Direction,
		PnL:       pnl,
		Reason:    reason,
	}
	e.Evaluator.Record(trade)
	e.Risk.ApplyDailyPnL(pnl)
	e.Position = Position{}
	return nil
}

func (e *Engine) stopTriggered(tick Tick) bool {
	if !e.Position.Open || e.Position.StopPrice == 0 {
		return false
	}
	if e.Position.Direction == Long {
		return tick.Close <= e.Position.StopPrice
	}
	return tick.Close >= e.Position.StopPrice
}

func (e *Engine) realizedPnL(exitPrice float64) float64 {
	if !e.Position.Open {
		return 0
	}
	move := exitPrice - e.Position.EntryPrice
	if e.Position.Direction == Short {
		move = -move
	}
	return move * float64(e.Position.Size)
}

func (e *Engine) Flush(reason string) error {
	if e.Position.Open {
		return e.closePosition(Tick{Timestamp: e.Position.EntryTime, Close: e.Position.EntryPrice}, reason)
	}
	return nil
}

func (e *Engine) Validate() error {
	if e.Strategy == nil {
		return fmt.Errorf("strategy required")
	}
	if e.Broker == nil {
		return fmt.Errorf("broker required")
	}
	if e.Evaluator == nil {
		return fmt.Errorf("evaluator required")
	}
	if e.Risk == nil {
		return fmt.Errorf("risk manager required")
	}
	return nil
}
