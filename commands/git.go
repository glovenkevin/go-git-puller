package commands

import (
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type node struct {
	name      string
	path      string
	auth      *Auth
	hardReset bool
}

type nodeOptions struct {
	// Define the root path of the action
	path string

	// Set the auth for the git
	auth *Auth

	// Set if the hard reset need to be done
	hardReset bool
}

// Start updating git folder start from given root directory.
// Update was doing used recursive function for every node folder
func (c *Command) updateGit() error {
	// Start the working tree of update
	node := makeNode(&nodeOptions{
		path:      c.dir,
		auth:      c.auth,
		hardReset: c.hardReset,
	})
	err := node.updateProject()
	if err != nil {
		return err
	}

	err = node.updateRepo()
	if err != nil {
		return err
	}

	return nil
}

func makeNode(opt *nodeOptions) *node {
	arrPath := strings.Split(opt.path, "/")
	node := node{
		auth:      opt.auth,
		path:      opt.path,
		name:      arrPath[len(arrPath)-1],
		hardReset: opt.hardReset,
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
		node := makeNode(&nodeOptions{
			path:      dirPath,
			auth:      n.auth,
			hardReset: n.hardReset,
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
		return git.ErrRepositoryNotExists
	}

	var err error
	auth := &http.BasicAuth{
		Username: n.auth.Username,
		Password: n.auth.Password,
	}

	repo, _ := git.PlainOpen(n.path)
	logs.Sugar().Debugf("Found repo %v", n.name)

	_ = repo.Fetch(&git.FetchOptions{Auth: auth})
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

	gitPullOption := git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	}

	// Sometimes happend error reference has changed,
	// wait for few seconds and than do pull again
	for {
		err = workTree.Pull(&gitPullOption)
		if err == nil || (err != nil && strings.Contains(err.Error(), "up-to-date")) {
			logs.Sugar().Debugf("Repo %v is pulled", n.name)
			break
		} else if strings.HasSuffix(err.Error(), "reference has changed concurrently") {
			time.Sleep(5 * time.Second)
			logs.Sugar().Debugf("Ref Changed on repo %v, wait for 5s", n.name)
		} else {
			return err
		}
	}

	return nil
}

// Checking current directory given is a
// git repository
func isRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
