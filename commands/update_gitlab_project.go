package commands

import (
	"os"
	"strings"

	"github.com/glovenkevin/go-git-puller/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/xanzy/go-gitlab"
)

var client *gitlab.Client

func UpdateGitlab() {

	var clientFuncOpt gitlab.ClientOptionFunc = nil
	if utils.Baseurl != "" {
		clientFuncOpt = gitlab.WithBaseURL(utils.Baseurl)
	}

	var err error
	client, err = gitlab.NewClient(utils.Token, clientFuncOpt)
	if err != nil {
		utils.Errorf("Failed to create client: %v", err)
		return
	}

	updateGitlab(client)
}

func updateGitlab(c *gitlab.Client) {

	rootGroups, resp, err := c.Groups.ListGroups(&gitlab.ListGroupsOptions{AllAvailable: gitlab.Bool(true), TopLevelOnly: gitlab.Bool(true)})
	if err != nil || resp.StatusCode != 200 {
		utils.Errorf("Failed to get root group: %v", err)
		return
	}

	for _, group := range rootGroups {
		path := utils.RootDir + "/" + group.Name
		go createDir(path)
		getSubgroups(group, path)
		getProjects(group, path)
	}
}

func getSubgroups(g *gitlab.Group, parent string) {
	subGroups, _, _ := client.Groups.ListSubgroups(g.ID, nil)
	for _, group := range subGroups {
		path := parent + "/" + group.Name
		go createDir(path)
		getSubgroups(group, path)
		getProjects(group, path)
	}
}

func createDir(path string) {
	err := os.Mkdir(path, os.ModeDir)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return
	}
	utils.CheckIsError(err)
}

func getProjects(g *gitlab.Group, parent string) {

	projects, _, err := client.Groups.ListGroupProjects(g.ID, nil)
	if err != nil {
		utils.Error(err.Error())
		return
	}

	for _, project := range projects {
		cloneOrUpdateProject(project, parent)
	}

}

func cloneOrUpdateProject(p *gitlab.Project, rootDir string) {
	path := rootDir + "/" + p.Name
	if utils.ValidateFolder(path) {
		utils.Debugf("Update existing repo %v", p.Name)
		UpdateRepository(path, p.Name)
	} else {
		utils.Debugf("Clone new repo %v", p.Name)
		CloneRepository(path, p.WebURL)
	}
}

func CloneRepository(path string, url string) {
	var option *git.CloneOptions = &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: "token",
			Password: utils.Token,
		},
	}

	_, err := git.PlainClone(path, false, option)
	if err != nil {
		utils.Warnf("Error repo dir %v", path)
		utils.Warn(err.Error())
	}
}
