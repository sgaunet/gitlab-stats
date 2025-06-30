// Package main provides test data generation for GitLab stats database.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/golang-module/carbon/v2"
	"github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"
)

type config struct {
	projectID int
	groupID   int
	dbFile    string
}

func parseFlags() config {
	var cfg config
	flag.IntVar(&cfg.projectID, "p", 0, "Project ID to get issues from")
	flag.IntVar(&cfg.groupID, "g", 0, "Group ID to get issues from (not compatible with -p option)")
	flag.StringVar(&cfg.dbFile, "db", "/tmp/db.sqlite3", "DB file (default /tmp/db.sqlite3))")
	flag.Parse()
	return cfg
}

func validateConfig(cfg config) {
	if cfg.projectID != 0 && cfg.groupID != 0 {
		fmt.Fprintln(os.Stderr, "-p and -g option are incompatible")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if cfg.projectID == 0 && cfg.groupID == 0 {
		fmt.Fprintln(os.Stderr, "-p or -g option is mandatory")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func initStorage(dbFile string) *sqlite.Storage {
	s, err := sqlite.NewStorage(dbFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = s.Init()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return s
}

func generateFakeData(s *sqlite.Storage, cfg config) {
	const monthsBack = 12
	dbegin := carbon.Now().SubMonths(monthsBack).StartOfMonth()
	dend := carbon.Now().StartOfMonth()

	openIssues := 20
	closedIssues := 10

	for dbegin.Compare("<", dend) {
		openIssues += rand.Intn(10)
		closedIssues += rand.Intn(10)
		fmt.Printf("openIssues: %d, closedIssues: %d\n", openIssues, closedIssues)
		allIssues := openIssues + closedIssues

		var err error
		if cfg.projectID != 0 {
			err = s.AddProjectStats(int64(cfg.projectID), int64(openIssues), int64(closedIssues), int64(allIssues), dbegin)
		} else {
			err = s.AddGroupStats(int64(cfg.groupID), int64(openIssues), int64(closedIssues), int64(allIssues), dbegin)
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		dbegin = dbegin.AddMonth()
	}
}

func main() {
	cfg := parseFlags()
	validateConfig(cfg)
	s := initStorage(cfg.dbFile)
	defer func() {
		if err := s.Close(); err != nil {
			fmt.Printf("Error closing storage: %v\n", err)
		}
	}()
	generateFakeData(s, cfg)
}
