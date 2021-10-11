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
	updateProject(utils.RootDir, logs)
}

func updateProject(dir string, logs *zap.Logger) {

	arrDir, _ := os.ReadDir(dir)
	for _, dirEntry := range arrDir {
		dirPath := dir + "/" + dirEntry.Name()
		utils.Debug("Dirpath: ", dirPath)

		if dirEntry.IsDir() && utils.CheckDirIsGitRepo(dirPath, logs) {
			utils.Debugf("%v is a repo", dirEntry.Name())
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

			if err == nil {
				utils.Debugf("%v : Change to branch master", dirEntry.Name())

				workTree.AddWithOptions(&git.AddOptions{All: true})
				err = workTree.Checkout(&git.CheckoutOptions{
					Force:  true,
					Keep:   true,
					Branch: plumbing.Master,
				})
				utils.CheckIsError(err, logs)
				if err != nil && utils.Verbose {
					logs.Sugar().Warnf("Repo %v Got error: %v", dirEntry.Name(), err.Error())
				}

				if err == nil {
					gitPullOption := git.PullOptions{
						RemoteName: "origin",
						// Progress:   os.Stderr,
						Auth: transport.AuthMethod(&http.BasicAuth{
							Username: utils.Username,
							Password: utils.Password,
						}),
					}

					// Sometimes happend error reference has changed, so wait for few seconds and than do pull again
					for {
						err = workTree.Pull(&gitPullOption)
						utils.CheckIsError(err, logs)
						if err == nil || (err != nil && strings.Contains(err.Error(), "up-to-date")) {
							utils.Debugf("Repo %v success pulled to the latest master", dirEntry.Name())
							break
						} else if strings.HasSuffix(err.Error(), "reference has changed concurrently") {
							utils.Debugf("Wait for %dS for error %v", 5, err)
							time.Sleep(5 * time.Second)
						} else {
							logs.Sugar().Fatalf("Repo %v failed to be pulled, cause: %v", dirEntry.Name(), err.Error())
						}
					}
				}

			} else {
				logs.Sugar().Warnf("Please check repo %v for changes", dirEntry.Name())
			}

		} else {
			updateProject(dirPath, logs)
		}
	}

}
