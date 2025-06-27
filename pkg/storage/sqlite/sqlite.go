// Package sqlite provides SQLite storage implementation for GitLab statistics.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite" // SQLite driver for dbmate
	"github.com/golang-module/carbon/v2"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/sgaunet/gitlab-stats/internal/database"
)

//go:embed db/migrations/*.sql
var fs embed.FS

// Storage provides SQLite-based storage for GitLab statistics.
type Storage struct {
	Now     func() time.Time
	db      *sql.DB
	dbFile  string
	queries *database.Queries
}

// NewStorage creates a new SQLite storage instance.
func NewStorage(dbFile string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &Storage{
		Now:     time.Now,
		db:      db,
		dbFile:  dbFile,
		queries: database.New(db),
	}, nil
}

// SetNow sets the function used to get the current time (useful for testing).
func (s *Storage) SetNow(now func() time.Time) {
	s.Now = now
}

// Close closes the database connection.
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

// Init initializes the database with migrations.
func (s *Storage) Init() error {
	u, _ := url.Parse("sqlite3://" + s.dbFile)
	db := dbmate.New(u)
	db.FS = fs

	fmt.Println("Migrations:")
	migrations, err := db.FindMigrations()
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}
	for _, m := range migrations {
		fmt.Println(m.Version, m.FilePath)
	}
	db.AutoDumpSchema = false
	err = db.CreateAndMigrate()
	if err != nil {
		return fmt.Errorf("failed to create and migrate database: %w", err)
	}
	return nil
}

// AddProjectStats adds statistics for a project.
func (s *Storage) AddProjectStats(
	projectID int64,
	opened int64,
	closed int64,
	total int64,
	dateExec *carbon.Carbon,
) error {
	// Check if project exists
	_, err := s.queries.GetProject(context.Background(), projectID)
	if errors.Is(err, sql.ErrNoRows) {
		// Create project
		_, err = s.queries.InsertNewProject(context.Background(), database.InsertNewProjectParams{
			ID:          projectID,
			ProjectName: "",
		})
		if err != nil {
			return fmt.Errorf("failed to insert new project: %w", err)
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
		return fmt.Errorf("failed to insert new stats: %w", err)
	}
	_, err = s.queries.InsertStatsProjects(context.Background(), database.InsertStatsProjectsParams{
		Statsid:   statsID,
		Projectid: projectID,
	})
	if err != nil {
		return fmt.Errorf("failed to insert stats projects: %w", err)
	}
	return nil
}

// AddGroupStats adds statistics for a group.
func (s *Storage) AddGroupStats(groupID int64, opened int64, closed int64, total int64, dateExec *carbon.Carbon) error {
	// Check if group exists
	group, err := s.queries.GetGroup(context.Background(), groupID)
	if errors.Is(err, sql.ErrNoRows) {
		// Create group
		groupID, err = s.queries.InsertNewGroup(context.Background(), database.InsertNewGroupParams{
			ID:        groupID,
			GroupName: "",
		})
		if err != nil {
			return fmt.Errorf("failed to insert new group: %w", err)
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
		return fmt.Errorf("failed to insert new stats: %w", err)
	}
	_, err = s.queries.InsertStatsGroups(context.Background(), database.InsertStatsGroupsParams{
		Statsid: statsID,
		Groupid: groupID,
	})
	if err != nil {
		return fmt.Errorf("failed to insert stats groups: %w", err)
	}
	return nil
}

// GetStatsByProjectID6Months gets project statistics for the last 6 months.
func (s *Storage) GetStatsByProjectID6Months(
	projectID int64,
	beginDate *carbon.Carbon,
	endDate *carbon.Carbon,
) ([]float64, []float64, []time.Time, error) {
	// Get project stats
	stats, err := s.queries.GetStatsByProjectID6Months(context.Background(), database.GetStatsByProjectID6MonthsParams{
		Projectid: projectID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get stats by project ID: %w", err)
	}
	openedSerie, closedSerie, dateExecSerie := getStatsProjects(stats)
	return openedSerie, closedSerie, dateExecSerie, nil
}

// GetStatsByGroupID6Months gets group statistics for the last 6 months.
func (s *Storage) GetStatsByGroupID6Months(
	groupID int64,
	beginDate *carbon.Carbon,
	endDate *carbon.Carbon,
) ([]float64, []float64, []time.Time, error) {
	// Get group stats
	stats, err := s.queries.GetStatsByGroupID6Months(context.Background(), database.GetStatsByGroupID6MonthsParams{
		Groupid:   groupID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get stats by group ID: %w", err)
	}
	opendSerie, closedSerie, dateExecSerie := getStatsGroups(stats)
	return opendSerie, closedSerie, dateExecSerie, nil
}

func getStatsGroups(stats []database.GetStatsByGroupID6MonthsRow) ([]float64, []float64, []time.Time) {
	openedSerie := make([]float64, 0, len(stats))
	closedSerie := make([]float64, 0, len(stats))
	dateExecSerie := make([]time.Time, 0, len(stats))
	for _, stat := range stats {
		openedSerie = append(openedSerie, float64(stat.Opened))
		closedSerie = append(closedSerie, float64(stat.Closed))
		dateExecSerie = append(dateExecSerie, stat.DateExec.In(time.UTC))
	}
	return openedSerie, closedSerie, dateExecSerie
}

func getStatsProjects(stats []database.GetStatsByProjectID6MonthsRow) ([]float64, []float64, []time.Time) {
	openedSerie := make([]float64, 0, len(stats))
	closedSerie := make([]float64, 0, len(stats))
	dateExecSerie := make([]time.Time, 0, len(stats))
	for _, stat := range stats {
		openedSerie = append(openedSerie, float64(stat.Opened))
		closedSerie = append(closedSerie, float64(stat.Closed))
		dateExecSerie = append(dateExecSerie, stat.DateExec.UTC())
	}
	return openedSerie, closedSerie, dateExecSerie
}

// EnhancedStats represents the four series for enhanced graphing.
type EnhancedStats struct {
	TotalOpenedSeries      []float64
	OpenedDuringPeriod     []float64
	ClosedDuringPeriod     []float64
	VelocitySeries         []float64
	DateExecSeries         []time.Time
}

// GetEnhancedStatsByProjectID gets enhanced project statistics with velocity calculations.
func (s *Storage) GetEnhancedStatsByProjectID(
	projectID int64,
	beginDate *carbon.Carbon,
	endDate *carbon.Carbon,
) (*EnhancedStats, error) {
	stats, err := s.queries.GetEnhancedStatsByProjectID(context.Background(), database.GetEnhancedStatsByProjectIDParams{
		Projectid: projectID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get enhanced stats by project ID: %w", err)
	}
	return processEnhancedStats(stats), nil
}

// GetEnhancedStatsByGroupID gets enhanced group statistics with velocity calculations.
func (s *Storage) GetEnhancedStatsByGroupID(
	groupID int64,
	beginDate *carbon.Carbon,
	endDate *carbon.Carbon,
) (*EnhancedStats, error) {
	stats, err := s.queries.GetEnhancedStatsByGroupID(context.Background(), database.GetEnhancedStatsByGroupIDParams{
		Groupid:   groupID,
		Begindate: beginDate.StdTime(),
		Enddate:   endDate.StdTime(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get enhanced stats by group ID: %w", err)
	}
	return processEnhancedStatsGroup(stats), nil
}

func processEnhancedStatsGeneric(
	totalOpened, currentOpened, currentClosed, prevTotal, prevClosed []interface{},
	dateExec []interface{},
) *EnhancedStats {
	count := len(totalOpened)
	totalOpenedSeries := make([]float64, 0, count)
	openedDuringPeriod := make([]float64, 0, count)
	closedDuringPeriod := make([]float64, 0, count)
	velocitySeries := make([]float64, 0, count)
	dateExecSeries := make([]time.Time, 0, count)

	for i := range totalOpened {
		// Convert interface{} to int64, handling potential nil values
		totalOpenedVal := convertToInt64(totalOpened[i])
		currentOpenedVal := convertToInt64(currentOpened[i])
		currentClosedVal := convertToInt64(currentClosed[i])
		prevTotalVal := convertToInt64(prevTotal[i])
		prevClosedVal := convertToInt64(prevClosed[i])
		
		// Calculate period metrics
		// opened during period = total new issues created in period = (current_total - prev_total)
		openedInPeriod := totalOpenedVal - prevTotalVal
		// closed during period = new closed issues in period = (current_closed - prev_closed)
		closedInPeriod := currentClosedVal - prevClosedVal
		// velocity = net change in open issues (positive = more open, negative = more closed)
		velocity := openedInPeriod - closedInPeriod

		totalOpenedSeries = append(totalOpenedSeries, float64(currentOpenedVal))  // Currently open issues
		openedDuringPeriod = append(openedDuringPeriod, float64(openedInPeriod))
		closedDuringPeriod = append(closedDuringPeriod, float64(closedInPeriod))
		velocitySeries = append(velocitySeries, float64(velocity))
		
		// Convert date
		dateTime := convertToTime(dateExec[i])
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

func processEnhancedStats(stats []database.GetEnhancedStatsByProjectIDRow) *EnhancedStats {
	count := len(stats)
	totalOpened := make([]interface{}, count)
	currentOpened := make([]interface{}, count)
	currentClosed := make([]interface{}, count)
	prevTotal := make([]interface{}, count)
	prevClosed := make([]interface{}, count)
	dateExec := make([]interface{}, count)
	
	for i, stat := range stats {
		totalOpened[i] = stat.TotalOpened
		currentOpened[i] = stat.CurrentOpened
		currentClosed[i] = stat.CurrentClosed
		prevTotal[i] = stat.PrevTotal
		prevClosed[i] = stat.PrevClosed
		dateExec[i] = stat.DateExec
	}
	
	return processEnhancedStatsGeneric(totalOpened, currentOpened, currentClosed, prevTotal, prevClosed, dateExec)
}

func processEnhancedStatsGroup(stats []database.GetEnhancedStatsByGroupIDRow) *EnhancedStats {
	count := len(stats)
	totalOpened := make([]interface{}, count)
	currentOpened := make([]interface{}, count)
	currentClosed := make([]interface{}, count)
	prevTotal := make([]interface{}, count)
	prevClosed := make([]interface{}, count)
	dateExec := make([]interface{}, count)
	
	for i, stat := range stats {
		totalOpened[i] = stat.TotalOpened
		currentOpened[i] = stat.CurrentOpened
		currentClosed[i] = stat.CurrentClosed
		prevTotal[i] = stat.PrevTotal
		prevClosed[i] = stat.PrevClosed
		dateExec[i] = stat.DateExec
	}
	
	return processEnhancedStatsGeneric(totalOpened, currentOpened, currentClosed, prevTotal, prevClosed, dateExec)
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
