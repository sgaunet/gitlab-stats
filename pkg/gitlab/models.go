package gitlab

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
