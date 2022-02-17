package commands

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
)

type nodeGitlab struct {
	client    *gitlab.Client
	group     *gitlab.Group
	subGroups *[]*gitlab.Group
	projects  *[]*gitlab.Project
	Rootdir   string
	bar       *progressbar.ProgressBar
}

// Update gitlab tree using given credential and root directory
// Do update if the repo/group present or clone/create the directory of repo is not present
func (c *Command) updateGitlab() error {

	logs.Debug("Start proccess update ...")
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

		node := &nodeGitlab{
			client:  client,
			group:   group,
			bar:     c.bar,
			Rootdir: path,
		}

		node.getSubgroups()
		node.updateProject()
	}

	logs.Debug("Finish execute update-gitlab action")
	return nil
}

// Perform clone action for every repository in gitlab tree
// that has not been cloned inside existing tree folder or given directory
func (c *Command) CloneGitlab() error {

	logs.Debug("Start proccess clone ...")
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
		if group.Name != "WingsDev" {
			continue
		}

		path := c.dir + "/" + group.Name
		createDir(path)

		node := &nodeGitlab{
			client:  client,
			group:   group,
			bar:     c.bar,
			Rootdir: path,
		}

		node.validateSubgroups()
		node.validateProject()
	}

	logs.Debug("Finish execute clone-gitlab action")
	return nil
}

// List subgroups inside a gitlab group. If there is subgroup than
// perform recursive check again if there is any project or a subgroup.
// Clone the project if it not present otherwise update the project master branch.
func (n *nodeGitlab) getSubgroups() {
	subGroups, _, _ := n.client.Groups.ListSubgroups(n.group.ID, nil)
	for _, group := range subGroups {

		path := n.Rootdir + "/" + group.Name
		createDir(path)

		// Recursive check the subgroups
		node := &nodeGitlab{
			client:  n.client,
			group:   group,
			Rootdir: path,
			bar:     n.bar,
		}
		node.getSubgroups()
		node.updateProject()
	}
}

// List project inside gitlab group and perform update or clone the project
func (n *nodeGitlab) updateProject() {
	projects, _, err := n.client.Groups.ListGroupProjects(n.group.ID, nil)
	if err != nil {
		logs.Warn(err.Error())
		return
	}

	for _, project := range projects {
		err = cloneOrUpdateRepo(project, n.Rootdir, n.bar)
		if err != nil {
			logs.Panic(err.Error())
		}
	}
}

// List subgroups inside a gitlab group. If there is still any subgroup than
// perform recursive check again if there is a subgroup again.
// It's also check if there is a project inside current group,
// if exist but not present in current directory then it will clone the project otherwise do nothing.
func (n *nodeGitlab) validateSubgroups() {
	n.getAllSubgroups()

	listGroup := ""
	for _, group := range *n.subGroups {
		listGroup += group.Name + ","
	}
	logs.Sugar().Debugf("List Group: %v", listGroup)

	for _, group := range *n.subGroups {
		path := n.Rootdir + "/" + group.Name
		createDir(path)

		// Recursive check the subgroups
		node := &nodeGitlab{
			client:  n.client,
			group:   group,
			Rootdir: path,
			bar:     n.bar,
		}
		node.validateSubgroups()
		node.validateProject()
	}
}

// Fetch all subgroups in a group. By default gitlab only returns 20 results at a time.
// We need to loop over the page to get all the projects and return it.
func (n *nodeGitlab) getAllSubgroups() {
	var (
		subGroups     []*gitlab.Group
		nextSubGroups []*gitlab.Group
		resp          *gitlab.Response
		err           error
	)

	subGroups, resp, err = n.client.Groups.ListSubgroups(n.group.ID, nil)
	if err != nil {
		logs.Warn(err.Error())
		return
	}

	logs.Sugar().Debugf("Total page subgroup %v : %v", n.group.Name, resp.TotalPages)

	if resp.TotalPages > 1 {
		for resp.NextPage != 0 {
			nextSubGroups, resp, _ = n.client.Groups.ListSubgroups(n.group.ID, &gitlab.ListSubgroupsOptions{
				ListOptions: gitlab.ListOptions{
					Page: resp.NextPage,
				},
			})

			subGroups = append(subGroups, nextSubGroups...)
		}
	}

	n.subGroups = &subGroups
}

// List project inside gitlab group. Clone project when inside the directory
// not present or do nothing
func (n *nodeGitlab) validateProject() {
	n.getAllProjects()

	listProject := ""
	for _, project := range *n.projects {
		listProject += project.Name + ","
	}
	logs.Sugar().Debugf("List Project in group %v: %v", n.group.Name, listProject)

	for _, project := range *n.projects {
		path := n.Rootdir + "/" + project.Name
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			cloneRepo(path, project, n.bar)
		}
	}

}

func (n *nodeGitlab) getAllProjects() {
	var (
		projects    []*gitlab.Project
		nextProject []*gitlab.Project
		resp        *gitlab.Response
		err         error
	)

	projects, resp, err = n.client.Groups.ListGroupProjects(n.group.ID, nil)
	if err != nil {
		logs.Warn(err.Error())
		return
	}

	if resp.TotalPages > 1 {
		for resp.NextPage != 0 {
			nextProject, resp, _ = n.client.Groups.ListGroupProjects(n.group.ID, &gitlab.ListGroupProjectsOptions{
				ListOptions: gitlab.ListOptions{
					Page: resp.NextPage,
				},
			})

			projects = append(projects, nextProject...)
		}
	}

	n.projects = &projects
}

// Create the folder of given directory if not exist
func createDir(path string) {
	err := os.Mkdir(path, os.ModeDir)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return
	}
}

// Check whether the repository is Exist
// Do update repo if exist otherwise clone the repo
func cloneOrUpdateRepo(p *gitlab.Project, rootDir string, bar *progressbar.ProgressBar) error {
	path := rootDir + "/" + p.Name
	_, err := os.Stat(path)
	if err == nil || (err != nil && os.IsExist(err)) {
		node := makeNode(&nodeOptions{
			path:      path,
			hardReset: false,
			bar:       bar,
		})
		err = node.updateRepo()
		if err != nil {
			logs.Warn(err.Error())
		}
	}

	if err != nil && os.IsNotExist(err) {
		cloneRepo(path, p, bar)
	}

	return err
}

// Clone repo from given path and url
// Repo name will using dir name (include case sensitive)
func cloneRepo(path string, p *gitlab.Project, bar *progressbar.ProgressBar) {
	var option *git.CloneOptions = &git.CloneOptions{
		URL: p.HTTPURLToRepo,
		Auth: &http.BasicAuth{
			Username: auth.Username,
			Password: auth.Password,
		},
		ReferenceName: plumbing.Master,
		SingleBranch:  true,
		Tags:          git.NoTags,
	}

	if bar != nil {
		bar.Describe("Clone: " + p.Name)
		option.Progress = os.Stdout
	}

	logs.Sugar().Debugf("Clonning %v", p.Name)
	_, err := git.PlainClone(path, false, option)
	if err != nil {
		logs.Sugar().Warnf("Repo %v: %v", p.Name, err.Error())
		return
	}
	logs.Sugar().Debugf("Finish Clonning %v", p.Name)
	logs.Sugar().Debugf("Path Clone: %v/%v", path, p.Name)

	if bar != nil {
		_ = bar.Add(1)
	}
}
