package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"

	"github.com/sgaunet/gitlab-stats/gitlabRequest"
)

type project struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	SshUrlToRepo  string `json:"ssh_url_to_repo"`
	HttpUrlToRepo string `json:"http_url_to_repo"`
}

func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	log.Infof("Try to find project %s in %s\n", projectName, os.Getenv("GITLAB_URI"))

	_, res, err := gitlabRequest.Request("search?scope=projects&search=" + projectName)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	var p []project
	err = json.Unmarshal(res, &p)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	for _, project := range p {
		log.Debugln(project.Name)
		log.Debugln(project.Id)
		log.Debugln(project.HttpUrlToRepo)
		log.Debugln(project.SshUrlToRepo)

		if project.SshUrlToRepo == remoteOrigin {
			return project, err
		}
	}
	return project{}, errors.New("project not found")
}

func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for cwd != "/" {
		log.Debugln(cwd)
		stat, err := os.Stat(cwd + string(os.PathSeparator) + ".git")
		if err == nil {
			if stat.IsDir() {
				return cwd, err
			}
		}
		cwd = filepath.Dir(cwd)
	}
	return "", errors.New(".git not found")
}

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
