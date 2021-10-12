package commands

import (
	"github.com/glovenkevin/go-git-puller/utils"
	"go.uber.org/zap"
)

// Collect the command dispatcher function as a set of key value array.
// Then execute action based on action key provided
func CommandDispatcher(logs *zap.Logger) {
	defer utils.ErrorHandler()

	dispatcher := make(map[string]func(logs *zap.Logger))
	setUpCommandDispatcher(&dispatcher)
	if dispatcher[utils.Action] == nil {
		logs.Panic("Action implementation not found")
	}

	dispatcher[utils.Action](logs)
}

// Set function into key value array so that
// it's not creating much mor if else statement
func setUpCommandDispatcher(dispatcher *map[string]func(logs *zap.Logger)) {

	(*dispatcher)["update"] = UpdatesProjectGit
}
