package replay

import (
	"time"

	"trading-algo-generator/internal/core"
)

// Engine replays ticks using their timestamps.
type Engine struct {
	Speed float64
}

func (e Engine) Run(ticks []core.Tick, handler func(core.Tick)) {
	var last *core.Tick
	for _, tick := range ticks {
		if last != nil {
			delta := tick.Timestamp.Sub(last.Timestamp)
			if delta > 0 {
				sleep := delta
				if e.Speed > 0 {
					sleep = time.Duration(float64(delta) / e.Speed)
				}
				if sleep > 0 {
					time.Sleep(sleep)
				}
			}
		}
		handler(tick)
		last = &tick
	}
}
