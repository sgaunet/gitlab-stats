// Package gitlab provides GitLab API client functionality.
package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// GitlabAPIEndpoint is the default GitLab API endpoint.
const GitlabAPIEndpoint = "https://gitlab.com/api/v4"

// Service provides access to GitLab API.
type Service struct {
	gitlabAPIEndpoint string
	token             string
	httpClient        *http.Client
}

// NewService returns a new Service.
func NewService() *Service {
	return &Service{
		gitlabAPIEndpoint: GitlabAPIEndpoint,
		token:             os.Getenv("GITLAB_TOKEN"),
		httpClient:        &http.Client{},
	}
}

// SetGitlabEndpoint sets the Gitlab API endpoint
// default: https://gitlab.com/v4/api/
func (r *Service) SetGitlabEndpoint(gitlabAPIEndpoint string) {
	if gitlabAPIEndpoint != "" {
		r.gitlabAPIEndpoint = gitlabAPIEndpoint
	}
}

// SetToken sets the Gitlab API token
// default: GITLAB_TOKEN env variable
func (r *Service) SetToken(token string) {
	r.token = token
}

// SetHTTPClient sets the HTTP client to use for requests.
func (r *Service) SetHTTPClient(httpClient *http.Client) {
	r.httpClient = httpClient
}

// Get sends a GET request to the Gitlab API to the given path.
func (r *Service) Get(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", r.gitlabAPIEndpoint, path)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("Private-Token", r.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GET request: %w", err)
	}
	return resp, nil
}

// Post sends a POST request to the Gitlab API to the given path.
func (r *Service) Post(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", r.gitlabAPIEndpoint, path)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Private-Token", r.token)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute POST request: %w", err)
	}
	return resp, nil
}
