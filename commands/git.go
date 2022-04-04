package commands

import (
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

type node struct {
	name      string
	path      string
	hardReset bool
	bar       *progressbar.ProgressBar
	log       *zap.Logger
	auth      *Auth
}

type nodeOptions struct {
	// Define the root path of the action
	path string

	// Set if the hard reset need to be done
	hardReset bool

	// Set the main progress bar
	bar *progressbar.ProgressBar

	// Set the default logger for the node
	log *zap.Logger

	// Set the default auth method (username/password or with token)
	auth *Auth
}

// Start updating git folder from the given root directory.
// The update was doing recursive function for every node folder inside given directory
func (c *Command) updateGit() error {

	if c.bar != nil {
		_ = c.bar.RenderBlank()
		defer func() {
			_ = c.bar.Finish()
		}()
	}

	// Start the working tree of update
	node := makeNode(&nodeOptions{
		path:      c.dir,
		hardReset: c.hardReset,
		bar:       c.bar,
		log:       c.log,
		auth:      c.auth,
	})

	err := node.updateProject()
	if err != nil {
		return err
	}

	err = node.updateRepo()
	if err != nil {
		return err
	}

	c.log.Debug("Finish updating project")
	return nil
}

// Create node for every folder being accessed.
// This will help to do better saving data about path and dir name
func makeNode(opt *nodeOptions) *node {
	arrPath := strings.Split(opt.path, "/")
	node := node{
		path:      opt.path,
		name:      arrPath[len(arrPath)-1],
		hardReset: opt.hardReset,
		bar:       opt.bar,
		log:       opt.log,
		auth:      opt.auth,
	}
	return &node
}

// Search for directory inside given path then check it
// if it was git repo than do update or check other dir inside the directory it self
func (n *node) updateProject() error {
	if isRepo(n.path) {
		return nil
	}

	arrDir, err := os.ReadDir(n.path)
	if err != nil {
		return err
	}

	for _, dirEntry := range arrDir {

		if !dirEntry.IsDir() {
			continue
		}

		dirPath := n.path + "/" + dirEntry.Name()
		n.log.Sugar().Debug(dirPath)

		node := makeNode(&nodeOptions{
			path:      dirPath,
			hardReset: n.hardReset,
			bar:       n.bar,
		})

		err = node.updateProject()
		if err != nil {
			return err
		}

		err = node.updateRepo()
		if err != nil {
			return err
		}
	}

	return nil
}

// Update git repository on master branch
// do git fetch all, restore anything that change and do git pull on master branch
func (n *node) updateRepo() error {
	if !isRepo(n.path) {
		return nil
	}

	var err error
	auth := &http.BasicAuth{
		Username: n.auth.Username,
		Password: n.auth.Password,
	}

	repo, _ := git.PlainOpen(n.path)
	n.log.Sugar().Debugf("Updating %v", n.name)

	gitPullOption := git.PullOptions{
		RemoteName:   git.DefaultRemoteName,
		Auth:         auth,
		SingleBranch: true,
	}

	if n.bar != nil {
		n.bar.Describe("Updating " + n.name)
		gitPullOption.Progress = os.Stdout
	}

	workTree, _ := repo.Worktree()
	if n.hardReset {
		_ = workTree.Reset(&git.ResetOptions{Mode: git.HardReset})
	} else {
		_ = workTree.Reset(&git.ResetOptions{Mode: git.SoftReset})
	}

	if err != nil {
		return err
	}

	_ = workTree.AddWithOptions(&git.AddOptions{All: true})
	err = workTree.Checkout(&git.CheckoutOptions{
		Force:  true,
		Keep:   true,
		Branch: plumbing.Master,
	})
	if err != nil {
		return err
	}

	// Sometimes happend error reference has changed,
	// wait for few seconds and than do pull again
	for {
		err = workTree.Pull(&gitPullOption)
		if err == nil || (err != nil && strings.Contains(err.Error(), "up-to-date")) {
			n.log.Sugar().Debugf("%v is pulled", n.name)
			break
		} else if strings.HasSuffix(err.Error(), "reference has changed concurrently") {
			time.Sleep(5 * time.Second)
			n.log.Sugar().Debugf("Error Ref Changed on repo %v, wait for 5s", n.name)
		} else {
			return err
		}
	}

	n.log.Sugar().Debugf("Finish updating repo %v", n.name)
	if n.bar != nil {
		_ = n.bar.Add(1)
	}
	return nil
}

// Checking current given directory is a
// git repository
func isRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
