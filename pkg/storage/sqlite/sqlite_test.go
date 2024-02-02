package sqlite_test

import (
	"os"
	"testing"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sgaunet/gitlab-stats/pkg/storage/jsonfile"
	"github.com/sgaunet/gitlab-stats/pkg/storage/sqlite"
)

func TestNewStorage(t *testing.T) {
	s, err := sqlite.NewStorage("/tmp/db.sqlite3")
	if s == nil {
		t.Errorf("NewStorage() returned an empty storage")
	}
	if err != nil {
		t.Errorf("err returned by NewStorage(): %v", err.Error())
	}
	// delete db file
	os.Remove("/tmp/db.sqlite3")
}

func TestClose(t *testing.T) {
	s, _ := sqlite.NewStorage("/tmp/db.sqlite3")
	err := s.Close()
	if err != nil {
		t.Errorf("err returned by Close(): %v", err.Error())
	}
	// delete db file
	os.Remove("/tmp/db.sqlite3")
}

func TestInit(t *testing.T) {
	s, _ := sqlite.NewStorage("/tmp/db.sqlite3")
	err := s.Init()
	if err != nil {
		t.Errorf("err returned by Init(): %v", err.Error())
	}
	// delete db file
	os.Remove("/tmp/db.sqlite3")
}

func TestMigrate(t *testing.T) {
	defer os.Remove("/tmp/db.sqlite3")
	s, _ := sqlite.NewStorage("/tmp/db.sqlite3")
	err := s.Init()
	if err != nil {
		t.Errorf("err returned by Init(): %v", err.Error())
	}
	jsonDB := jsonfile.NewDBStats(os.Getenv("HOME") + "/.gitlab-stats/db.json")
	err = s.MigrateDBFile(jsonDB)
	if err != nil {
		t.Errorf("err returned by MigrateDBFile(): %v", err.Error())
	}
}
