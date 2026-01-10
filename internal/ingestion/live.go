package ingestion

import (
	"time"

	"trading-algo-generator/internal/core"
)

// LiveSimulator emits ticks from a historical stream, preserving timestamp gaps.
type LiveSimulator struct {
	Speed float64
}

func (s LiveSimulator) Stream(ticks <-chan core.Tick) <-chan core.Tick {
	out := make(chan core.Tick)
	go func() {
		defer close(out)
		var last *core.Tick
		for tick := range ticks {
			if last != nil {
				delta := tick.Timestamp.Sub(last.Timestamp)
				if delta > 0 {
					sleep := delta
					if s.Speed > 0 {
						sleep = time.Duration(float64(delta) / s.Speed)
					}
					if sleep > 0 {
						time.Sleep(sleep)
					}
				}
			}
			clone := tick
			out <- clone
			last = &clone
		}
	}()
	return out
}
