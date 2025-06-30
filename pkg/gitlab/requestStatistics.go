package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const httpOK = 200

// ErrNon200Response is returned when the HTTP response status is not 200.
var ErrNon200Response = errors.New("non-200 response code")

// ServiceStatistics provides access to GitLab issue statistics API.
// See: https://docs.gitlab.com/ee/api/issues_statistics.html
type ServiceStatistics struct {
	uri string
}

// NewProjectStatistics creates a new ServiceStatistics for a project.
func NewProjectStatistics(projectID int) *ServiceStatistics {
	r := ServiceStatistics{
		uri: fmt.Sprintf("projects/%d/issues_statistics?", projectID),
	}
	return &r
}

// NewGroupStatistics creates a new ServiceStatistics for a group.
func NewGroupStatistics(groupID int) *ServiceStatistics {
	r := ServiceStatistics{
		uri: fmt.Sprintf("groups/%d/issues_statistics?", groupID),
	}
	return &r
}

// GetStatistics retrieves statistics from GitLab API.
func (r *ServiceStatistics) GetStatistics(gs *Service) (Statistics, error) {
	// if r.uri == "" {
	// 	return result, errors.New("no project or group specified")
	// }
	resp, err := gs.Get(r.uri)
	if err != nil {
		return Statistics{}, err
	}
	defer func() {
		_ = resp.Body.Close() // ignore error
	}()
	if resp.StatusCode != httpOK {
		// !TODO: handle 404
		return Statistics{}, fmt.Errorf("%w: %d", ErrNon200Response, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Statistics{}, fmt.Errorf("failed to read response body: %w", err)
	}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	return result, err
	// }
	var result Statistics
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err != nil {
		return Statistics{}, fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return result, nil
}
