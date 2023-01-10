package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	gitlabstatistics "github.com/sgaunet/gitlab-stats/gitlabStatistics"
	"github.com/sgaunet/gitlab-stats/graphissues"
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
		sinceMonth    int
	)
	// Parameters treatment (except src + dest)
	flag.StringVar(&graphFilePath, "o", "", "file path to generate statistic graph (do not fullfill DB)")
	flag.StringVar(&debugLevel, "d", "error", "Debug level (info,warn,debug)")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.IntVar(&projectId, "p", 0, "Project ID to get issues from")
	flag.IntVar(&groupId, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.IntVar(&sinceMonth, "s", 6, "graph last X month")
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

	n := gitlabstatistics.NewRequestStatistics()
	if projectId != 0 {
		n.SetProjectId(projectId)
	}
	if groupId != 0 {
		n.SetGroupId(groupId)
	}

	if graphFilePath != "" {
		logrus.Infoln("retrieve stats from file")
		r, err := gitlabstatistics.GetLastMonthStatsFromFile(sinceMonth)
		if err != nil {
			logrus.Errorln("error when retrieving data: ", err.Error())
			os.Exit(1)
		}
		if projectId != 0 {
			r = gitlabstatistics.FilterWithProject(r, projectId)
		}
		if groupId != 0 {
			r = gitlabstatistics.FilterWithGroup(r, groupId)
		}
		err = graphissues.CreateGraph(graphFilePath, r)
		if err != nil {
			logrus.Errorln("error when creating file: ", err.Error())
			os.Exit(1)
		}
	} else {
		statistics, err := n.GetStatistics()
		if err != nil {
			logrus.Errorln(err.Error())
			os.Exit(1)
		}
		newRecord := gitlabstatistics.Record{
			DateExec:  time.Now(),
			Counts:    statistics.Statistics.Counts,
			GroupID:   groupId,
			ProjectID: projectId,
		}
		err = gitlabstatistics.AppendStatsToFile(newRecord)
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
