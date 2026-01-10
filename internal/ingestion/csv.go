package ingestion

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"trading-algo-generator/internal/core"
)

const (
	// CSVTimestampLayout is the expected layout for timestamp fields.
	CSVTimestampLayout = time.RFC3339Nano
)

// CSVReader loads ticks from a CSV file.
type CSVReader struct {
	Path string
}

func (r CSVReader) Stream() (<-chan core.Tick, <-chan error) {
	out := make(chan core.Tick)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		file, err := os.Open(r.Path)
		if err != nil {
			errCh <- err
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		header, err := reader.Read()
		if err != nil {
			errCh <- err
			return
		}
		indices := headerIndex(header)

		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errCh <- err
				return
			}

			tick, err := parseRecord(rec, indices)
			if err != nil {
				errCh <- err
				return
			}
			out <- tick
		}
	}()

	return out, errCh
}

func headerIndex(header []string) map[string]int {
	indices := make(map[string]int, len(header))
	for i, col := range header {
		indices[strings.ToLower(strings.TrimSpace(col))] = i
	}
	return indices
}

func parseRecord(rec []string, idx map[string]int) (core.Tick, error) {
	get := func(key string) string {
		pos, ok := idx[key]
		if !ok || pos >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[pos])
	}

	tsStr := get("timestamp")
	if tsStr == "" {
		return core.Tick{}, fmt.Errorf("missing timestamp")
	}
	ts, err := time.Parse(CSVTimestampLayout, tsStr)
	if err != nil {
		return core.Tick{}, fmt.Errorf("bad timestamp: %w", err)
	}

	open, err := parseFloat(get("open"))
	if err != nil {
		return core.Tick{}, err
	}
	high, err := parseFloat(get("high"))
	if err != nil {
		return core.Tick{}, err
	}
	low, err := parseFloat(get("low"))
	if err != nil {
		return core.Tick{}, err
	}
	closePrice, err := parseFloat(get("close"))
	if err != nil {
		return core.Tick{}, err
	}
	volume, err := parseInt(get("volume"))
	if err != nil {
		return core.Tick{}, err
	}
	bidAskDelta, err := parseInt(get("bid_ask_delta"))
	if err != nil {
		return core.Tick{}, err
	}
	profile := parseVolumeProfile(get("volume_profile"))

	session := get("session")
	symbol := get("symbol")

	return core.Tick{
		Timestamp:     ts,
		Open:          open,
		High:          high,
		Low:           low,
		Close:         closePrice,
		Volume:        volume,
		BidAskDelta:   bidAskDelta,
		VolumeProfile: profile,
		Session:       session,
		Symbol:        symbol,
	}, nil
}

func parseFloat(raw string) (float64, error) {
	if raw == "" {
		return 0, fmt.Errorf("missing float")
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("bad float %q", raw)
	}
	return val, nil
}

func parseInt(raw string) (int64, error) {
	if raw == "" {
		return 0, fmt.Errorf("missing int")
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("bad int %q", raw)
	}
	return val, nil
}

// volume_profile format: price:vol|price:vol
func parseVolumeProfile(raw string) []core.PriceLevel {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	pairs := strings.Split(raw, "|")
	levels := make([]core.PriceLevel, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			continue
		}
		vol, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			continue
		}
		levels = append(levels, core.PriceLevel{Price: price, Volume: vol})
	}
	return levels
}
