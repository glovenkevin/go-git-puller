package main

import (
	"flag"

	"github.com/glovenkevin/go-git-puller/commands"
	"github.com/glovenkevin/go-git-puller/utils"
	"go.uber.org/zap"
)

func init() {
	// utils.Environtment = "DEV"
	if utils.Environtment == "PROD" {
		utils.Logs, _ = zap.NewProduction()
	} else {
		utils.Logs, _ = zap.NewDevelopment()
	}
	utils.SetUsageFlag()
	defer utils.Logs.Sync()
}

func main() {
	defer utils.ErrorHandler()

	flag.StringVar(&utils.Action, "c", ".", "Set action")
	flag.StringVar(&utils.Action, "action", ".", "Set action")

	flag.StringVar(&utils.Username, "U", "", "Put username for git")
	flag.StringVar(&utils.Username, "username", "", "Put username for git")

	flag.StringVar(&utils.Password, "S", "", "Put password for git")
	flag.StringVar(&utils.Password, "password", "", "Put password for git")

	flag.StringVar(&utils.RootDir, "path", ".", "Set Working directory root path")
	flag.BoolVar(&utils.Verbose, "verbose", false, "Activate verbose/debug print")
	flag.BoolVar(&utils.HardReset, "hard-reset", false, "Set false to use softreset or otherwise")
	flag.Parse()

	valid := utils.ValidateEnvirontment()
	if valid {
		commands.CommandDispatcher(utils.Logs)
	}
}
