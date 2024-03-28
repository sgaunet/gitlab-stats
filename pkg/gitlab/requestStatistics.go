package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// https://docs.gitlab.com/ee/api/issues_statistics.html
type ServiceStatistics struct {
	uri string
}

func NewProjectStatistics(projectId int) *ServiceStatistics {
	r := ServiceStatistics{
		uri: fmt.Sprintf("projects/%d/issues_statistics?", projectId),
	}
	return &r
}

func NewGroupStatistics(groupId int) *ServiceStatistics {
	r := ServiceStatistics{
		uri: fmt.Sprintf("groups/%d/issues_statistics?", groupId),
	}
	return &r
}

func (r *ServiceStatistics) GetStatistics(gs *GitlabService) (result Statistics, err error) {
	// if r.uri == "" {
	// 	return result, errors.New("no project or group specified")
	// }
	resp, err := gs.Get(r.uri)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// !TODO: handle 404
		return result, fmt.Errorf("status code: %d", resp.StatusCode)
	}
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
