package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"github.com/golang-module/carbon/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sgaunet/gitlab-stats/internal/database"
)

//go:embed db/migrations/*.sql
var fs embed.FS

type Storage struct {
	Now     func() time.Time
	db      *sql.DB
	dbFile  string
	queries *database.Queries
}

func NewStorage(dbFile string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Now:     time.Now,
		db:      db,
		dbFile:  dbFile,
		queries: database.New(db),
	}, nil
}

func (s *Storage) SetNow(now func() time.Time) {
	s.Now = now
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Init() error {
	u, _ := url.Parse(fmt.Sprintf("sqlite3://%s", s.dbFile))
	db := dbmate.New(u)
	db.FS = fs

	fmt.Println("Migrations:")
	migrations, err := db.FindMigrations()
	if err != nil {
		return err
	}
	for _, m := range migrations {
		fmt.Println(m.Version, m.FilePath)
	}
	db.AutoDumpSchema = false
	err = db.CreateAndMigrate()
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddProjectStats(projectID int64, opened int64, closed int64, total int64, dateExec *carbon.Carbon) error {
	// Check if project exists
	_, err := s.queries.GetProject(context.Background(), projectID)
	if err == sql.ErrNoRows {
		// Create project
		_, err = s.queries.InsertNewProject(context.Background(), database.InsertNewProjectParams{
			ID:          projectID,
			ProjectName: "",
		})
		if err != nil {
			return err
		}
	}
	// Add Stats
	statsID, err := s.queries.InsertNewStats(context.Background(), database.InsertNewStatsParams{
		Total:    total,
		Closed:   closed,
		Opened:   opened,
		DateExec: dateExec.StdTime(),
	})
	if err != nil {
		return err
	}
	_, err = s.queries.InsertStatsProjects(context.Background(), database.InsertStatsProjectsParams{
		Statsid:   statsID,
		Projectid: projectID,
	})
	return err
}

func (s *Storage) AddGroupStats(groupID int64, opened int64, closed int64, total int64, dateExec *carbon.Carbon) error {
	// Check if group exists
	group, err := s.queries.GetGroup(context.Background(), groupID)
	if err == sql.ErrNoRows {
		// Create group
		groupID, err = s.queries.InsertNewGroup(context.Background(), database.InsertNewGroupParams{
			ID:        groupID,
			GroupName: "",
		})
		if err != nil {
			return err
		}
	} else {
		groupID = group.ID
	}
	// Add Stats
	statsID, err := s.queries.InsertNewStats(context.Background(), database.InsertNewStatsParams{
		Total:    total,
		Closed:   closed,
		Opened:   opened,
		DateExec: dateExec.StdTime(),
	})
	if err != nil {
		return err
	}
	_, err = s.queries.InsertStatsGroups(context.Background(), database.InsertStatsGroupsParams{
		Statsid: statsID,
		Groupid: groupID,
	})
	return err
}

func (s *Storage) GetStatsByProjectId6Months(projectID int64, beginDate *carbon.Carbon, endDate *carbon.Carbon) (openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time, err error) {
	// Get project stats
	stats, err := s.queries.GetStatsByProjectID6Months(context.Background(), database.GetStatsByProjectID6MonthsParams{
		Projectid: projectID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, nil, nil, err
	}
	opendSerie, closedSerie, dateExecSerie := getStatsProjects(stats)
	return opendSerie, closedSerie, dateExecSerie, nil
}

func (s *Storage) GetStatsByGroupID6Months(groupID int64, beginDate *carbon.Carbon, endDate *carbon.Carbon) ([]float64, []float64, []time.Time, error) {
	// Get group stats
	stats, err := s.queries.GetStatsByGroupID6Months(context.Background(), database.GetStatsByGroupID6MonthsParams{
		Groupid:   groupID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, nil, nil, err
	}
	opendSerie, closedSerie, dateExecSerie := getStatsGroups(stats)
	return opendSerie, closedSerie, dateExecSerie, nil
}

func getStatsGroups(stats []database.GetStatsByGroupID6MonthsRow) ([]float64, []float64, []time.Time) {
	var openedSerie []float64
	var closedSerie []float64
	var dateExecSerie []time.Time
	for _, stat := range stats {
		openedSerie = append(openedSerie, float64(stat.Opened))
		closedSerie = append(closedSerie, float64(stat.Closed))
		dateExecSerie = append(dateExecSerie, stat.DateExec.In(time.UTC))
	}
	return openedSerie, closedSerie, dateExecSerie
}

func getStatsProjects(stats []database.GetStatsByProjectID6MonthsRow) ([]float64, []float64, []time.Time) {
	var openedSerie []float64
	var closedSerie []float64
	var dateExecSerie []time.Time
	for _, stat := range stats {
		openedSerie = append(openedSerie, float64(stat.Opened))
		closedSerie = append(closedSerie, float64(stat.Closed))
		dateExecSerie = append(dateExecSerie, stat.DateExec.UTC())
	}
	return openedSerie, closedSerie, dateExecSerie
}

// EnhancedStats represents the four series for enhanced graphing
type EnhancedStats struct {
	TotalOpenedSeries      []float64
	OpenedDuringPeriod     []float64
	ClosedDuringPeriod     []float64
	VelocitySeries         []float64
	DateExecSeries         []time.Time
}

func (s *Storage) GetEnhancedStatsByProjectID(projectID int64, beginDate *carbon.Carbon, endDate *carbon.Carbon) (*EnhancedStats, error) {
	stats, err := s.queries.GetEnhancedStatsByProjectID(context.Background(), database.GetEnhancedStatsByProjectIDParams{
		Projectid: projectID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, err
	}
	return processEnhancedStats(stats), nil
}

func (s *Storage) GetEnhancedStatsByGroupID(groupID int64, beginDate *carbon.Carbon, endDate *carbon.Carbon) (*EnhancedStats, error) {
	stats, err := s.queries.GetEnhancedStatsByGroupID(context.Background(), database.GetEnhancedStatsByGroupIDParams{
		Groupid:   groupID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, err
	}
	return processEnhancedStatsGroup(stats), nil
}

func processEnhancedStats(stats []database.GetEnhancedStatsByProjectIDRow) *EnhancedStats {
	var totalOpenedSeries []float64
	var openedDuringPeriod []float64
	var closedDuringPeriod []float64
	var velocitySeries []float64
	var dateExecSeries []time.Time

	for _, stat := range stats {
		// Convert interface{} to int64, handling potential nil values
		totalOpened := convertToInt64(stat.TotalOpened)
		currentOpened := convertToInt64(stat.CurrentOpened)
		currentClosed := convertToInt64(stat.CurrentClosed)
		prevTotal := convertToInt64(stat.PrevTotal)
		prevClosed := convertToInt64(stat.PrevClosed)
		
		// Calculate period metrics
		// opened during period = total new issues created in period = (current_total - prev_total)
		openedInPeriod := totalOpened - prevTotal
		// closed during period = new closed issues in period = (current_closed - prev_closed)
		closedInPeriod := currentClosed - prevClosed
		// velocity = net change in open issues (positive = more open, negative = more closed)
		velocity := openedInPeriod - closedInPeriod

		totalOpenedSeries = append(totalOpenedSeries, float64(currentOpened))  // Currently open issues
		openedDuringPeriod = append(openedDuringPeriod, float64(openedInPeriod))
		closedDuringPeriod = append(closedDuringPeriod, float64(closedInPeriod))
		velocitySeries = append(velocitySeries, float64(velocity))
		
		// Convert date
		dateTime := convertToTime(stat.DateExec)
		if !dateTime.IsZero() {
			dateExecSeries = append(dateExecSeries, dateTime.UTC())
		}
	}

	return &EnhancedStats{
		TotalOpenedSeries:      totalOpenedSeries,
		OpenedDuringPeriod:     openedDuringPeriod,
		ClosedDuringPeriod:     closedDuringPeriod,
		VelocitySeries:         velocitySeries,
		DateExecSeries:         dateExecSeries,
	}
}

func processEnhancedStatsGroup(stats []database.GetEnhancedStatsByGroupIDRow) *EnhancedStats {
	var totalOpenedSeries []float64
	var openedDuringPeriod []float64
	var closedDuringPeriod []float64
	var velocitySeries []float64
	var dateExecSeries []time.Time

	for _, stat := range stats {
		// Convert interface{} to int64, handling potential nil values
		totalOpened := convertToInt64(stat.TotalOpened)
		currentOpened := convertToInt64(stat.CurrentOpened)
		currentClosed := convertToInt64(stat.CurrentClosed)
		prevTotal := convertToInt64(stat.PrevTotal)
		prevClosed := convertToInt64(stat.PrevClosed)
		
		// Calculate period metrics
		// opened during period = total new issues created in period = (current_total - prev_total)
		openedInPeriod := totalOpened - prevTotal
		// closed during period = new closed issues in period = (current_closed - prev_closed)
		closedInPeriod := currentClosed - prevClosed
		// velocity = net change in open issues (positive = more open, negative = more closed)
		velocity := openedInPeriod - closedInPeriod

		totalOpenedSeries = append(totalOpenedSeries, float64(currentOpened))  // Currently open issues
		openedDuringPeriod = append(openedDuringPeriod, float64(openedInPeriod))
		closedDuringPeriod = append(closedDuringPeriod, float64(closedInPeriod))
		velocitySeries = append(velocitySeries, float64(velocity))
		
		// Convert date
		dateTime := convertToTime(stat.DateExec)
		if !dateTime.IsZero() {
			dateExecSeries = append(dateExecSeries, dateTime.UTC())
		}
	}

	return &EnhancedStats{
		TotalOpenedSeries:      totalOpenedSeries,
		OpenedDuringPeriod:     openedDuringPeriod,
		ClosedDuringPeriod:     closedDuringPeriod,
		VelocitySeries:         velocitySeries,
		DateExecSeries:         dateExecSeries,
	}
}

func convertToInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func convertToTime(val interface{}) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		// Try to parse common SQLite datetime formats
		formats := []string{
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05+00:00",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
		return time.Time{}
	default:
		return time.Time{}
	}
}
