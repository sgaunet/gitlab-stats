package graphissues

import (
	"os"
	"path/filepath"

	gitlabstatistics "github.com/sgaunet/gitlab-stats/pkg/gitlabStatistics"

	charts "github.com/vicanso/go-charts/v2"
)

func CreateGraph(graphFilePath string, records []gitlabstatistics.DatabaseBFileRecord) error {
	var totalOpened []float64
	var openedInThePeriod []float64
	var closedDuringPeriod []float64
	var labels []string

	for r := range records {
		if r != 0 {
			totalOpened = append(totalOpened, float64(records[r].Counts.Opened))
			openedInThePeriod = append(openedInThePeriod, float64(records[r].Counts.Opened-records[r-1].Counts.Opened))
			closedDuringPeriod = append(closedDuringPeriod, float64(records[r].Counts.Closed-records[r-1].Counts.Closed))
			labels = append(labels, records[r].DateExec.Format("2006-01"))
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
	err = writeFile(graphFilePath, buf)
	return err
}

func writeFile(filename string, buf []byte) error {
	tmpPath := filepath.Dir(filename)
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, buf, 0644)
	if err != nil {
		return err
	}
	return nil
}
