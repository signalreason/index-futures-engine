package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"trading-algo-generator/internal/core"
)

// TickStore stores ticks as JSON lines.
type TickStore struct {
	Path string
}

func (s TickStore) Append(tick core.Tick) error {
	file, err := os.OpenFile(s.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(&tick)
}

func (s TickStore) LoadAll() ([]core.Tick, error) {
	file, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var ticks []core.Tick
	for decoder.More() {
		var tick core.Tick
		if err := decoder.Decode(&tick); err != nil {
			return nil, fmt.Errorf("decode tick: %w", err)
		}
		ticks = append(ticks, tick)
	}
	return ticks, nil
}

func (s TickStore) Stream() (<-chan core.Tick, <-chan error) {
	out := make(chan core.Tick)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		file, err := os.Open(s.Path)
		if err != nil {
			errCh <- err
			return
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		dec := json.NewDecoder(reader)
		for dec.More() {
			var tick core.Tick
			if err := dec.Decode(&tick); err != nil {
				errCh <- err
				return
			}
			out <- tick
		}
	}()

	return out, errCh
}
