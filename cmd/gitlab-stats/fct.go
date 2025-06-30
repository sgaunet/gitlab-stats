// Package main provides GitLab statistics collection and visualization.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"

	"github.com/sgaunet/gitlab-stats/pkg/gitlab"
)

var (
	// ErrProjectNotFound is returned when a project cannot be found.
	ErrProjectNotFound = errors.New("project not found")
	// ErrGitNotFound is returned when .git directory is not found.
	ErrGitNotFound     = errors.New(".git not found")
)

type project struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	SSHURLToRepo string `json:"ssh_url_to_repo"`
	HTTPURLToRepo string `json:"http_url_to_repo"`
}

func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	log.Infof("Try to find project %s in %s\n", projectName, os.Getenv("GITLAB_URI"))

	gs := gitlab.NewService()
	gs.SetGitlabEndpoint(os.Getenv("GITLAB_URI"))

	resp, err := gs.Get("search?scope=projects&search=" + projectName)
	if err != nil {
		return project{}, fmt.Errorf("failed to search for project: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return project{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var p []project
	err = json.Unmarshal(res, &p)
	if err != nil {
		return project{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	for _, project := range p {
		log.Debugln(project.Name)
		log.Debugln(project.ID)
		log.Debugln(project.HTTPURLToRepo)
		log.Debugln(project.SSHURLToRepo)

		if project.SSHURLToRepo == remoteOrigin {
			return project, nil
		}
	}
	return project{}, ErrProjectNotFound
}

func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for cwd != "/" {
		log.Debugln(cwd)
		stat, err := os.Stat(cwd + string(os.PathSeparator) + ".git")
		if err == nil {
			if stat.IsDir() {
				return cwd, nil
			}
		}
		cwd = filepath.Dir(cwd)
	}
	return "", ErrGitNotFound
}

// GetRemoteOrigin extracts the remote origin URL from git config file.
func GetRemoteOrigin(gitConfigFile string) string {
	cfg, err := ini.Load(gitConfigFile)
	if err != nil {
		log.Errorf("Fail to read file: %v", err)
		os.Exit(1)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	log.Debugln("url:", url)
	return url
}
