package commands

import (
	"errors"
	"os"

	"go.uber.org/zap"
)

type Options struct {
	// Flag for activating logging while the program running
	Verbose bool

	// Define action to be executed in command
	Action string

	// Define root directory of the action
	Dir string

	// a flag that used to make decision wether it's need to be hard reset
	// before the pull being executed
	Hardreset bool

	// Define authentication used for interacting with git repository
	Auth *Auth

	// Define base url (usefull for update-gitlab or clone project)
	Baseurl string

	// Set the zap logger
	Logs *zap.Logger
}

type Auth struct {
	// Username being used for authentication with git.
	// If this token was set, then this field will have default value "token" (acording to go-git docs to use like this)
	Username string

	// Password for authentication with git
	// If token was set, then it's gonna be put in here
	Password string
}

var (
	ErrCommandNotFound    = errors.New("Command not found/unrecognize")
	ErrCredentialNotFound = errors.New("Credential has not been set completely")
	ErrActionNotFound     = errors.New("Action not been initialize")
	ErrLogsNotDefined     = errors.New("Zap logger has not been defined")
	ErrDirNotExist        = errors.New("Directory not valid/exist")

	logs *zap.Logger
)

func New(opt *Options) (*Command, error) {

	if err := validate(opt); err != nil {
		return nil, err
	}

	c := Command{
		verbose:   opt.Verbose,
		action:    opt.Action,
		auth:      opt.Auth,
		dir:       opt.Dir,
		hardReset: opt.Hardreset,
		baseurl:   opt.Baseurl,
	}

	// Set the logger for the command package
	logs = opt.Logs

	return &c, nil
}

// Validate given options is enough to do the task
// Credential, action performed, directory and the logs
func validate(opt *Options) error {

	if opt.Action == "" {
		return ErrActionNotFound
	}

	if _, err := os.Stat(opt.Dir); err != nil {
		return ErrDirNotExist
	}

	if opt.Auth == nil ||
		(opt.Auth != nil && (opt.Auth.Username == "" || opt.Auth.Password == "")) {
		return ErrCredentialNotFound
	}

	if opt.Logs == nil {
		return ErrLogsNotDefined
	}

	return nil
}

type Command struct {
	verbose   bool
	action    string
	dir       string
	auth      *Auth
	hardReset bool
	baseurl   string
}

// Execute action based on action key provided
func (c *Command) Execute() error {

	dispatcher := map[string]func() error{
		"update": func() error {
			return c.updateGit()
		},
		"update-gitlab": func() error {
			return c.updateGitlab()
		},
	}

	if dispatcher[c.action] == nil {
		return ErrCommandNotFound
	}

	err := dispatcher[c.action]()
	if err != nil {
		return err
	}

	return nil
}
