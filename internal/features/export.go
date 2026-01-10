package features

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"trading-algo-generator/internal/core"
)

// Export writes features to a CSV file with a fixed column order.
func Export(path string, featureSets []core.FeatureSet, labelFunc func(idx int, fs core.FeatureSet) int) error {
	if len(featureSets) == 0 {
		return fmt.Errorf("no features to export")
	}

	columns := collectColumns(featureSets)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := append([]string{"timestamp"}, columns...)
	header = append(header, "label")
	if err := writer.Write(header); err != nil {
		return err
	}

	for i, fs := range featureSets {
		row := make([]string, 0, len(columns)+2)
		row = append(row, fs.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"))
		for _, col := range columns {
			row = append(row, fmt.Sprintf("%.6f", fs.Values[col]))
		}
		label := 0
		if labelFunc != nil {
			label = labelFunc(i, fs)
		}
		row = append(row, fmt.Sprintf("%d", label))
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return writer.Error()
}

func collectColumns(featureSets []core.FeatureSet) []string {
	colSet := make(map[string]struct{})
	for _, fs := range featureSets {
		for key := range fs.Values {
			colSet[key] = struct{}{}
		}
	}
	cols := make([]string, 0, len(colSet))
	for key := range colSet {
		cols = append(cols, key)
	}
	sort.Strings(cols)
	return cols
}
