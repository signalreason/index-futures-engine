package eval

import (
	"math"

	"trading-algo-generator/internal/core"
)

// Summary captures strategy performance statistics.
type Summary struct {
	TotalTrades   int
	Wins          int
	Losses        int
	WinRate       float64
	Expectancy    float64
	MaxDrawdown   float64
	AverageWin    float64
	AverageLoss   float64
	EquityCurve   []float64
	WinLossDist   map[string]int
}

// Evaluator aggregates trades into metrics.
type Evaluator struct {
	Trades []core.Trade
}

func (e *Evaluator) Record(trade core.Trade) {
	e.Trades = append(e.Trades, trade)
}

func (e *Evaluator) Summary() Summary {
	var equity float64
	var peak float64
	var maxDD float64
	var wins int
	var losses int
	var winSum float64
	var lossSum float64
	dist := map[string]int{
		"loss_small":  0,
		"loss_medium": 0,
		"loss_large":  0,
		"win_small":   0,
		"win_medium":  0,
		"win_large":   0,
	}
	curve := make([]float64, 0, len(e.Trades))

	for _, trade := range e.Trades {
		equity += trade.PnL
		curve = append(curve, equity)
		if equity > peak {
			peak = equity
		}
		drawdown := peak - equity
		if drawdown > maxDD {
			maxDD = drawdown
		}
		if trade.PnL > 0 {
			wins++
			winSum += trade.PnL
			bucket(dist, trade.PnL, true)
		} else if trade.PnL < 0 {
			losses++
			lossSum += trade.PnL
			bucket(dist, trade.PnL, false)
		}
	}

	total := wins + losses
	winRate := 0.0
	if total > 0 {
		winRate = float64(wins) / float64(total)
	}
	avgWin := 0.0
	if wins > 0 {
		avgWin = winSum / float64(wins)
	}
	avgLoss := 0.0
	if losses > 0 {
		avgLoss = lossSum / float64(losses)
	}
	expectancy := 0.0
	if total > 0 {
		expectancy = (winRate * avgWin) + ((1 - winRate) * avgLoss)
	}

	return Summary{
		TotalTrades: total,
		Wins:        wins,
		Losses:      losses,
		WinRate:     winRate,
		Expectancy:  expectancy,
		MaxDrawdown: maxDD,
		AverageWin:  avgWin,
		AverageLoss: avgLoss,
		EquityCurve: curve,
		WinLossDist: dist,
	}
}

func (s Summary) ProfitFactor() float64 {
	if s.AverageLoss == 0 {
		return math.Inf(1)
	}
	return (s.AverageWin * float64(s.Wins)) / math.Abs(s.AverageLoss*float64(s.Losses))
}

func bucket(dist map[string]int, pnl float64, isWin bool) {
	absPnL := math.Abs(pnl)
	if isWin {
		if absPnL < 100 {
			dist["win_small"]++
		} else if absPnL < 300 {
			dist["win_medium"]++
		} else {
			dist["win_large"]++
		}
		return
	}
	if absPnL < 100 {
		dist["loss_small"]++
	} else if absPnL < 300 {
		dist["loss_medium"]++
	} else {
		dist["loss_large"]++
	}
}
