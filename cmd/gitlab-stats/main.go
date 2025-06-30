package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang-module/carbon/v2"
	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
	"github.com/sgaunet/gitlab-stats/pkg/graphissues"
	"github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"

	// storage "github.com/sgaunet/gitlab-stats/pkg/storage/sqlite".
	"github.com/sirupsen/logrus"
)

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate -f ../../sqlc.yaml

var version = "development"

func printVersion() {
	fmt.Println(version)
}

type config struct {
	debugLevel    string
	projectID     int
	groupID       int
	vOption       bool
	graphFilePath string
	dbFile        string
	sinceMonth    int
}

func parseAndValidateFlags() config {
	cfg := config{}
	
	// Parameters treatment
	flag.StringVar(&cfg.graphFilePath, "o", "", "file path to generate statistic graph (do not fulfill DB)")
	flag.StringVar(&cfg.debugLevel, "d", "error", "Debug level (info,warn,debug)")
	defaultDBFile := os.Getenv("HOME") + "/.gitlab-stats/db.sqlite3"
	flag.StringVar(&cfg.dbFile, "db", defaultDBFile, "DB file (default $HOME/.gitlab-stats/db.sqlite3))")
	flag.BoolVar(&cfg.vOption, "v", false, "Get version")
	flag.IntVar(&cfg.projectID, "p", 0, "Project ID to get issues from")
	flag.IntVar(&cfg.groupID, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	const defaultSinceMonths = 6
	flag.IntVar(&cfg.sinceMonth, "s", defaultSinceMonths, "graph last X month")
	flag.Parse()

	if cfg.vOption {
		printVersion()
		os.Exit(0)
	}

	validateConfig(cfg)
	return cfg
}

func validateConfig(cfg config) {
	if cfg.sinceMonth < 1 {
		logrus.Errorf("sinceMonth should be greater than 0\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if cfg.debugLevel != "info" && cfg.debugLevel != "error" && cfg.debugLevel != "debug" {
		logrus.Errorf("debuglevel should be info or error or debug\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	if cfg.projectID != 0 && cfg.groupID != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incompatible")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func setupEnvironment() {
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		logrus.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}
	
	if len(os.Getenv("GITLAB_URI")) == 0 {
		if err := os.Setenv("GITLAB_URI", "https://gitlab.com"); err != nil {
			logrus.Warnf("Failed to set GITLAB_URI: %v", err)
		}
	}
}

func detectProjectIfNeeded(cfg *config) {
	if cfg.groupID != 0 || cfg.projectID != 0 {
		return
	}
	
	// Try to find git repository and project
	gitFolder, err := findGitRepository()
	if err != nil {
		logrus.Errorf("Folder .git not found")
		os.Exit(1)
	}
	
	remoteOrigin := GetRemoteOrigin(gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config")
	project, err := findProject(remoteOrigin)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	logrus.Infoln("Project found: ", project.SSHURLToRepo)
	logrus.Infoln("Project found: ", project.ID)
	cfg.projectID = project.ID
}

func initializeDatabase(dbFile string) *sqlite.Storage {
	// Check existence of DB file
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		logrus.Infoln("DB file not found, create it")
		s, err := sqlite.NewStorage(dbFile)
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}
		err = s.Init()
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}
		if err := s.Close(); err != nil {
			logrus.Warnf("Failed to close storage: %v", err)
		}
	}

	s, err := sqlite.NewStorage(dbFile)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
	return s
}

func generateGraph(s *sqlite.Storage, cfg config) {
	begindate := carbon.CreateFromStdTime(s.Now()).AddMonths(-cfg.sinceMonth).StartOfMonth()
	enddate := carbon.CreateFromStdTime(s.Now()).StartOfMonth()
	
	logrus.Infoln("retrieve enhanced stats from database")
	var enhancedStats *sqlite.EnhancedStats
	var err error
	
	if cfg.projectID != 0 {
		enhancedStats, err = s.GetEnhancedStatsByProjectID(int64(cfg.projectID), begindate, enddate)
	} else {
		enhancedStats, err = s.GetEnhancedStatsByGroupID(int64(cfg.groupID), begindate, enddate)
	}
	if err != nil {
		logrus.Errorln("error when retrieving enhanced stats: ", err.Error())
		os.Exit(1)
	}
	
	err = graphissues.CreateEnhancedGraph(
		cfg.graphFilePath, 
		enhancedStats.TotalOpenedSeries,
		enhancedStats.OpenedDuringPeriod,
		enhancedStats.ClosedDuringPeriod,
		enhancedStats.VelocitySeries,
		enhancedStats.DateExecSeries,
	)
	if err != nil {
		logrus.Errorln("error when creating enhanced graph: ", err.Error())
		os.Exit(1)
	}
}

func collectData(s *sqlite.Storage, cfg config) {
	gs := gitlab.NewService()
	var n *gitlab.ServiceStatistics
	
	if cfg.projectID != 0 {
		n = gitlab.NewProjectStatistics(cfg.projectID)
	} else {
		n = gitlab.NewGroupStatistics(cfg.groupID)
	}
	
	statistics, err := n.GetStatistics(gs)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
	
	if cfg.projectID != 0 {
		err = s.AddProjectStats(
			int64(cfg.projectID), 
			int64(statistics.Statistics.Counts.Opened), 
			int64(statistics.Statistics.Counts.Closed), 
			int64(statistics.Statistics.Counts.All), 
			carbon.Now(),
		)
	} else {
		err = s.AddGroupStats(
			int64(cfg.groupID), 
			int64(statistics.Statistics.Counts.Opened), 
			int64(statistics.Statistics.Counts.Closed), 
			int64(statistics.Statistics.Counts.All), 
			carbon.Now(),
		)
	}
	
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}
}

func main() {
	cfg := parseAndValidateFlags()
	initTrace(cfg.debugLevel)
	setupEnvironment()
	detectProjectIfNeeded(&cfg)
	
	s := initializeDatabase(cfg.dbFile)
	
	if cfg.graphFilePath != "" {
		generateGraph(s, cfg)
	} else {
		collectData(s, cfg)
	}
}

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	switch debugLevel {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}
