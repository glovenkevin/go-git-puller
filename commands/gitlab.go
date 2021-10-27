package commands

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
)

// Update gitlab tree using given credential and root directory
// Do update if the repo/group present or clone/create the directory of repo is not present
func (c *Command) updateGitlab() error {

	if c.bar != nil {
		_ = c.bar.RenderBlank()
		defer func() {
			_ = c.bar.Finish()
		}()
	}

	var clientFuncOpt gitlab.ClientOptionFunc = nil
	if c.baseurl != "" {
		clientFuncOpt = gitlab.WithBaseURL(c.baseurl)
	}
	auth = c.auth

	var err error
	client, err := gitlab.NewClient(c.auth.Password, clientFuncOpt)
	if err != nil {
		logs.Fatal(err.Error())
		return err
	}

	rootGroups, resp, err := client.Groups.ListGroups(
		&gitlab.ListGroupsOptions{
			AllAvailable: gitlab.Bool(true),
			TopLevelOnly: gitlab.Bool(true),
		},
	)

	if err != nil || resp.StatusCode != 200 {
		return err
	}

	for _, group := range rootGroups {
		path := c.dir + "/" + group.Name
		createDir(path)
		getSubgroups(client, group, path, c.bar)
		getProjects(client, group, path, c.bar)
	}

	return nil
}

// List subgroups inside a gitlab group. If there is subgroup than do
// Recursive check again if there is project or a subgroup again
func getSubgroups(c *gitlab.Client, g *gitlab.Group, parent string, bar *progressbar.ProgressBar) {
	subGroups, _, _ := c.Groups.ListSubgroups(g.ID, nil)
	for _, group := range subGroups {
		path := parent + "/" + group.Name
		createDir(path)
		getSubgroups(c, group, path, bar)
		getProjects(c, group, path, bar)
	}
}

// Create the folder of given directory if not exist
func createDir(path string) {
	err := os.Mkdir(path, os.ModeDir)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return
	}
}

// List project inside gitlab group and do update or clone the project
func getProjects(c *gitlab.Client, g *gitlab.Group, parent string, bar *progressbar.ProgressBar) {

	projects, _, err := c.Groups.ListGroupProjects(g.ID, nil)
	if err != nil {
		logs.Warn(err.Error())
		return
	}

	for _, project := range projects {
		cloneOrUpdateRepo(project, parent, bar)
	}

}

// Check wether the repository is Exist or not
// Do update repo if exist or clone repo if it not present in directory
func cloneOrUpdateRepo(p *gitlab.Project, rootDir string, bar *progressbar.ProgressBar) {
	path := rootDir + "/" + p.Name
	_, err := os.Stat(path)
	if err != nil {
		node := makeNode(&nodeOptions{
			path:      path,
			hardReset: false,
			bar:       bar,
		})
		err = node.updateRepo()
		if err != nil {
			logs.Warn(err.Error())
		}
	} else {
		cloneRepo(path, p, bar)
	}
}

// Clone repo from given path and url
// Repo name will using dir name (include case sensitive)
func cloneRepo(path string, p *gitlab.Project, bar *progressbar.ProgressBar) {
	var option *git.CloneOptions = &git.CloneOptions{
		URL: p.WebURL,
		Auth: &http.BasicAuth{
			Username: auth.Username,
			Password: auth.Password,
		},
	}

	if bar != nil {
		bar.Describe("Clone: " + p.Name)
	}

	_, err := git.PlainClone(path, false, option)
	if err != nil {
		logs.Warn(err.Error())
		return
	}

	if bar != nil {
		_ = bar.Add(1)
	}
}
