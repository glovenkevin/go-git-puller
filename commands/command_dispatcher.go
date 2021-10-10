package commands

import (
	"github.com/glovenkevin/go-git-puller/utils"
	"go.uber.org/zap"
)

func CommandDispatcher(logs *zap.Logger) {
	defer utils.ErrorHandler()

	dispatcher := make(map[string]func(logs *zap.Logger))
	setUpCommandDispatcher(&dispatcher)
	if dispatcher[utils.Action] == nil {
		logs.Panic("Action implementation not found")
	}

	dispatcher[utils.Action](logs)
}

func setUpCommandDispatcher(dispatcher *map[string]func(logs *zap.Logger)) {

	(*dispatcher)["update"] = UpdatesProjectGit
}
