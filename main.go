package main

import (
	"github.com/glovenkevin/go-git-puller/commands"
	"github.com/glovenkevin/go-git-puller/utils"
)

func main() {
	defer utils.ErrorHandler()
	utils.SetUsageFlag()
	utils.ParseArgs()

	valid := utils.ValidateEnvirontment()
	if valid {
		commands.CommandDispatcher()
	}
}
