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

func writeFile(filename string, buf []byte) error {
	tmpPath := filepath.Dir(filename)
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, buf, 0644)
}
