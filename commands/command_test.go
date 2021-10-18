package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var Log *zap.Logger

func TestMain(m *testing.M) {
	dir := "test"
	Log, _ = zap.NewDevelopment()

	// Preapre dir for directory testing
	_ = os.Mkdir(dir, os.ModeDir)

	code := m.Run()

	os.Remove(dir)
	os.Exit(code)
}

func TestNewCommands(t *testing.T) {

	tests := []struct {
		Name      string
		Input     *Options
		CmdStrct  *Command
		ErrOutput error
	}{
		{
			Name:      "NoParameterGiven",
			Input:     &Options{},
			CmdStrct:  nil,
			ErrOutput: ErrActionNotFound,
		},
		{
			Name: "NoAuthGiven",
			Input: &Options{
				Action: "test",
				Dir:    ".",
				Logs:   Log,
			},
			CmdStrct:  nil,
			ErrOutput: ErrCredentialNotFound,
		},
		{
			Name: "AuthNotFullySet1",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "",
				},
				Dir:  ".",
				Logs: Log,
			},
			CmdStrct:  nil,
			ErrOutput: ErrCredentialNotFound,
		},
		{
			Name: "AuthNotFullySet2",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Password: "",
				},
				Dir:  ".",
				Logs: Log,
			},
			CmdStrct:  nil,
			ErrOutput: ErrCredentialNotFound,
		},
		{
			Name: "AuthNotFullySet3",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "asdf",
					Password: "",
				},
				Dir:  ".",
				Logs: Log,
			},
			CmdStrct:  nil,
			ErrOutput: ErrCredentialNotFound,
		},
		{
			Name: "DirectoryNotSet",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				Logs: Log,
			},
			CmdStrct:  nil,
			ErrOutput: ErrDirNotExist,
		},
		{
			Name: "DirectorySetToFault",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				Logs: Log,
				Dir:  "test2",
			},
			CmdStrct:  nil,
			ErrOutput: ErrDirNotExist,
		},
		{
			Name: "DirectoryIsValid",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				Logs: Log,
				Dir:  "test",
			},
			CmdStrct: &Command{
				action: "test",
				auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				dir: "test",
			},
			ErrOutput: nil,
		},
		{
			Name: "LogNotSet",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				Dir:  ".",
				Logs: nil,
			},
			CmdStrct:  nil,
			ErrOutput: ErrLogsNotDefined,
		},
		{
			Name: "AllFullSetUp",
			Input: &Options{
				Action: "test",
				Auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				Dir:  ".",
				Logs: Log,
			},
			CmdStrct: &Command{
				action: "test",
				auth: &Auth{
					Username: "user",
					Password: "pass",
				},
				dir: ".",
			},
			ErrOutput: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cmd, err := New(test.Input)
			require.Equal(t, test.ErrOutput, err)
			require.Equal(t, test.CmdStrct, cmd)
		})
	}

}

func TestExecuteAction(t *testing.T) {
	tests := []struct {
		Name      string
		Input     *Command
		ErrOutput error
	}{
		{
			Name: "ActionNotRecognize",
			Input: &Command{
				action: "tes",
				dir:    ".",
			},
			ErrOutput: ErrCommandNotFound,
		},
		{
			Name: "ActionNotBeenSet",
			Input: &Command{
				action: "",
				dir:    ".",
			},
			ErrOutput: ErrCommandNotFound,
		},
		{
			Name: "AllSet",
			Input: &Command{
				action: "update",
				dir:    ".",
			},
			ErrOutput: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := test.Input.Execute()
			require.Equal(t, test.ErrOutput, err)
		})
	}
}
