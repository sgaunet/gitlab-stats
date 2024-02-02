package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/golang-module/carbon/v2"
	"github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"
)

func main() {
	var projectID int
	var groupID int
	var dbFile string

	flag.IntVar(&projectID, "p", 0, "Project ID to get issues from")
	flag.IntVar(&groupID, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.StringVar(&dbFile, "db", "/tmp/db.sqlite3", "DB file (default /tmp/db.sqlite3))")
	flag.Parse()

	if projectID != 0 && groupID != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incompatible")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if projectID == 0 && groupID == 0 {
		fmt.Fprintln(os.Stderr, "-p or -g option is mandatory")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// For the last 12 months, generate a fake $HOME/.gitlab-stats/db.json file
	s, err := sqlite.NewStorage(dbFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer s.Close()
	err = s.Init()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Generate fake data for the last 12 months
	dbegin := carbon.Now().SubMonths(12).StartOfMonth()
	dend := carbon.Now().StartOfMonth()

	// init with 20 open issues and 10 closed issues
	openIssues := 20
	closedIssues := 10

	// for the last 12 months
	for dbegin.Compare("<", dend) {
		// fmt.Println(dbegin.String())
		dbegin = dbegin.AddMonth()
		// add a random number of issues between 0 and 10
		openIssues += 0 + rand.Intn(10)
		// same for closed issues
		closedIssues += 0 + rand.Intn(10)
		fmt.Printf("openIssues: %d, closedIssues: %d\n", openIssues, closedIssues)
		allIssues := openIssues + closedIssues
		// append stats
		if projectID != 0 {
			err = s.AddProjectStats(int64(projectID), int64(openIssues), int64(closedIssues), int64(allIssues), dbegin.ToStdTime())
		}
		if groupID != 0 {
			err = s.AddGroupStats(int64(groupID), int64(openIssues), int64(closedIssues), int64(allIssues), dbegin.ToStdTime())
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}
