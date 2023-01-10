package gitlabstatistics

import "time"

type Statistics struct {
	Statistics Statistic `json:"statistics"`
}

type Statistic struct {
	Counts Counts `json:"counts"`
}

type Counts struct {
	All    int `json:"all"`
	Closed int `json:"closed"`
	Opened int `json:"opened"`
}

type structDBFile struct {
	Records []Record
}

type Record struct {
	DateExec  time.Time `json:"dateExec"`
	ProjectID int       `json:"projectId"`
	GroupID   int       `json:"groupId"`
	Counts    Counts    `json:"counts"`
}
