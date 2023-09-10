package gitlab_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
)

func TestGitlabService_Groups(t *testing.T) {
	response := []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}{
		{
			Id:   1,
			Name: "test",
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

	r := gitlab.NewService()
	r.SetHttpClient(client)
	r.SetGitlabEndpoint(ts.URL)
	// retrieve groups
	resp, err := r.Get("groups")
	if err != nil {
		t.Error(err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGitlabService_CheckToken(t *testing.T) {
	tokenGitlab := "test"
	response := []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}{
		{
			Id:   1,
			Name: "",
		},
	}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerPrivateToken := r.Header.Get("PRIVATE-TOKEN")
			response[0].Name = headerPrivateToken
			responseJSON, _ := json.Marshal(response)
			fmt.Fprintln(w, string(responseJSON))
		}))
	defer ts.Close()

	// get request
	client := ts.Client()

	r := gitlab.NewService()
	r.SetHttpClient(client)
	r.SetGitlabEndpoint(ts.URL)
	r.SetToken(tokenGitlab)

	// retrieve groups
	resp, err := r.Get("groups")
	if err != nil {
		t.Error(err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Error(err.Error())
	}

	if response[0].Name != tokenGitlab {
		t.Error("tokn not found in response")
	}
}

func TestGitlabService_Post_CheckToken(t *testing.T) {
	tokenGitlab := "test"
	response := []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}{
		{
			Id:   1,
			Name: "",
		},
	}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerPrivateToken := r.Header.Get("PRIVATE-TOKEN")
			response[0].Name = headerPrivateToken
			responseJSON, _ := json.Marshal(response)
			fmt.Fprintln(w, string(responseJSON))
		}))
	defer ts.Close()

	// get request
	client := ts.Client()

	r := gitlab.NewService()
	r.SetHttpClient(client)
	r.SetGitlabEndpoint(ts.URL)
	r.SetToken(tokenGitlab)

	// retrieve groups
	resp, err := r.Post("groups")
	if err != nil {
		t.Error(err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Error(err.Error())
	}

	if response[0].Name != tokenGitlab {
		t.Error("tokn not found in response")
	}
}
