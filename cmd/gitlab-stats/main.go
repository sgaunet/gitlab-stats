package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
	"github.com/sgaunet/gitlab-stats/pkg/graphissues"
	"github.com/sgaunet/gitlab-stats/pkg/storage/jsonfile"
	"github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"

	// storage "github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"
	"github.com/sirupsen/logrus"
)

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var (
		debugLevel    string
		projectId     int
		groupId       int
		vOption       bool
		graphFilePath string
		dbFile        string
		// sinceMonth    int
		n *gitlab.ServiceStatistics
	)
	// Parameters treatment (except src + dest)
	flag.StringVar(&graphFilePath, "o", "", "file path to generate statistic graph (do not fullfill DB)")
	flag.StringVar(&debugLevel, "d", "error", "Debug level (info,warn,debug)")
	flag.StringVar(&dbFile, "db", os.Getenv("HOME")+"/.gitlab-stats/db.sqlite3", "DB file (default $HOME/.gitlab-stats/db.sqlite3))")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.IntVar(&projectId, "p", 0, "Project ID to get issues from")
	flag.IntVar(&groupId, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	// flag.IntVar(&sinceMonth, "s", 6, "graph last X month")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if debugLevel != "info" && debugLevel != "error" && debugLevel != "debug" {
		logrus.Errorf("debuglevel should be info or error or debug\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if projectId != 0 && groupId != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incompatible")
		flag.PrintDefaults()
		os.Exit(1)
	}
	initTrace(debugLevel)
	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		logrus.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}
	if len(os.Getenv("GITLAB_URI")) == 0 {
		os.Setenv("GITLAB_URI", "https://gitlab.com")
	}

	if groupId == 0 && projectId == 0 {
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

		logrus.Infoln("Project found: ", project.SshUrlToRepo)
		logrus.Infoln("Project found: ", project.Id)
		projectId = project.Id
	}

	// Check existence of DB file
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		logrus.Infoln("DB file not found, create it")
		// migrate DB
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
		defer s.Close()
		oldDbFile := strings.Replace(dbFile, ".sqlite3", ".json", 1)
		if _, err := os.Stat(oldDbFile); err == nil {
			logrus.Infoln("Migrate old DB file")
			oldJSONDB := jsonfile.NewDBStats(oldDbFile)
			err = s.MigrateDBFile(oldJSONDB)
			if err != nil {
				logrus.Errorln("error when migrating data: ", err.Error())
				os.Exit(1)
			}
		}
	}

	s, err := sqlite.NewStorage(dbFile)
	if err != nil {
		logrus.Errorln(err.Error())
		os.Exit(1)
	}

	if graphFilePath != "" {
		var openedSerie []float64
		var closedSerie []float64
		var dateExecSerie []time.Time
		var err error
		logrus.Infoln("retrieve stats from file")
		if projectId != 0 {
			openedSerie, closedSerie, dateExecSerie, err = s.GetStatsByProjectId6Months(int64(projectId))
		} else {
			openedSerie, closedSerie, dateExecSerie, err = s.GetStatsByGroupID6Months(int64(groupId))
		}
		if err != nil {
			logrus.Errorln("error when retrieving stats: ", err.Error())
			os.Exit(1)
		}
		err = graphissues.CreateGraph(graphFilePath, openedSerie, closedSerie, dateExecSerie)
		if err != nil {
			logrus.Errorln("error when creating file: ", err.Error())
			os.Exit(1)
		}
	} else {
		gs := gitlab.NewService()
		if projectId != 0 {
			n = gitlab.NewProjectStatistics(projectId)
		}
		if groupId != 0 {
			n = gitlab.NewGroupStatistics(groupId)
		}
		statistics, err := n.GetStatistics(gs)
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}
		if projectId != 0 {
			err = s.AddProjectStats(int64(projectId), int64(statistics.Statistics.Counts.Opened), int64(statistics.Statistics.Counts.Closed), int64(statistics.Statistics.Counts.All), time.Now())
		}
		if groupId != 0 {
			err = s.AddGroupStats(int64(groupId), int64(statistics.Statistics.Counts.Opened), int64(statistics.Statistics.Counts.Closed), int64(statistics.Statistics.Counts.All), time.Now())
		}
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}
	}
}

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})
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
