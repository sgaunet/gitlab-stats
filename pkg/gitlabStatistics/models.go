package gitlabstatistics

import "time"

type Statistics struct {
	Statistics Statistic `json:"statistics"`
}

type Statistic struct {
	Counts Counts `json:"counts"`
}

type Counts struct {
	All    int `json:"all,omitempty"`
	Closed int `json:"closed,omitempty"`
	Opened int `json:"opened,omitempty"`
}

type DatabaseBFile struct {
	Records []DatabaseBFileRecord
}

type DatabaseBFileRecord struct {
	DateExec  time.Time `json:"dateExec"`
	ProjectID int       `json:"projectId"`
	GroupID   int       `json:"groupId"`
	Counts    Counts    `json:"counts"`
}
