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
	client     *gitlab.Client
	group      *gitlab.Group
	subGroups  []*gitlab.Group
	projects   []*gitlab.Project
	Rootdir    string
	bar        *progressbar.ProgressBar
	wg         *sync.WaitGroup
	log        *zap.Logger
	auth       *Auth
	exGroups   map[string]struct{}
	exProjects map[string]struct{}
}

// Update gitlab tree using given credential and root directory
// Do update if the repo/group present or clone/create the directory of repo is not present
func (c *Command) UpdateGitlab() error {
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	c.log.Debug("Start proccess update ...")
	defer func() {
		if c.bar != nil {
			_ = c.bar.Finish()
		}
		c.log.Debug("Finish execute clone-gitlab action")
	}()

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
		node.updateProject()
	}
	wg.Wait()

	if c.bar != nil {
		_ = c.bar.Add(1)
	}
	c.log.Debug("Finish execute update-gitlab action")
	return nil
}

// Perform clone action for every repository in gitlab tree
// that has not been cloned inside existing tree folder or given directory
func (c *Command) CloneGitlab() error {
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	c.log.Debug("Start proccess clone ...")
	defer func() {
		if c.bar != nil {
			_ = c.bar.Finish()
		}
		c.log.Debug("Finish execute clone-gitlab action")
	}()

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
			client:     client,
			group:      group,
			bar:        c.bar,
			Rootdir:    path,
			wg:         wg,
			auth:       c.auth,
			log:        c.log,
			exGroups:   c.exGroups,
			exProjects: c.exProjects,
		}

		wg.Add(1)
		go node.validateSubgroups()
		node.validateProject()
	}
	wg.Wait()

	if c.bar != nil {
		_ = c.bar.Add(1)
	}
	return nil
}

// List subgroups inside a gitlab group. If there is subgroup than
// perform recursive check again if there is any project or a subgroup.
// Clone the project if it not present otherwise update the project master branch.
func (n *nodeGitlab) getSubgroups() {
	defer n.wg.Done()
	n.getAllSubgroups()
	n.filterGroups()

	for _, group := range n.subGroups {

		path := n.Rootdir + "/" + group.Name
		createDir(path)

		// Recursive check the subgroups
		node := n
		node.group = group
		node.Rootdir = path

		n.wg.Add(1)
		go node.getSubgroups()
		node.updateProject()
	}
	n.wg.Wait()
}

// List project inside gitlab group and perform update or clone the project
func (n *nodeGitlab) updateProject() {
	n.getAllProjects()
	n.filterProjects()

	for _, project := range n.projects {
		err := n.cloneOrUpdateRepo(project)
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
	defer n.wg.Done()
	n.getAllSubgroups()
	n.filterGroups()

	if n.bar == nil {
		listGroup := ""
		for _, group := range n.subGroups {
			listGroup += group.Name + ","
		}
		n.log.Sugar().Debugf("List Group: %v", listGroup)
	}

	for _, group := range n.subGroups {
		path := n.Rootdir + "/" + group.Name
		createDir(path)

		// Recursive check the subgroups
		node := n
		node.group = group
		node.Rootdir = path

		n.wg.Add(1)
		go node.validateSubgroups()
		node.validateProject()
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

	subGroups, resp, err = n.client.Groups.ListSubGroups(n.group.ID, nil)
	if err != nil {
		n.log.Error(err.Error())
		return
	}

	n.log.Sugar().Debugf("Total page subgroup %v : %v", n.group.Name, resp.TotalPages)

	if resp.TotalPages > 1 {
		for resp.NextPage != 0 {
			nextSubGroups, resp, _ = n.client.Groups.ListSubGroups(n.group.ID, &gitlab.ListSubGroupsOptions{
				ListOptions: gitlab.ListOptions{
					Page: resp.NextPage,
				},
			})

			subGroups = append(subGroups, nextSubGroups...)
		}
	}

	n.subGroups = subGroups
}

func (n *nodeGitlab) filterGroups() {
	size := len(n.subGroups)
	if size == 0 {
		return
	}

	for i := 0; i < size; i++ {
		if _, ok := n.exGroups[n.subGroups[i].Name]; ok {
			newGroups := make([]*gitlab.Group, 0)
			newGroups = append(newGroups, n.subGroups[:i]...)
			n.subGroups = append(newGroups, n.subGroups[i+1:]...)
			size--
			i--
		}
	}
}

// List project inside gitlab group. Clone project when inside the directory
// not present or do nothing
func (n *nodeGitlab) validateProject() {
	n.getAllProjects()
	n.filterProjects()

	if n.bar == nil {
		listProject := ""
		for _, project := range n.projects {
			listProject += project.Name + ","
		}
		n.log.Sugar().Debugf("List Project in group %v: %v", n.group.Name, listProject)
	}

	if n.bar != nil {
		n.bar.ChangeMax64(int64(n.bar.GetMax() + len(n.projects)))
	}

	for _, project := range n.projects {
		if n.bar != nil {
			_ = n.bar.Add(1)
		}

		path := n.Rootdir + "/" + project.Name
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			n.cloneRepo(path, project)
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

	n.projects = projects
}

func (n *nodeGitlab) filterProjects() {
	size := len(n.projects)
	if size == 0 {
		return
	}

	for i := 0; i < size; i++ {
		if _, ok := n.exProjects[n.projects[i].Name]; ok {
			newProjects := make([]*gitlab.Project, 0)
			newProjects = append(newProjects, n.projects[:i]...)
			n.projects = append(newProjects, n.projects[i+1:]...)
			size--
			i--
		}
	}
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
	var option *git.CloneOptions = &git.CloneOptions{
		URL: p.HTTPURLToRepo,
		Auth: &http.BasicAuth{
			Username: n.auth.Username,
			Password: n.auth.Password,
		},
		Depth:         1,
		ReferenceName: plumbing.HEAD,
		SingleBranch:  true,
		Tags:          git.NoTags,
	}

	if n.bar != nil {
		n.bar.Describe("Clone: " + p.Name)
	}

	n.log.Sugar().Debugf("Clonning %v", p.Name)
	_, err := git.PlainClone(path, false, option)
	if err != nil {
		n.log.Sugar().Errorf("Repo %v: %v, Path: %v", p.Name, err.Error(), path)
		return
	}
	n.log.Sugar().Debugf("Finish Clonning %v", p.Name)
	n.log.Sugar().Debugf("Path Clone: %v/%v", path, p.Name)
}
