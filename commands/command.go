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
	Username string
	Password string
}

var (
	ErrCommandNotFound    = errors.New("Command not found/unrecognize")
	ErrCredentialNotFound = errors.New("Credential has not been set completely")
	ErrActionNotFound     = errors.New("Action not been initialize")
	ErrLogsNotDefine      = errors.New("Zap logger has not been defined")

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

	// Set dispatcher inside command struct
	c.dispatcher = make(map[string]func() error)
	c.dispatcher["update"] = c.updateGit
	c.dispatcher["update-gitlab"] = c.updateGitlab

	return &c, nil
}

func validate(opt *Options) error {

	if _, err := os.Stat(opt.Dir); err != nil {
		return err
	}

	if opt.Auth.Username == "" || opt.Auth.Password == "" {
		return ErrCredentialNotFound
	}

	if opt.Action == "" {
		return ErrActionNotFound
	}

	if opt.Logs == nil {
		return ErrLogsNotDefine
	}

	return nil
}

type Command struct {
	verbose    bool
	action     string
	dir        string
	auth       *Auth
	hardReset  bool
	baseurl    string
	dispatcher map[string]func() error
}

// Execute action based on action key provided
func (c *Command) Execute() error {

	if c.dispatcher[c.action] == nil {
		return ErrCommandNotFound
	}

	err := c.dispatcher[c.action]()
	if err != nil {
		return err
	}

	return nil
}
