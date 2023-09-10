package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/golang-module/carbon/v2"
	gitlabstatistics "github.com/sgaunet/gitlab-stats/pkg/gitlabStatistics"
)

func main() {
	var projectID int
	var groupID int
	var dbFile string

	flag.IntVar(&projectID, "p", 0, "Project ID to get issues from")
	flag.IntVar(&groupID, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.StringVar(&dbFile, "db", "", "DB file (default /tmp/db.json))")
	flag.Parse()

	if dbFile == "" {
		dbFile = "/tmp/db.json"
	}

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
	s := gitlabstatistics.DatabaseBFile{}

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
		// append a new record
		s.Records = append(s.Records, gitlabstatistics.DatabaseBFileRecord{
			DateExec:  dbegin.ToStdTime(),
			ProjectID: projectID,
			GroupID:   groupID,
			Counts: gitlabstatistics.Counts{
				All:    allIssues,
				Closed: closedIssues,
				Opened: openIssues,
			},
		})
	}

	// convert to json
	dbContent, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = os.Remove(dbFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	// write to file
	f, err := os.OpenFile(dbFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer f.Close()
	_, err = f.Write(dbContent)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
