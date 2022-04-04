package commands

import (
	"os"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/schollz/progressbar/v3"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

type nodeGitlab struct {
	client    *gitlab.Client
	group     *gitlab.Group
	subGroups *[]*gitlab.Group
	projects  *[]*gitlab.Project
	Rootdir   string
	bar       *progressbar.ProgressBar
	wg        *sync.WaitGroup
	log       *zap.Logger
	auth      *Auth
}

// Update gitlab tree using given credential and root directory
// Do update if the repo/group present or clone/create the directory of repo is not present
func (c *Command) updateGitlab() error {
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	c.log.Debug("Start proccess update ...")
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

	var err error
	client, err := gitlab.NewClient(c.auth.Password, clientFuncOpt)
	if err != nil {
		c.log.Fatal(err.Error())
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
			wg:      wg,
			auth:    c.auth,
			log:     c.log,
		}

		wg.Add(1)
		go node.getSubgroups()
		wg.Add(1)
		go node.updateProject()
	}
	wg.Wait()

	c.log.Debug("Finish execute update-gitlab action")
	return nil
}

// Perform clone action for every repository in gitlab tree
// that has not been cloned inside existing tree folder or given directory
func (c *Command) CloneGitlab() error {
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	c.log.Debug("Start proccess clone ...")
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

	var err error
	client, err := gitlab.NewClient(c.auth.Password, clientFuncOpt)
	if err != nil {
		c.log.Fatal(err.Error())
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
			wg:      wg,
			auth:    c.auth,
			log:     c.log,
		}

		wg.Add(1)
		go node.validateSubgroups()
		wg.Add(1)
		go node.validateProject()
	}
	wg.Wait()

	c.log.Debug("Finish execute clone-gitlab action")
	return nil
}

// List subgroups inside a gitlab group. If there is subgroup than
// perform recursive check again if there is any project or a subgroup.
// Clone the project if it not present otherwise update the project master branch.
func (n *nodeGitlab) getSubgroups() {
	defer n.wg.Done()

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
			wg:      n.wg,
			auth:    n.auth,
			log:     n.log,
		}

		n.wg.Add(1)
		go node.getSubgroups()
		n.wg.Add(1)
		go node.updateProject()
	}
	n.wg.Wait()
}

// List project inside gitlab group and perform update or clone the project
func (n *nodeGitlab) updateProject() {
	defer n.wg.Done()

	projects, _, err := n.client.Groups.ListGroupProjects(n.group.ID, nil)
	if err != nil {
		n.log.Error(err.Error())
		return
	}

	for _, project := range projects {
		err = n.cloneOrUpdateRepo(project)
		if err != nil {
			n.log.Panic(err.Error())
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
	n.log.Sugar().Debugf("List Group: %v", listGroup)

	for _, group := range *n.subGroups {
		path := n.Rootdir + "/" + group.Name
		createDir(path)

		// Recursive check the subgroups
		node := &nodeGitlab{
			client:  n.client,
			group:   group,
			Rootdir: path,
			bar:     n.bar,
			wg:      n.wg,
			auth:    n.auth,
			log:     n.log,
		}

		n.wg.Add(1)
		go node.validateSubgroups()
		n.wg.Add(1)
		go node.validateProject()
	}

	n.wg.Wait()
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
		n.log.Error(err.Error())
		return
	}

	n.log.Sugar().Debugf("Total page subgroup %v : %v", n.group.Name, resp.TotalPages)

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
	n.log.Sugar().Debugf("List Project in group %v: %v", n.group.Name, listProject)

	for _, project := range *n.projects {
		path := n.Rootdir + "/" + project.Name
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			n.wg.Add(1)
			go n.cloneRepo(path, project)
		}
	}
	n.wg.Wait()
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
		n.log.Error(err.Error())
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
func (n *nodeGitlab) cloneOrUpdateRepo(p *gitlab.Project) error {
	path := n.Rootdir + "/" + p.Name
	_, err := os.Stat(path)
	if err == nil || (err != nil && os.IsExist(err)) {
		node := makeNode(&nodeOptions{
			path:      path,
			hardReset: false,
			bar:       n.bar,
		})
		err = node.updateRepo()
		if err != nil {
			n.log.Error(err.Error())
		}
	}

	if err != nil && os.IsNotExist(err) {
		n.cloneRepo(path, p)
	}

	return err
}

// Clone repo from given path and url
// Repo name will using dir name (include case sensitive)
func (n *nodeGitlab) cloneRepo(path string, p *gitlab.Project) {
	defer n.wg.Done()

	var option *git.CloneOptions = &git.CloneOptions{
		URL: p.HTTPURLToRepo,
		Auth: &http.BasicAuth{
			Username: n.auth.Username,
			Password: n.auth.Password,
		},
		Depth:             1,
		ReferenceName:     plumbing.Master,
		SingleBranch:      true,
		Tags:              git.NoTags,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}

	if n.bar != nil {
		n.bar.Describe("Clone: " + p.Name)
		option.Progress = os.Stdout
	}

	n.log.Sugar().Debugf("Clonning %v", p.Name)
	_, err := git.PlainClone(path, false, option)
	if err != nil {
		n.log.Sugar().Errorf("Repo %v: %v", p.Name, err.Error())
		return
	}
	n.log.Sugar().Debugf("Finish Clonning %v", p.Name)
	n.log.Sugar().Debugf("Path Clone: %v/%v", path, p.Name)

	if n.bar != nil {
		_ = n.bar.Add(1)
	}
}
