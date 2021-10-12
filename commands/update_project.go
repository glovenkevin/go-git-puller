package commands

import (
	"os"
	"strings"
	"time"

	"github.com/glovenkevin/go-git-puller/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.uber.org/zap"
)

func UpdatesProjectGit(logs *zap.Logger) {
	if utils.CheckDirIsGitRepo(utils.RootDir, logs) {
		arrPath := strings.Split(utils.RootDir, "/")
		dirName := arrPath[len(arrPath)-1]
		updateRepository(utils.RootDir, dirName, logs)
	} else {
		updateProject(utils.RootDir, logs)
	}
}

// Search for directory inside given path then check it
// if it was git repo than do update or check other dir inside the directory it self
func updateProject(dir string, logs *zap.Logger) {
	arrDir, _ := os.ReadDir(dir)
	for _, dirEntry := range arrDir {

		if !dirEntry.IsDir() {
			continue
		}

		dirPath := dir + "/" + dirEntry.Name()
		utils.Debug("Dirpath: ", dirPath)

		if utils.CheckDirIsGitRepo(dirPath, logs) {
			updateRepository(dirPath, dirEntry.Name(), logs)
		} else {
			updateProject(dirPath, logs)
		}
	}
}

// Update git repository on master branch
// do git fetch all, restore anything that change and do git pull on master branch
func updateRepository(dirPath string, dirName string, logs *zap.Logger) {
	utils.Debugf("%v is a repo", dirName)
	repo, _ := git.PlainOpen(dirPath)
	repo.Fetch(&git.FetchOptions{})
	workTree, _ := repo.Worktree()

	var err error
	if utils.HardReset {
		err = workTree.Reset(&git.ResetOptions{Mode: git.HardReset})
		utils.CheckIsError(err, logs)
	} else {
		err = workTree.Reset(&git.ResetOptions{Mode: git.SoftReset})
		utils.CheckIsError(err, logs)
	}

	if err != nil {
		return
	}

	workTree.AddWithOptions(&git.AddOptions{All: true})
	err = workTree.Checkout(&git.CheckoutOptions{
		Force:  true,
		Keep:   true,
		Branch: plumbing.Master,
	})
	utils.CheckIsError(err, logs)
	if err != nil && utils.Verbose {
		logs.Sugar().Warnf("Repo %v Got error: %v", dirName, err.Error())
		return
	}

	var auth transport.AuthMethod
	if utils.Auth == "http" {
		auth = transport.AuthMethod(&http.BasicAuth{
			Username: utils.Username,
			Password: utils.Password,
		})
	} else {
		auth = transport.AuthMethod(&http.TokenAuth{
			Token: utils.Token,
		})
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
			utils.Debugf("Repo %v success pulled to the latest master", dirName)
			break
		} else if strings.HasSuffix(err.Error(), "reference has changed concurrently") {
			utils.Debugf("Wait for %dS for error %v", 5, err)
			time.Sleep(5 * time.Second)
		} else {
			logs.Sugar().Fatalf("Repo %v failed to be pulled, cause: %v", dirName, err.Error())
		}
	}
}
