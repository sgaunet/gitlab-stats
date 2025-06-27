package graphissues

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	charts "github.com/vicanso/go-charts/v2"
)

func CreateGraph(graphFilePath string, openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time) error {
	var labels []string

	for r := range openedSerie {
		labels = append(labels, dateExecSerie[r].Format("2006-01"))
	}
	if len(openedSerie) != len(closedSerie) || len(openedSerie) != len(dateExecSerie) || len(closedSerie) != len(labels) {
		return errors.New("openedSerie, closedSerie and dateExecSerie should have the same length")
	}

	values := [][]float64{
		openedSerie,
	}
	charts.SetDefaultHeight(600)
	charts.SetDefaultWidth(1200)
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
		return err
	}
	buf, err := p.Bytes()
	if err != nil {
		return err
	}
	return writeFile(graphFilePath, buf)
}

// CreateEnhancedGraph creates a graph with 4 series: total opened, opened during period, closed during period, and velocity
func CreateEnhancedGraph(graphFilePath string, totalOpenedSeries []float64, openedDuringPeriod []float64, closedDuringPeriod []float64, velocitySeries []float64, dateExecSeries []time.Time) error {
	var labels []string

	for r := range totalOpenedSeries {
		labels = append(labels, dateExecSeries[r].Format("2006-01"))
	}
	
	// Validate all series have the same length
	seriesCount := len(totalOpenedSeries)
	if len(openedDuringPeriod) != seriesCount || len(closedDuringPeriod) != seriesCount || 
	   len(velocitySeries) != seriesCount || len(dateExecSeries) != seriesCount {
		return errors.New("all series should have the same length")
	}

	values := [][]float64{
		totalOpenedSeries,
		openedDuringPeriod,
		closedDuringPeriod,
		velocitySeries,
	}
	
	charts.SetDefaultHeight(600)
	charts.SetDefaultWidth(1200)
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
		return err
	}
	buf, err := p.Bytes()
	if err != nil {
		return err
	}
	return writeFile(graphFilePath, buf)
}

func writeFile(filename string, buf []byte) error {
	tmpPath := filepath.Dir(filename)
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, buf, 0644)
}
