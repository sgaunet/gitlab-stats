package gitlabstatistics

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
)

// https://docs.gitlab.com/ee/api/issues_statistics.html
type ServiceStatistics struct {
	uri string
}

func NewServiceStatistics() *ServiceStatistics {
	r := ServiceStatistics{
		uri: "",
	}
	return &r
}

func (r *ServiceStatistics) SetProjectId(projectId int) {
	// list issues of a project : /projects/:id/issues
	r.uri = fmt.Sprintf("projects/%d/issues_statistics?", projectId)
}

func (r *ServiceStatistics) SetGroupId(groupId int) {
	// list issues of a group   : /groups/:id/issues
	r.uri = fmt.Sprintf("groups/%d/issues_statistics?", groupId)
}

func (r *ServiceStatistics) GetStatistics(gs *gitlab.GitlabService) (result Statistics, err error) {
	if r.uri == "" {
		return result, errors.New("no project or group specified")
	}
	resp, err := gs.Get(r.uri)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	return result, err
	// }
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}
