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

func (s *Storage) AddProjectStats(projectID int64, opened int64, closed int64, total int64, dateExec time.Time) error {
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
		DateExec: dateExec,
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

func (s *Storage) AddGroupStats(groupID int64, opened int64, closed int64, total int64, dateExec time.Time) error {
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
		DateExec: dateExec,
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

func (s *Storage) GetStatsByProjectId6Months(projectID int64, beginDate carbon.Carbon, endDate carbon.Carbon) (openedSerie []float64, closedSerie []float64, dateExecSerie []time.Time, err error) {
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

func (s *Storage) GetStatsByGroupID6Months(groupID int64, beginDate carbon.Carbon, endDate carbon.Carbon) ([]float64, []float64, []time.Time, error) {
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
