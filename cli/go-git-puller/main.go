package main

import (
	"fmt"

	"github.com/glovenkevin/go-git-puller/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	cli := cli.New()
	err := cli.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set Logger config
	conf := zap.NewProductionConfig()
	conf.Encoding = "console"
	conf.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	conf.DisableStacktrace = true
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if cli.Verbose {
		conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	var zlog, _ = conf.Build()

	cmd, err := cli.NewCommand(zlog)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cmd.Execute()
	if err != nil {
		fmt.Println(err)
		return
	}
}
