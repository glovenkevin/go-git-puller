package commands

import (
	"os"
	"strings"

	"github.com/glovenkevin/go-git-puller/utils"
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
		go checkPath(path)
		getSubgroups(group, path)
		getProjects(group, path)
	}
}

func getSubgroups(g *gitlab.Group, parent string) {
	subGroups, _, _ := client.Groups.ListSubgroups(g.ID, nil)
	for _, group := range subGroups {
		path := parent + "/" + group.Name
		go checkPath(path)
		getSubgroups(group, path)
		getProjects(group, path)
	}
}

func checkPath(path string) {
	err := os.Mkdir(path, os.ModeDir)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return
	}
	utils.CheckIsError(err)
}

func getProjects(g *gitlab.Group, parent string) {
	// for _, project := range g.Projects {

	// }
}
