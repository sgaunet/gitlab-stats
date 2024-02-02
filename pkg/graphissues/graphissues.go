package graphissues

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	charts "github.com/vicanso/go-charts/v2"
)

func CreateGraph(graphFilePath string, openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time) error {
	var totalOpened []float64
	var openedInThePeriod []float64
	var closedDuringPeriod []float64
	var labels []string

	if len(openedSerie) != len(closedSerie) || len(openedSerie) != len(dateExecSerie) {
		return errors.New("openedSerie, closedSerie and dateExecSerie should have the same length")
	}
	for r := range openedSerie {
		if r != 0 {
			totalOpened = append(totalOpened, float64(openedSerie[r]))
			openedInThePeriod = append(openedInThePeriod, float64(openedSerie[r]-openedSerie[r-1]))
			closedDuringPeriod = append(closedDuringPeriod, float64(closedSerie[r]-closedSerie[r-1]))
			labels = append(labels, dateExecSerie[r].Format("2006-01"))
		}
	}

	values := [][]float64{
		totalOpened,
		openedInThePeriod,
		closedDuringPeriod,
	}
	charts.SetDefaultHeight(600)
	charts.SetDefaultWidth(1200)
	p, err := charts.LineRender(
		values,
		// charts.TitleTextOptionFunc("Line"),
		charts.XAxisDataOptionFunc(labels),
		charts.LegendLabelsOptionFunc([]string{
			"Total Opened",
			"Opened during period",
			"Closed during period",
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
