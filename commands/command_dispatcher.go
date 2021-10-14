package commands

import (
	"github.com/glovenkevin/go-git-puller/utils"
)

// Collect the command dispatcher function as a set of key value array.
// Then execute action based on action key provided
func CommandDispatcher() {
	defer utils.ErrorHandler()

	dispatcher := make(map[string]func())
	setUpCommandDispatcher(&dispatcher)
	if dispatcher[utils.Action] == nil {
		utils.Panic("Action implementation not found")
	}

	dispatcher[utils.Action]()
}

// Set function into key value array so that
// it's not creating much mor if else statement
func setUpCommandDispatcher(dispatcher *map[string]func()) {

	(*dispatcher)["update"] = UpdatesProjectGit
	(*dispatcher)["update-gitlab"] = UpdateGitlab
}
