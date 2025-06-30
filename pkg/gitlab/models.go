package gitlab

// Statistics represents the top-level response from GitLab statistics API.
type Statistics struct {
	Statistics Statistic `json:"statistics"`
}

// Statistic represents the statistics data structure.
type Statistic struct {
	Counts Counts `json:"counts"`
}

// Counts represents the issue counts (all, closed, opened).
type Counts struct {
	All    int `json:"all,omitempty"`
	Closed int `json:"closed,omitempty"`
	Opened int `json:"opened,omitempty"`
}
