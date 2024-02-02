package jsonfile

import (
	"time"

	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
)

type DatabaseBFile struct {
	Records []databaseBFileRecord
}

type databaseBFileRecord struct {
	DateExec  time.Time     `json:"dateExec"`
	ProjectID int           `json:"projectId"`
	GroupID   int           `json:"groupId"`
	Counts    gitlab.Counts `json:"counts"`
}
