package main

import (
	"github.com/glovenkevin/go-git-puller/cli"
	"github.com/glovenkevin/go-git-puller/commands"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	conf := zap.NewProductionConfig()
	conf.Encoding = "console"
	conf.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	conf.DisableStacktrace = true

	log, _ = conf.Build()
}

func main() {

	cli := cli.New()
	err := cli.Parse()
	if err != nil {
		log.Fatal(err.Error())
	}

	command, err := commands.New(&commands.Options{
		Verbose: cli.Verbose,
		Action:  cli.Action,
		Dir:     cli.Rootdir,
		Baseurl: cli.Baseurl,
		Auth: &commands.Auth{
			Username: cli.Username,
			Password: cli.Password,
		},
		Logs: log,
	})

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	err = command.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
