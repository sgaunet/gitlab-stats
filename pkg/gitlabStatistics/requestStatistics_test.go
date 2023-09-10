package gitlabstatistics_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
	gitlabstatistics "github.com/sgaunet/gitlab-stats/pkg/gitlabStatistics"
)

func TestGetStatisticsGoodResponse(t *testing.T) {
	response := gitlabstatistics.Statistics{
		Statistics: gitlabstatistics.Statistic{
			Counts: gitlabstatistics.Counts{
				All:    1,
				Closed: 1,
				Opened: 1,
			},
		},
	}
	responseJSON, _ := json.Marshal(response)
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, string(responseJSON))
		}))
	defer ts.Close()

	// get request
	client := ts.Client()

	s := gitlab.NewService()
	s.SetHttpClient(client)
	s.SetGitlabEndpoint(ts.URL)

	r := gitlabstatistics.NewServiceStatistics()
	r.SetProjectId(1)
	res, err := r.GetStatistics(s)
	if err != nil {
		t.Errorf("GetStatistics() error = %v", err)
	}

	if cmp.Equal(res, response) == false {
		t.Errorf("GetStatistics() = %v, want %v", res, response)
	}
}

func TestGetStatisticsWrongResponse(t *testing.T) {
	type response struct {
		Field int `json:"field"`
	}
	resp := response{
		Field: 1,
	}
	responseJSON, _ := json.Marshal(resp)
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, string(responseJSON))
		}))
	defer ts.Close()

	// get request
	client := ts.Client()

	s := gitlab.NewService()
	s.SetHttpClient(client)
	s.SetGitlabEndpoint(ts.URL)

	r := gitlabstatistics.NewServiceStatistics()
	r.SetProjectId(1)
	_, err := r.GetStatistics(s)
	if err == nil {
		t.Errorf("GetStatistics() should return an error")
	}
}
