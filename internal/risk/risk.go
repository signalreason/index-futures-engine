package risk

import (
	"time"

	"trading-algo-generator/internal/core"
)

// Settings define risk limits and stop logic.
type Settings struct {
	DailyStopLoss    float64
	PerTradeStopTicks int64
	BreakevenTicks   int64
	BreakevenPlus    int64
	TrailingTicks    int64
	TickSize         float64
	MaxDailyTrades   int
}

// Manager enforces risk rules and updates stops.
type Manager struct {
	Settings       Settings
	DailyPnL       float64
	DailyTrades    int
	LastSessionDay time.Time
	Halted         bool
}

func (m *Manager) ResetIfNewSession(tick core.Tick) {
	sessionDay := time.Date(tick.Timestamp.Year(), tick.Timestamp.Month(), tick.Timestamp.Day(), 0, 0, 0, 0, tick.Timestamp.Location())
	if m.LastSessionDay.IsZero() || !sessionDay.Equal(m.LastSessionDay) {
		m.DailyPnL = 0
		m.DailyTrades = 0
		m.Halted = false
		m.LastSessionDay = sessionDay
	}
}

// AllowEntry checks if a new trade can be opened.
func (m *Manager) AllowEntry() bool {
	if m.Halted {
		return false
	}
	if m.Settings.MaxDailyTrades > 0 && m.DailyTrades >= m.Settings.MaxDailyTrades {
		return false
	}
	return true
}

// UpdateStops modifies position stop price based on BE+1 and trailing rules.
func (m *Manager) UpdateStops(position *core.Position, tick core.Tick) {
	if !position.Open {
		return
	}
	moveTicks := ticksMoved(position, tick, m.Settings.TickSize)
	if moveTicks > position.MaxFavorableTicks {
		position.MaxFavorableTicks = moveTicks
	}

	// Per-trade stop
	stop := stopFromEntry(position, m.Settings.PerTradeStopTicks, m.Settings.TickSize)
	if position.StopPrice == 0 {
		position.StopPrice = stop
	}

	// Breakeven + 1 tick
	if m.Settings.BreakevenTicks > 0 && position.MaxFavorableTicks >= m.Settings.BreakevenTicks {
		beStop := breakevenStop(position, m.Settings.BreakevenPlus, m.Settings.TickSize)
		position.StopPrice = betterStop(position, position.StopPrice, beStop)
	}

	// Trailing stop
	if m.Settings.TrailingTicks > 0 && position.MaxFavorableTicks >= m.Settings.TrailingTicks {
		trailStop := trailingStop(position, position.MaxFavorableTicks, m.Settings.TrailingTicks, m.Settings.TickSize)
		position.StopPrice = betterStop(position, position.StopPrice, trailStop)
	}
}

// ApplyDailyPnL enforces the hard daily stop loss.
func (m *Manager) ApplyDailyPnL(pnl float64) {
	m.DailyPnL += pnl
	if m.Settings.DailyStopLoss < 0 && m.DailyPnL <= m.Settings.DailyStopLoss {
		m.Halted = true
	}
}

func ticksMoved(position *core.Position, tick core.Tick, tickSize float64) int64 {
	if tickSize <= 0 {
		return 0
	}
	var move float64
	if position.Direction == core.Long {
		move = tick.Close - position.EntryPrice
	} else {
		move = position.EntryPrice - tick.Close
	}
	return int64(move / tickSize)
}

func stopFromEntry(position *core.Position, stopTicks int64, tickSize float64) float64 {
	if stopTicks <= 0 || tickSize <= 0 {
		return 0
	}
	if position.Direction == core.Long {
		return position.EntryPrice - float64(stopTicks)*tickSize
	}
	return position.EntryPrice + float64(stopTicks)*tickSize
}

func breakevenStop(position *core.Position, plusTicks int64, tickSize float64) float64 {
	if tickSize <= 0 {
		return position.EntryPrice
	}
	adj := float64(plusTicks) * tickSize
	if position.Direction == core.Long {
		return position.EntryPrice + adj
	}
	return position.EntryPrice - adj
}

func trailingStop(position *core.Position, maxTicks int64, trailTicks int64, tickSize float64) float64 {
	if tickSize <= 0 {
		return position.EntryPrice
	}
	trailFromEntry := float64(maxTicks-trailTicks) * tickSize
	if position.Direction == core.Long {
		return position.EntryPrice + trailFromEntry
	}
	return position.EntryPrice - trailFromEntry
}

func betterStop(position *core.Position, current, candidate float64) float64 {
	if current == 0 {
		return candidate
	}
	if position.Direction == core.Long {
		if candidate > current {
			return candidate
		}
		return current
	}
	if candidate < current {
		return candidate
	}
	return current
}
