package main

import (
	"os"

	"github.com/glovenkevin/go-git-puller/cli"
	"github.com/glovenkevin/go-git-puller/commands"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	conf := zap.NewProductionConfig()
	conf.Encoding = "console"
	conf.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	conf.DisableStacktrace = true

	log, _ = conf.Build()
}

func main() {

	cli := cli.New()
	err := cli.Parse()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Set logger into debug level if verbose was activated
	if cli.Verbose {
		log = log.WithOptions(zap.WrapCore(
			func(c zapcore.Core) zapcore.Core {
				return zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
					zapcore.AddSync(os.Stderr), zap.DebugLevel)
			},
		))
	}

	cmd, err := constructCommand(cli)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = cmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func constructCommand(c *cli.Cli) (*commands.Command, error) {
	command, err := commands.New(&commands.Options{
		Verbose: c.Verbose,
		Action:  c.Action,
		Dir:     c.Rootdir,
		Baseurl: c.Baseurl,
		Auth: &commands.Auth{
			Username: c.Username,
			Password: c.Password,
		},
		Logs: log,
	})

	return command, err
}
