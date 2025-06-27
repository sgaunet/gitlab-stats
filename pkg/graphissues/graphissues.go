// Package graphissues provides chart generation functionality for GitLab statistics.
package graphissues

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	charts "github.com/vicanso/go-charts/v2"
)

const (
	defaultHeight = 600
	defaultWidth  = 1200
	dirPerm       = 0o700
	filePerm      = 0o600
)

var (
	// ErrSeriesLengthMismatch is returned when input series have different lengths.
	ErrSeriesLengthMismatch    = errors.New("openedSerie, closedSerie and dateExecSerie should have the same length")
	// ErrAllSeriesLengthMismatch is returned when enhanced graph series have different lengths.
	ErrAllSeriesLengthMismatch = errors.New("all series should have the same length")
)

// CreateGraph creates a simple line chart from the provided data series.
func CreateGraph(graphFilePath string, openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time) error {
	labels := make([]string, 0, len(openedSerie))

	for r := range openedSerie {
		labels = append(labels, dateExecSerie[r].Format("2006-01"))
	}
	if len(openedSerie) != len(closedSerie) || len(openedSerie) != len(dateExecSerie) || len(closedSerie) != len(labels) {
		return ErrSeriesLengthMismatch
	}

	values := [][]float64{
		openedSerie,
	}
	charts.SetDefaultHeight(defaultHeight)
	charts.SetDefaultWidth(defaultWidth)
	p, err := charts.LineRender(
		values,
		// charts.TitleTextOptionFunc("Line"),
		charts.XAxisDataOptionFunc(labels),
		charts.LegendLabelsOptionFunc([]string{
			"Opened issues",
			// "Closed issues",
		}, charts.PositionCenter),
	)
	if err != nil {
		return fmt.Errorf("failed to render line chart: %w", err)
	}
	buf, err := p.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get chart bytes: %w", err)
	}
	return writeFile(graphFilePath, buf)
}

// CreateEnhancedGraph creates a graph with 4 series: total opened, opened during period,
// closed during period, and velocity.
func CreateEnhancedGraph(
	graphFilePath string,
	totalOpenedSeries []float64,
	openedDuringPeriod []float64,
	closedDuringPeriod []float64,
	velocitySeries []float64,
	dateExecSeries []time.Time,
) error {
	labels := make([]string, 0, len(totalOpenedSeries))

	for r := range totalOpenedSeries {
		labels = append(labels, dateExecSeries[r].Format("2006-01"))
	}
	
	// Validate all series have the same length
	seriesCount := len(totalOpenedSeries)
	if len(openedDuringPeriod) != seriesCount || len(closedDuringPeriod) != seriesCount || 
	   len(velocitySeries) != seriesCount || len(dateExecSeries) != seriesCount {
		return ErrAllSeriesLengthMismatch
	}

	values := [][]float64{
		totalOpenedSeries,
		openedDuringPeriod,
		closedDuringPeriod,
		velocitySeries,
	}
	
	charts.SetDefaultHeight(defaultHeight)
	charts.SetDefaultWidth(defaultWidth)
	p, err := charts.LineRender(
		values,
		charts.TitleTextOptionFunc("GitLab Issues Statistics"),
		charts.XAxisDataOptionFunc(labels),
		charts.LegendLabelsOptionFunc([]string{
			"Currently Open Issues",
			"Issues Opened This Period",
			"Issues Closed This Period",
			"Velocity (Net Change)",
		}, charts.PositionCenter),
	)
	if err != nil {
		return fmt.Errorf("failed to render line chart: %w", err)
	}
	buf, err := p.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get chart bytes: %w", err)
	}
	return writeFile(graphFilePath, buf)
}

func writeFile(filename string, buf []byte) error {
	tmpPath := filepath.Dir(filename)
	err := os.MkdirAll(tmpPath, dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(filename, buf, filePerm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
