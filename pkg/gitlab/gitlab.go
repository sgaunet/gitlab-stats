package gitlab

import (
	"fmt"
	"net/http"
	"os"
)

const GitlabApiEndpoint = "https://gitlab.com/api/v4"

type GitlabService struct {
	gitlabApiEndpoint string
	token             string
	httpClient        *http.Client
}

// NewRequest returns a new GitlabService
func NewService() *GitlabService {
	return &GitlabService{
		gitlabApiEndpoint: GitlabApiEndpoint,
		token:             os.Getenv("GITLAB_TOKEN"),
		httpClient:        &http.Client{},
	}
}

// SetGitlabEndpoint sets the Gitlab API endpoint
// default: https://gitlab.com/v4/api/
func (r *GitlabService) SetGitlabEndpoint(gitlabApiEndpoint string) {
	if gitlabApiEndpoint != "" {
		r.gitlabApiEndpoint = gitlabApiEndpoint
	}
}

// SetToken sets the Gitlab API token
// default: GITLAB_TOKEN env variable
func (r *GitlabService) SetToken(token string) {
	r.token = token
}

// SetHttpClient sets the http client
// default: http.Client{}
func (r *GitlabService) SetHttpClient(httpClient *http.Client) {
	r.httpClient = httpClient
}

// Get sends a GET request to the Gitlab API to the given path
func (r *GitlabService) Get(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", r.gitlabApiEndpoint, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", r.token)
	req.Header.Set("Content-Type", "application/json")
	return r.httpClient.Do(req)
}

// Post sends a POST request to the Gitlab API to the given path
func (r *GitlabService) Post(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", r.gitlabApiEndpoint, path)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", r.token)
	return r.httpClient.Do(req)
}
